package base

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/openai/openai-go/v3"
	"github.com/snipwise/nova/nova-sdk/agents"
	"github.com/snipwise/nova/nova-sdk/toolbox/logger"
)

// ── helpers ───────────────────────────────────────────────────────────────────

func newTestAgent(keepHistory bool) *Agent {
	return &Agent{
		Config: agents.Config{KeepConversationHistory: keepHistory},
		Log:    logger.GetLoggerFromEnv(),
		ChatCompletionParams: openai.ChatCompletionNewParams{
			Messages: []openai.ChatCompletionMessageParamUnion{},
		},
	}
}

func userMsg(content string) openai.ChatCompletionMessageParamUnion {
	return openai.UserMessage(content)
}

// stubStream implements streamCloser for tests.
type stubStream struct {
	errResult   error
	closeResult error
}

func (s *stubStream) Err() error   { return s.errResult }
func (s *stubStream) Close() error { return s.closeResult }

// chunkWithContent builds a ChatCompletionChunk containing a text delta.
func chunkWithContent(content string) openai.ChatCompletionChunk {
	return openai.ChatCompletionChunk{
		Choices: []openai.ChatCompletionChunkChoice{
			{Delta: openai.ChatCompletionChunkChoiceDelta{Content: content}},
		},
	}
}

// chunkWithReasoning builds a ChatCompletionChunk where the delta carries
// a reasoning_content field in its raw JSON.
func chunkWithReasoning(reasoningContent string) openai.ChatCompletionChunk {
	raw := `{"reasoning_content":"` + reasoningContent + `"}`
	var delta openai.ChatCompletionChunkChoiceDelta
	_ = json.Unmarshal([]byte(raw), &delta)
	return openai.ChatCompletionChunk{
		Choices: []openai.ChatCompletionChunkChoice{{Delta: delta}},
	}
}

// ── prepareMessagesToSend ─────────────────────────────────────────────────────

func TestPrepareMessagesToSend_KeepHistory_AppendsAndReturnsHistory(t *testing.T) {
	a := newTestAgent(true)
	a.ChatCompletionParams.Messages = []openai.ChatCompletionMessageParamUnion{userMsg("existing")}

	result := a.prepareMessagesToSend([]openai.ChatCompletionMessageParamUnion{userMsg("new")})

	if len(result) != 2 {
		t.Fatalf("want 2 messages, got %d", len(result))
	}
	if len(a.ChatCompletionParams.Messages) != 2 {
		t.Error("history should be mutated when KeepConversationHistory is true")
	}
}

func TestPrepareMessagesToSend_NoHistory_ReturnsEphemeralList(t *testing.T) {
	a := newTestAgent(false)
	a.ChatCompletionParams.Messages = []openai.ChatCompletionMessageParamUnion{userMsg("system")}

	result := a.prepareMessagesToSend([]openai.ChatCompletionMessageParamUnion{userMsg("user")})

	if len(result) != 2 {
		t.Fatalf("want 2 messages, got %d", len(result))
	}
	if len(a.ChatCompletionParams.Messages) != 1 {
		t.Error("history must NOT be mutated when KeepConversationHistory is false")
	}
}

func TestPrepareMessagesToSend_NoHistory_EmptyIncoming_ReturnsHistoryOnly(t *testing.T) {
	a := newTestAgent(false)
	a.ChatCompletionParams.Messages = []openai.ChatCompletionMessageParamUnion{userMsg("base")}

	result := a.prepareMessagesToSend(nil)

	if len(result) != 1 {
		t.Fatalf("want 1 message, got %d", len(result))
	}
}

// ── appendAssistantToHistory ──────────────────────────────────────────────────

func TestAppendAssistantToHistory_KeepHistory_NonEmpty_Appends(t *testing.T) {
	a := newTestAgent(true)
	a.appendAssistantToHistory("hello")
	if len(a.ChatCompletionParams.Messages) != 1 {
		t.Fatalf("want 1 message appended, got %d", len(a.ChatCompletionParams.Messages))
	}
}

func TestAppendAssistantToHistory_KeepHistory_EmptyResponse_DoesNotAppend(t *testing.T) {
	a := newTestAgent(true)
	a.appendAssistantToHistory("")
	if len(a.ChatCompletionParams.Messages) != 0 {
		t.Error("empty response should not be appended")
	}
}

func TestAppendAssistantToHistory_NoHistory_DoesNotAppend(t *testing.T) {
	a := newTestAgent(false)
	a.appendAssistantToHistory("hello")
	if len(a.ChatCompletionParams.Messages) != 0 {
		t.Error("history should not be appended when KeepConversationHistory is false")
	}
}

// ── finalizeStream ────────────────────────────────────────────────────────────

func TestFinalizeStream_NoError_ReturnsNil(t *testing.T) {
	a := newTestAgent(false)
	if err := a.finalizeStream(&stubStream{}); err != nil {
		t.Errorf("want nil, got %v", err)
	}
}

func TestFinalizeStream_ErrError_ReturnsIt(t *testing.T) {
	a := newTestAgent(false)
	want := errors.New("stream failed")
	err := a.finalizeStream(&stubStream{errResult: want})
	if err != want {
		t.Errorf("want %v, got %v", want, err)
	}
}

func TestFinalizeStream_CloseError_ReturnsIt(t *testing.T) {
	a := newTestAgent(false)
	want := errors.New("close failed")
	err := a.finalizeStream(&stubStream{closeResult: want})
	if err != want {
		t.Errorf("want %v, got %v", want, err)
	}
}

func TestFinalizeStream_ErrTakesPriorityOverClose(t *testing.T) {
	a := newTestAgent(false)
	errErr := errors.New("stream err")
	closeErr := errors.New("close err")
	err := a.finalizeStream(&stubStream{errResult: errErr, closeResult: closeErr})
	if err != errErr {
		t.Errorf("Err() should take priority, got %v", err)
	}
}

// ── processReasoningChunk ─────────────────────────────────────────────────────

func TestProcessReasoningChunk_NoChoices_ReturnsNil(t *testing.T) {
	var received bool
	cb := func(_, _ string) error { received = true; return nil }

	err := processReasoningChunk(openai.ChatCompletionChunk{}, "", new(string), new(bool), cb)

	if err != nil || received {
		t.Error("no choices: want nil error and no callback")
	}
}

func TestProcessReasoningChunk_NoReasoningContent_ReturnsNil(t *testing.T) {
	var received bool
	cb := func(_, _ string) error { received = true; return nil }

	err := processReasoningChunk(chunkWithContent("hello"), "", new(string), new(bool), cb)

	if err != nil || received {
		t.Error("no reasoning_content: want nil error and no callback")
	}
}

func TestProcessReasoningChunk_ValidReasoning_CallsCallbackAndAccumulates(t *testing.T) {
	var got string
	cb := func(partial, _ string) error { got = partial; return nil }
	reasoning := ""
	hasReceived := false

	err := processReasoningChunk(chunkWithReasoning("thinking..."), "fr", &reasoning, &hasReceived, cb)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasReceived {
		t.Error("hasReceivedReasoning should be true")
	}
	if reasoning != "thinking..." {
		t.Errorf("reasoning: want %q, got %q", "thinking...", reasoning)
	}
	if got != "thinking..." {
		t.Errorf("callback arg: want %q, got %q", "thinking...", got)
	}
}

func TestProcessReasoningChunk_CallbackError_Propagates(t *testing.T) {
	want := errors.New("cb error")
	cb := func(_, _ string) error { return want }

	err := processReasoningChunk(chunkWithReasoning("x"), "", new(string), new(bool), cb)

	if err != want {
		t.Errorf("want %v, got %v", want, err)
	}
}

// ── processResponseChunk ─────────────────────────────────────────────────────

func TestProcessResponseChunk_NoChoices_ReturnsNil(t *testing.T) {
	var received bool
	cb := func(_, _ string) error { received = true; return nil }

	err := processResponseChunk(openai.ChatCompletionChunk{}, "", new(string), false, new(bool), cb, cb)

	if err != nil || received {
		t.Error("no choices: want nil error and no callback")
	}
}

func TestProcessResponseChunk_EmptyContent_ReturnsNil(t *testing.T) {
	var received bool
	cb := func(_, _ string) error { received = true; return nil }

	err := processResponseChunk(chunkWithContent(""), "", new(string), false, new(bool), cb, cb)

	if err != nil || received {
		t.Error("empty content: want nil error and no callback")
	}
}

func TestProcessResponseChunk_ContentWithoutReasoning_CallsResponseCallback(t *testing.T) {
	var got string
	responseCb := func(content, _ string) error { got = content; return nil }
	reasoningCb := func(_, _ string) error { t.Error("reasoning callback must not be called"); return nil }

	response := ""
	err := processResponseChunk(chunkWithContent("hi"), "", &response, false, new(bool), reasoningCb, responseCb)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if response != "hi" || got != "hi" {
		t.Errorf("want %q, got response=%q cb=%q", "hi", response, got)
	}
}

func TestProcessResponseChunk_FirstContentAfterReasoning_SignalsEndThenForwards(t *testing.T) {
	var reasoningSignal string
	reasoningCb := func(content, signal string) error { reasoningSignal = signal; return nil }

	var responseCbContent string
	responseCb := func(content, _ string) error { responseCbContent = content; return nil }

	response := ""
	reasoningEnded := false

	err := processResponseChunk(chunkWithContent("answer"), "", &response, true, &reasoningEnded, reasoningCb, responseCb)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reasoningEnded {
		t.Error("reasoningEnded should be true")
	}
	if reasoningSignal != "end_of_reasoning" {
		t.Errorf("want end_of_reasoning signal, got %q", reasoningSignal)
	}
	if responseCbContent != "answer" {
		t.Errorf("want response %q, got %q", "answer", responseCbContent)
	}
}

func TestProcessResponseChunk_ResponseCallbackError_Propagates(t *testing.T) {
	want := errors.New("resp error")
	responseCb := func(_, _ string) error { return want }
	reasoningCb := func(_, _ string) error { return nil }

	err := processResponseChunk(chunkWithContent("x"), "", new(string), false, new(bool), reasoningCb, responseCb)

	if err != want {
		t.Errorf("want %v, got %v", want, err)
	}
}

// ── processStreamChunk ────────────────────────────────────────────────────────

func TestProcessStreamChunk_NoChoices_ReturnsNil(t *testing.T) {
	a := newTestAgent(false)
	var received bool
	cb := func(_, _ string) error { received = true; return nil }

	finishReason := ""
	response := ""
	err := a.processStreamChunk(openai.ChatCompletionChunk{}, &finishReason, &response, cb)

	if err != nil || received {
		t.Error("no choices: want nil error and no callback")
	}
}

func TestProcessStreamChunk_EmptyContent_CapturesFinishReason(t *testing.T) {
	a := newTestAgent(false)
	var received bool
	cb := func(_, _ string) error { received = true; return nil }

	chunk := openai.ChatCompletionChunk{
		Choices: []openai.ChatCompletionChunkChoice{
			{FinishReason: "stop"},
		},
	}
	finishReason := ""
	response := ""
	err := a.processStreamChunk(chunk, &finishReason, &response, cb)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if finishReason != "stop" {
		t.Errorf("want finishReason %q, got %q", "stop", finishReason)
	}
	if received {
		t.Error("callback must not be called when content is empty")
	}
}

func TestProcessStreamChunk_Content_AccumulatesAndCallsCallback(t *testing.T) {
	a := newTestAgent(false)
	var cbArg string
	cb := func(content, _ string) error { cbArg = content; return nil }

	finishReason := ""
	response := ""
	err := a.processStreamChunk(chunkWithContent("hello"), &finishReason, &response, cb)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if response != "hello" {
		t.Errorf("want response %q, got %q", "hello", response)
	}
	if cbArg != "hello" {
		t.Errorf("want callback arg %q, got %q", "hello", cbArg)
	}
}

func TestProcessStreamChunk_CallbackError_Propagates(t *testing.T) {
	a := newTestAgent(false)
	want := errors.New("cb error")
	cb := func(_, _ string) error { return want }

	finishReason := ""
	response := ""
	err := a.processStreamChunk(chunkWithContent("x"), &finishReason, &response, cb)

	if err != want {
		t.Errorf("want %v, got %v", want, err)
	}
}
