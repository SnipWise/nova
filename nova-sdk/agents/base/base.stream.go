package base

import (
	"encoding/json"
	"errors"

	"github.com/openai/openai-go/v3"
)

// prepareMessagesToSend builds the message list for an API call.
// When KeepConversationHistory is true, the incoming messages are appended to the
// agent's history permanently and the full history is returned.
// Otherwise a one-shot list (history + new messages) is returned without mutating history.
func (agent *Agent) prepareMessagesToSend(
	messages []openai.ChatCompletionMessageParamUnion,
) []openai.ChatCompletionMessageParamUnion {
	if agent.Config.KeepConversationHistory {
		agent.ChatCompletionParams.Messages = append(agent.ChatCompletionParams.Messages, messages...)
		return agent.ChatCompletionParams.Messages
	}
	return append(agent.ChatCompletionParams.Messages, messages...)
}

// appendAssistantToHistory saves the assistant response to the conversation history
// when KeepConversationHistory is enabled and the response is non-empty.
func (agent *Agent) appendAssistantToHistory(response string) {
	if agent.Config.KeepConversationHistory && response != "" {
		agent.ChatCompletionParams.Messages = append(
			agent.ChatCompletionParams.Messages,
			openai.AssistantMessage(response),
		)
	}
}

// streamCloser is the minimal interface required by finalizeStream.
type streamCloser interface {
	Err() error
	Close() error
}

// finalizeStream validates and closes a stream after the read loop finishes.
// Returns the first error encountered (Err takes priority over Close).
func (agent *Agent) finalizeStream(stream streamCloser) error {
	if err := stream.Err(); err != nil {
		agent.Log.Error("Stream error: %v", err)
		return err
	}
	if err := stream.Close(); err != nil {
		agent.Log.Error("Stream close error: %v", err)
		return err
	}
	return nil
}

// processReasoningChunk extracts the reasoning_content field from the current chunk delta,
// appends it to reasoning, sets hasReceivedReasoning, and calls reasoningCallback.
// Returns nil when no reasoning is present in the chunk.
func processReasoningChunk(
	chunk openai.ChatCompletionChunk,
	finishReason string,
	reasoning *string,
	hasReceivedReasoning *bool,
	reasoningCallback func(string, string) error,
) error {
	if len(chunk.Choices) == 0 {
		return nil
	}
	var content struct {
		ReasoningContent string `json:"reasoning_content"`
	}
	if err := json.Unmarshal([]byte(chunk.Choices[0].Delta.RawJSON()), &content); err != nil {
		return nil
	}
	if content.ReasoningContent == "" {
		return nil
	}
	*hasReceivedReasoning = true
	*reasoning += content.ReasoningContent
	return reasoningCallback(content.ReasoningContent, finishReason)
}

// processResponseChunk forwards the Delta.Content of the current chunk to responseCallback.
// When this is the first content chunk after a reasoning phase, it signals the end of
// reasoning to reasoningCallback before forwarding the content.
// Returns nil when the chunk carries no content.
func processResponseChunk(
	chunk openai.ChatCompletionChunk,
	finishReason string,
	response *string,
	hasReceivedReasoning bool,
	reasoningEnded *bool,
	reasoningCallback func(string, string) error,
	responseCallback func(string, string) error,
) error {
	if len(chunk.Choices) == 0 || chunk.Choices[0].Delta.Content == "" {
		return nil
	}
	if hasReceivedReasoning && !*reasoningEnded {
		*reasoningEnded = true
		if err := reasoningCallback("", "end_of_reasoning"); err != nil {
			return err
		}
	}
	content := chunk.Choices[0].Delta.Content
	*response += content
	return responseCallback(content, finishReason)
}

// processStreamChunk handles one chunk from a non-reasoning stream.
// It captures the finishReason when present and forwards any content to callBack.
// Having no Choices is treated as a no-op; returns nil in that case.
func (agent *Agent) processStreamChunk(
	chunk openai.ChatCompletionChunk,
	finishReason *string,
	response *string,
	callBack func(string, string) error,
) error {
	if len(chunk.Choices) == 0 {
		return nil
	}
	if chunk.Choices[0].FinishReason != "" {
		agent.SaveLastChunkResponse(&chunk)
		*finishReason = chunk.Choices[0].FinishReason
	}
	content := chunk.Choices[0].Delta.Content
	if content == "" {
		return nil
	}
	*response += content
	return callBack(content, *finishReason)
}

// canceledError is returned when the stream is stopped via StopStream.
var canceledError = errors.New("stream canceled by user")
