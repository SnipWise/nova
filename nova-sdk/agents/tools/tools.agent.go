package tools

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/openai/openai-go/v3"
	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/mcptools"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
	"github.com/snipwise/nova/nova-sdk/models"
	"github.com/snipwise/nova/nova-sdk/toolbox/conversion"
	"github.com/snipwise/nova/nova-sdk/toolbox/logger"
)

// ToolCallResult represents the result of tool call detection
type ToolCallResult struct {
	FinishReason         string
	Results              []string
	LastAssistantMessage string
}

// ToolCallback is a function called when a tool needs to be executed
type ToolCallback func(functionName string, arguments string) (string, error)
type ConfirmationCallback func(functionName string, arguments string) ConfirmationResponse

// StreamCallback is a function called for each chunk of streaming response
type StreamCallback func(chunk string) error

// Agent represents a simplified tools agent that hides OpenAI SDK details
type Agent struct {
	config         agents.Config
	modelConfig    models.Config
	internalAgent  *BaseAgent
	log            logger.Logger
	toolsFunctions map[string]func(args ...any) (any, error)

	// Tool execution callbacks (can be set via options)
	executeFunction             ToolCallback
	confirmationPromptFunction  ConfirmationCallback

	// Lifecycle hooks
	beforeCompletion func(*Agent)
	afterCompletion  func(*Agent)
}

// ToolAgentOption is a functional option for configuring an Agent during creation
type ToolAgentOption func(*openai.ChatCompletionNewParams)

// ToolsAgentOption is a functional option for configuring lifecycle hooks on the Agent
type ToolsAgentOption func(*Agent)

// BeforeCompletion sets a hook that is called before each tool call detection
func BeforeCompletion(fn func(*Agent)) ToolsAgentOption {
	return func(a *Agent) {
		a.beforeCompletion = fn
	}
}

// AfterCompletion sets a hook that is called after each tool call detection
func AfterCompletion(fn func(*Agent)) ToolsAgentOption {
	return func(a *Agent) {
		a.afterCompletion = fn
	}
}

// WithExecuteFn sets the default tool execution callback for the agent
// This callback will be used by all detection methods if no callback is explicitly provided
func WithExecuteFn(fn ToolCallback) ToolsAgentOption {
	return func(a *Agent) {
		a.executeFunction = fn
	}
}

// WithConfirmationPromptFn sets the default confirmation callback for the agent
// This callback will be used by all confirmation methods if no callback is explicitly provided
func WithConfirmationPromptFn(fn ConfirmationCallback) ToolsAgentOption {
	return func(a *Agent) {
		a.confirmationPromptFunction = fn
	}
}

// WithTools sets custom tools for the agent
func WithOpenAITools(tools []openai.ChatCompletionToolUnionParam) ToolAgentOption {
	return func(params *openai.ChatCompletionNewParams) {
		params.Tools = tools
	}
}

func WithTools(tools []*Tool) ToolAgentOption {
	return func(params *openai.ChatCompletionNewParams) {
		params.Tools = ToOpenAITools(tools)
	}
}

func WithMCPTools(tools []mcp.Tool) ToolAgentOption {
	return func(params *openai.ChatCompletionNewParams) {
		params.Tools = mcptools.ConvertMCPToolsToOpenAITools(tools)
	}
}

// TODO: WithMCPToolsWithFilter

// NewAgent creates a new simplified tools agent
func NewAgent(
	ctx context.Context,
	agentConfig agents.Config,
	modelConfig models.Config,
	options ...any,
) (*Agent, error) {
	log := logger.GetLoggerFromEnv()

	// Separate ToolAgentOption (for OpenAI params) from ToolsAgentOption (for Agent hooks)
	var toolOptions []ToolAgentOption
	var agentOptions []ToolsAgentOption
	for _, opt := range options {
		switch o := opt.(type) {
		case ToolAgentOption:
			toolOptions = append(toolOptions, o)
		case ToolsAgentOption:
			agentOptions = append(agentOptions, o)
		}
	}

	// Create internal OpenAI-based agent with converted parameters
	openaiModelConfig := models.ConvertToOpenAIModelConfig(modelConfig)

	// Apply ToolAgentOption configurations (tools, etc.)
	for _, opt := range toolOptions {
		opt(&openaiModelConfig)
	}

	internalAgent, err := NewBaseAgent(ctx, agentConfig, openaiModelConfig)
	if err != nil {
		return nil, err
	}

	agent := &Agent{
		config:         agentConfig,
		modelConfig:    modelConfig,
		internalAgent:  internalAgent,
		log:            log,
		toolsFunctions: make(map[string]func(args ...any) (any, error)),
	}

	// Apply ToolsAgentOption configurations (hooks)
	for _, opt := range agentOptions {
		opt(agent)
	}

	return agent, nil
}

// Kind returns the agent type
func (agent *Agent) Kind() agents.Kind {
	return agents.Tools
}

func (agent *Agent) GetName() string {
	return agent.config.Name
}

func (agent *Agent) GetModelID() string {
	return agent.modelConfig.Name
}

// GetMessages returns all conversation messages
func (agent *Agent) GetMessages() []messages.Message {
	openaiMessages := agent.internalAgent.GetMessages()
	agentMessages := messages.ConvertFromOpenAIMessages(openaiMessages)
	return agentMessages
}

func (agent *Agent) ExportMessagesToJSON() (string, error) {
	messagesList := agent.GetMessages()
	jsonData, err := json.MarshalIndent(messagesList, "", "  ")
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}

// GetContextSize returns the approximate size of the current context
func (agent *Agent) GetContextSize() int {
	return agent.internalAgent.GetCurrentContextSize()
}

// ResetMessages clears all messages except the system instruction
func (agent *Agent) ResetMessages() {
	agent.internalAgent.ResetMessages()
}

// AddMessage adds a message to the conversation history
func (agent *Agent) AddMessage(role roles.Role, content string) {
	agent.internalAgent.AddMessage(
		messages.ConvertToOpenAIMessage(messages.Message{
			Role:    role,
			Content: content,
		}),
	)
}

// AddMessages adds multiple messages to the conversation history
func (agent *Agent) AddMessages(msgs []messages.Message) {
	openaiMessages := messages.ConvertToOpenAIMessages(msgs)
	agent.internalAgent.AddMessages(openaiMessages)
}

// NOTE: IMPORTANT: Not all LLMs with tool support support parallel tool calls.
// DetectParallelToolCalls detects and executes multiple tool calls in parallel
// The toolCallback parameter is optional. If not provided, uses the callback set via WithExecuteFn option.
func (agent *Agent) DetectParallelToolCalls(
	userMessages []messages.Message,
	toolCallback ...ToolCallback,
) (*ToolCallResult, error) {
	if len(userMessages) == 0 {
		return nil, errors.New("no messages provided")
	}

	// Determine which callback to use: parameter takes priority over option
	var callback ToolCallback
	if len(toolCallback) > 0 && toolCallback[0] != nil {
		callback = toolCallback[0]
	} else if agent.executeFunction != nil {
		callback = agent.executeFunction
	} else {
		return nil, errors.New("no tool callback provided: either pass toolCallback parameter or set it via WithExecuteFn option")
	}

	// Call before completion hook if set
	if agent.beforeCompletion != nil {
		agent.beforeCompletion(agent)
	}

	// Convert to OpenAI format
	openaiMessages := messages.ConvertToOpenAIMessages(userMessages)

	// Call internal agent
	finishReason, results, lastAssistantMessage, err := agent.internalAgent.DetectParallelToolCalls(openaiMessages, callback)
	if err != nil {
		return nil, err
	}

	// Add assistant response to history only if KeepConversationHistory is true
	if agent.config.KeepConversationHistory {
		// Add assistant response to history if present
		if lastAssistantMessage != "" {
			agent.internalAgent.AddMessage(
				openai.AssistantMessage(lastAssistantMessage),
			)
		}
	}

	result := &ToolCallResult{
		FinishReason:         finishReason,
		Results:              results,
		LastAssistantMessage: lastAssistantMessage,
	}

	// Call after completion hook if set
	if agent.afterCompletion != nil {
		agent.afterCompletion(agent)
	}

	return result, nil
}

// NOTE: IMPORTANT: Not all LLMs with tool support support parallel tool calls.
// DetectParallelToolCallsWithConfirmation detects and executes multiple tool calls with user confirmation
// Optional callbacks can be provided in order: toolCallback, confirmationCallback
// Usage:
//   - DetectParallelToolCallsWithConfirmation(messages) → uses both from options
//   - DetectParallelToolCallsWithConfirmation(messages, toolCallback) → toolCallback from param, confirmationCallback from option
//   - DetectParallelToolCallsWithConfirmation(messages, toolCallback, confirmationCallback) → both from params
func (agent *Agent) DetectParallelToolCallsWithConfirmation(
	userMessages []messages.Message,
	callbacks ...any,
) (*ToolCallResult, error) {
	if len(userMessages) == 0 {
		return nil, errors.New("no messages provided")
	}

	// Extract callbacks by position (order matters!)
	var callback ToolCallback
	var confirmation ConfirmationCallback

	// First parameter (if provided) is toolCallback
	if len(callbacks) >= 1 && callbacks[0] != nil {
		// Try type alias first
		if tc, ok := callbacks[0].(ToolCallback); ok {
			callback = tc
		} else {
			// Try underlying function type (type aliases may not work in type assertions)
			v := reflect.ValueOf(callbacks[0])
			if v.Kind() == reflect.Func {
				// Check signature matches: func(string, string) (string, error)
				t := v.Type()
				if t.NumIn() == 2 && t.NumOut() == 2 &&
					t.In(0).Kind() == reflect.String && t.In(1).Kind() == reflect.String &&
					t.Out(0).Kind() == reflect.String {
					// Convert to ToolCallback via direct assignment
					if fn, ok := callbacks[0].(func(string, string) (string, error)); ok {
						callback = fn
					}
				}
			}
		}
	}
	// Use option if not provided as parameter
	if callback == nil {
		if agent.executeFunction != nil {
			callback = agent.executeFunction
		} else {
			return nil, errors.New("no tool callback provided: either pass ToolCallback parameter or set it via WithExecuteFn option")
		}
	}

	// Second parameter (if provided) is confirmationCallback
	if len(callbacks) >= 2 && callbacks[1] != nil {
		// Try type alias first
		if cc, ok := callbacks[1].(ConfirmationCallback); ok {
			confirmation = cc
		} else {
			// Try underlying function type
			v := reflect.ValueOf(callbacks[1])
			if v.Kind() == reflect.Func {
				// Check signature matches: func(string, string) ConfirmationResponse
				t := v.Type()
				if t.NumIn() == 2 && t.NumOut() == 1 &&
					t.In(0).Kind() == reflect.String && t.In(1).Kind() == reflect.String {
					// Convert to ConfirmationCallback via direct assignment
					if fn, ok := callbacks[1].(func(string, string) ConfirmationResponse); ok {
						confirmation = fn
					}
				}
			}
		}
	}
	// Use option if not provided as parameter
	if confirmation == nil {
		if agent.confirmationPromptFunction != nil {
			confirmation = agent.confirmationPromptFunction
		} else {
			return nil, errors.New("no confirmation callback provided: either pass ConfirmationCallback parameter or set it via WithConfirmationPromptFn option")
		}
	}

	// Call before completion hook if set
	if agent.beforeCompletion != nil {
		agent.beforeCompletion(agent)
	}

	// Convert to OpenAI format
	openaiMessages := messages.ConvertToOpenAIMessages(userMessages)

	// Call internal agent
	finishReason, results, lastAssistantMessage, err := agent.internalAgent.DetectParallelToolCallsWitConfirmation(
		openaiMessages,
		callback,
		confirmation,
	)
	if err != nil {
		return nil, err
	}

	// Add assistant response to history only if KeepConversationHistory is true
	if agent.config.KeepConversationHistory {
		// Add assistant response to history if present
		if lastAssistantMessage != "" {
			agent.internalAgent.AddMessage(
				openai.AssistantMessage(lastAssistantMessage),
			)
		}
	}

	result := &ToolCallResult{
		FinishReason:         finishReason,
		Results:              results,
		LastAssistantMessage: lastAssistantMessage,
	}

	// Call after completion hook if set
	if agent.afterCompletion != nil {
		agent.afterCompletion(agent)
	}

	return result, nil
}

// DetectToolCallsLoop sends messages and detects tool calls, executing them via callback
// The toolCallback parameter is optional. If not provided, uses the callback set via WithExecuteFn option.
// If neither is provided, returns an error.
func (agent *Agent) DetectToolCallsLoop(
	userMessages []messages.Message,
	toolCallback ...ToolCallback,
) (*ToolCallResult, error) {
	if len(userMessages) == 0 {
		return nil, errors.New("no messages provided")
	}

	// Determine which callback to use: parameter takes priority over option
	var callback ToolCallback
	if len(toolCallback) > 0 && toolCallback[0] != nil {
		callback = toolCallback[0]
	} else if agent.executeFunction != nil {
		callback = agent.executeFunction
	} else {
		return nil, errors.New("no tool callback provided: either pass toolCallback parameter or set it via WithExecuteFn option")
	}

	// Call before completion hook if set
	if agent.beforeCompletion != nil {
		agent.beforeCompletion(agent)
	}

	// Convert to OpenAI format
	openaiMessages := messages.ConvertToOpenAIMessages(userMessages)

	// Call internal agent
	finishReason, results, lastAssistantMessage, err := agent.internalAgent.DetectToolCallsLoop(openaiMessages, callback)

	if err != nil {
		return nil, err
	}

	// Add assistant response to history only if KeepConversationHistory is true
	if agent.config.KeepConversationHistory {
		// Add assistant response to history if present
		if lastAssistantMessage != "" {
			agent.internalAgent.AddMessage(
				openai.AssistantMessage(lastAssistantMessage),
			)
		}
	}

	result := &ToolCallResult{
		FinishReason:         finishReason,
		Results:              results,
		LastAssistantMessage: lastAssistantMessage,
	}

	// Call after completion hook if set
	if agent.afterCompletion != nil {
		agent.afterCompletion(agent)
	}

	return result, nil
}

// DetectToolCallsLoopWithConfirmation sends messages and detects tool calls with user confirmation
// Optional callbacks can be provided in order: toolCallback, confirmationCallback
// If not provided, uses callbacks set via WithExecuteFn and WithConfirmationPromptFn options.
// Usage:
//   - DetectToolCallsLoopWithConfirmation(messages) → uses both from options
//   - DetectToolCallsLoopWithConfirmation(messages, toolCallback) → toolCallback from param, confirmationCallback from option
//   - DetectToolCallsLoopWithConfirmation(messages, toolCallback, confirmationCallback) → both from params
func (agent *Agent) DetectToolCallsLoopWithConfirmation(
	userMessages []messages.Message,
	callbacks ...any,
) (*ToolCallResult, error) {
	if len(userMessages) == 0 {
		return nil, errors.New("no messages provided")
	}

	// Extract callbacks by position (order matters!)
	var callback ToolCallback
	var confirmation ConfirmationCallback

	// First parameter (if provided) is toolCallback
	if len(callbacks) >= 1 && callbacks[0] != nil {
		// Try type alias first
		if tc, ok := callbacks[0].(ToolCallback); ok {
			callback = tc
		} else {
			// Try underlying function type (type aliases may not work in type assertions)
			v := reflect.ValueOf(callbacks[0])
			if v.Kind() == reflect.Func {
				// Check signature matches: func(string, string) (string, error)
				t := v.Type()
				if t.NumIn() == 2 && t.NumOut() == 2 &&
					t.In(0).Kind() == reflect.String && t.In(1).Kind() == reflect.String &&
					t.Out(0).Kind() == reflect.String {
					// Convert to ToolCallback via direct assignment
					if fn, ok := callbacks[0].(func(string, string) (string, error)); ok {
						callback = fn
					}
				}
			}
		}
	}
	// Use option if not provided as parameter
	if callback == nil {
		if agent.executeFunction != nil {
			callback = agent.executeFunction
		} else {
			return nil, errors.New("no tool callback provided: either pass ToolCallback parameter or set it via WithExecuteFn option")
		}
	}

	// Second parameter (if provided) is confirmationCallback
	if len(callbacks) >= 2 && callbacks[1] != nil {
		// Try type alias first
		if cc, ok := callbacks[1].(ConfirmationCallback); ok {
			confirmation = cc
		} else {
			// Try underlying function type
			v := reflect.ValueOf(callbacks[1])
			if v.Kind() == reflect.Func {
				// Check signature matches: func(string, string) ConfirmationResponse
				t := v.Type()
				if t.NumIn() == 2 && t.NumOut() == 1 &&
					t.In(0).Kind() == reflect.String && t.In(1).Kind() == reflect.String {
					// Convert to ConfirmationCallback via direct assignment
					if fn, ok := callbacks[1].(func(string, string) ConfirmationResponse); ok {
						confirmation = fn
					}
				}
			}
		}
	}
	// Use option if not provided as parameter
	if confirmation == nil {
		if agent.confirmationPromptFunction != nil {
			confirmation = agent.confirmationPromptFunction
		} else {
			return nil, errors.New("no confirmation callback provided: either pass ConfirmationCallback parameter or set it via WithConfirmationPromptFn option")
		}
	}

	// Call before completion hook if set
	if agent.beforeCompletion != nil {
		agent.beforeCompletion(agent)
	}

	// Convert to OpenAI format
	openaiMessages := messages.ConvertToOpenAIMessages(userMessages)

	// Call internal agent
	finishReason, results, lastAssistantMessage, err := agent.internalAgent.DetectToolCallsLoopWithConfirmation(
		openaiMessages,
		callback,
		confirmation,
	)

	if err != nil {
		return nil, err
	}

	// Add assistant response to history only if KeepConversationHistory is true
	if agent.config.KeepConversationHistory {
		// Add assistant response to history if present
		if lastAssistantMessage != "" {
			agent.internalAgent.AddMessage(
				openai.AssistantMessage(lastAssistantMessage),
			)
		}
	}

	result := &ToolCallResult{
		FinishReason:         finishReason,
		Results:              results,
		LastAssistantMessage: lastAssistantMessage,
	}

	// Call after completion hook if set
	if agent.afterCompletion != nil {
		agent.afterCompletion(agent)
	}

	return result, nil
}

// DetectToolCallsLoopStream sends messages and detects tool calls with streaming
// The toolCallback parameter is optional. If not provided, uses the callback set via WithExecuteFn option.
// The streamCallback parameter is required (it's the purpose of this method).
func (agent *Agent) DetectToolCallsLoopStream(
	userMessages []messages.Message,
	streamCallback StreamCallback,
	toolCallback ...ToolCallback,
) (*ToolCallResult, error) {
	if len(userMessages) == 0 {
		return nil, errors.New("no messages provided")
	}

	if streamCallback == nil {
		return nil, errors.New("streamCallback is required for DetectToolCallsLoopStream")
	}

	// Determine which callback to use: parameter takes priority over option
	var callback ToolCallback
	if len(toolCallback) > 0 && toolCallback[0] != nil {
		callback = toolCallback[0]
	} else if agent.executeFunction != nil {
		callback = agent.executeFunction
	} else {
		return nil, errors.New("no tool callback provided: either pass toolCallback parameter or set it via WithExecuteFn option")
	}

	// Call before completion hook if set
	if agent.beforeCompletion != nil {
		agent.beforeCompletion(agent)
	}

	// Convert to OpenAI format
	openaiMessages := messages.ConvertToOpenAIMessages(userMessages)

	// Call internal agent with streaming
	finishReason, results, lastAssistantMessage, err := agent.internalAgent.DetectToolCallsLoopStream(
		openaiMessages,
		callback,
		streamCallback,
	)
	if err != nil {
		return nil, err
	}

	// Add assistant response to history only if KeepConversationHistory is true
	if agent.config.KeepConversationHistory {
		// Add assistant response to history if present
		if lastAssistantMessage != "" {
			agent.internalAgent.AddMessage(
				openai.AssistantMessage(lastAssistantMessage),
			)
		}
	}

	result := &ToolCallResult{
		FinishReason:         finishReason,
		Results:              results,
		LastAssistantMessage: lastAssistantMessage,
	}

	// Call after completion hook if set
	if agent.afterCompletion != nil {
		agent.afterCompletion(agent)
	}

	return result, nil
}

// DetectToolCallsLoopWithConfirmationStream sends messages and detects tool calls with confirmation and streaming
// The streamCallback parameter is required. Optional callbacks can be provided in order: toolCallback, confirmationCallback
// Usage:
//   - DetectToolCallsLoopWithConfirmationStream(messages, streamCallback) → uses both callbacks from options
//   - DetectToolCallsLoopWithConfirmationStream(messages, streamCallback, toolCallback) → toolCallback from param, confirmationCallback from option
//   - DetectToolCallsLoopWithConfirmationStream(messages, streamCallback, toolCallback, confirmationCallback) → both from params
func (agent *Agent) DetectToolCallsLoopWithConfirmationStream(
	userMessages []messages.Message,
	streamCallback StreamCallback,
	callbacks ...any,
) (*ToolCallResult, error) {
	if len(userMessages) == 0 {
		return nil, errors.New("no messages provided")
	}

	if streamCallback == nil {
		return nil, errors.New("streamCallback is required for DetectToolCallsLoopWithConfirmationStream")
	}

	// Extract callbacks by position (order matters!)
	var callback ToolCallback
	var confirmation ConfirmationCallback

	// First parameter (if provided) is toolCallback
	if len(callbacks) >= 1 && callbacks[0] != nil {
		// Try type alias first
		if tc, ok := callbacks[0].(ToolCallback); ok {
			callback = tc
		} else {
			// Try underlying function type (type aliases may not work in type assertions)
			v := reflect.ValueOf(callbacks[0])
			if v.Kind() == reflect.Func {
				// Check signature matches: func(string, string) (string, error)
				t := v.Type()
				if t.NumIn() == 2 && t.NumOut() == 2 &&
					t.In(0).Kind() == reflect.String && t.In(1).Kind() == reflect.String &&
					t.Out(0).Kind() == reflect.String {
					// Convert to ToolCallback via direct assignment
					if fn, ok := callbacks[0].(func(string, string) (string, error)); ok {
						callback = fn
					}
				}
			}
		}
	}
	// Use option if not provided as parameter
	if callback == nil {
		if agent.executeFunction != nil {
			callback = agent.executeFunction
		} else {
			return nil, errors.New("no tool callback provided: either pass ToolCallback parameter or set it via WithExecuteFn option")
		}
	}

	// Second parameter (if provided) is confirmationCallback
	if len(callbacks) >= 2 && callbacks[1] != nil {
		// Try type alias first
		if cc, ok := callbacks[1].(ConfirmationCallback); ok {
			confirmation = cc
		} else {
			// Try underlying function type
			v := reflect.ValueOf(callbacks[1])
			if v.Kind() == reflect.Func {
				// Check signature matches: func(string, string) ConfirmationResponse
				t := v.Type()
				if t.NumIn() == 2 && t.NumOut() == 1 &&
					t.In(0).Kind() == reflect.String && t.In(1).Kind() == reflect.String {
					// Convert to ConfirmationCallback via direct assignment
					if fn, ok := callbacks[1].(func(string, string) ConfirmationResponse); ok {
						confirmation = fn
					}
				}
			}
		}
	}
	// Use option if not provided as parameter
	if confirmation == nil {
		if agent.confirmationPromptFunction != nil {
			confirmation = agent.confirmationPromptFunction
		} else {
			return nil, errors.New("no confirmation callback provided: either pass ConfirmationCallback parameter or set it via WithConfirmationPromptFn option")
		}
	}

	// Call before completion hook if set
	if agent.beforeCompletion != nil {
		agent.beforeCompletion(agent)
	}

	// Convert to OpenAI format
	openaiMessages := messages.ConvertToOpenAIMessages(userMessages)

	// Call internal agent with streaming
	finishReason, results, lastAssistantMessage, err := agent.internalAgent.DetectToolCallsLoopWithConfirmationStream(
		openaiMessages,
		callback,
		confirmation,
		streamCallback,
	)
	if err != nil {
		return nil, err
	}

	// Add assistant response to history only if KeepConversationHistory is true
	if agent.config.KeepConversationHistory {
		// Add assistant response to history if present
		if lastAssistantMessage != "" {
			agent.internalAgent.AddMessage(
				openai.AssistantMessage(lastAssistantMessage),
			)
		}
	}

	result := &ToolCallResult{
		FinishReason:         finishReason,
		Results:              results,
		LastAssistantMessage: lastAssistantMessage,
	}

	// Call after completion hook if set
	if agent.afterCompletion != nil {
		agent.afterCompletion(agent)
	}

	return result, nil
}

// === Config Getters and Setters ===

// GetConfig returns the agent configuration
func (agent *Agent) GetConfig() agents.Config {
	return agent.config
}

// SetConfig updates the agent configuration
func (agent *Agent) SetConfig(config agents.Config) {
	agent.config = config
	agent.internalAgent.Config = config
}

// GetModelConfig returns the model configuration
func (agent *Agent) GetModelConfig() models.Config {
	return agent.modelConfig
}

// SetModelConfig updates the model configuration
// Note: This updates the stored config but doesn't regenerate the internal OpenAI params
// For most parameters to take effect, create a new agent with the new config
func (agent *Agent) SetModelConfig(config models.Config) {
	agent.modelConfig = config
	// Update the internal OpenAI params with the new config
	openaiModelConfig := models.ConvertToOpenAIModelConfig(config)
	// Preserve the existing tools
	openaiModelConfig.Tools = agent.internalAgent.ChatCompletionParams.Tools
	agent.internalAgent.ChatCompletionParams = openaiModelConfig
}

func (agent *Agent) GetLastRequestRawJSON() string {
	return agent.internalAgent.GetLastRequestRawJSON()
}

func (agent *Agent) GetLastResponseRawJSON() string {
	return agent.internalAgent.GetLastResponseRawJSON()
}

func (agent *Agent) GetLastRequestSON() (string, error) {
	return conversion.PrettyPrint(agent.internalAgent.GetLastRequestRawJSON())
}

func (agent *Agent) GetLastResponseJSON() (string, error) {
	return conversion.PrettyPrint(agent.internalAgent.GetLastResponseRawJSON())
}

// GetLastStatedToolCalls returns the last stated tool calls state
//
//   - IMPORTANT: Allows checking the state of tool calls across multiple invocations
//   - USEFUL: for maintaining continuity in tool call confirmations and executions
//
// and when transfering context between different agent instances
func (agent *Agent) GetLastStateToolCalls() LastToolCallsState {
	return agent.internalAgent.lastState
}
func (agent *Agent) ResetLastStateToolCalls() {
	agent.internalAgent.lastState = LastToolCallsState{}
}

// GetContext returns the agent's context
func (agent *Agent) GetContext() context.Context {
	return agent.internalAgent.GetContext()
}

// SetContext updates the agent's context
func (agent *Agent) SetContext(ctx context.Context) {
	agent.internalAgent.SetContext(ctx)
}

// GetTools returns the tools configured for this agent
func (agent *Agent) GetTools() []openai.ChatCompletionToolUnionParam {
	return agent.internalAgent.Agent.ChatCompletionParams.Tools
}
