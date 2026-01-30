# Tool Response `name` Field — Local Model Compatibility Fix

## Problem

Some local LLM models (e.g., **FunctionGemma**, **jan-nano**) use Jinja chat templates that require a `name` field on tool response messages. When a tool call is executed and its result is sent back to the model, the response message must include the name of the function that was called.

The standard OpenAI API uses `tool_call_id` to correlate tool responses with tool calls, so the `name` field is not required. The `openai-go` SDK's `ToolMessage()` helper only sets `content` and `tool_call_id`.

This causes a **500 Internal Server Error** with local engines (e.g., llama.cpp) when the chat template tries to access `message['name']` and it doesn't exist:

```
Invalid tool response: 'name' must be provided.
```

## Solution Applied (Option 3 — Always Include `name`)

We chose to **always include the function name** in tool response messages. This is the approach implemented in the Nova SDK.

### How It Works

A new helper function `createToolResponseMessage` in `nova-sdk/agents/tools/tools.helpers.go` builds tool response messages using the OpenAI Go SDK's `SetExtraFields` mechanism to inject the `name` field:

```go
func createToolResponseMessage(content string, toolCallID string, functionName string) openai.ChatCompletionMessageParamUnion {
    toolMsg := openai.ToolMessage(content, toolCallID)
    toolMsg.OfTool.SetExtraFields(map[string]any{
        "name": functionName,
    })
    return toolMsg
}
```

The resulting JSON sent to the LLM engine looks like:

```json
{
  "role": "tool",
  "content": "{\"result\": 42}",
  "tool_call_id": "call_abc123",
  "name": "calculate_sum"
}
```

### Why This Is Safe

- OpenAI-compatible APIs **ignore unknown fields** — the extra `name` field causes no issues.
- Local models that **require** `name` now work correctly.
- No configuration needed — it works out of the box for all models.

## Alternative Approaches

### Option 1 — Add `name` Only in `processToolCalls()`

Instead of creating a reusable helper, the `name` field could be injected directly at the call site in `processToolCalls()`:

```go
toolMsg := openai.ToolMessage(result.Content, toolCall.ID)
toolMsg.OfTool.SetExtraFields(map[string]any{
    "name": functionName,
})
messages = append(messages, toolMsg)
```

**Pros:**
- Minimal change, localized to one line.
- No new function to maintain.

**Cons:**
- If other parts of the SDK also build tool response messages in the future, the fix won't apply there.
- Less discoverable — developers won't find a clearly-named helper explaining *why* the `name` is needed.

### Option 2 — Conditional via Model Config

Add a flag to `models.Config` such as `IncludeNameInToolResponse`:

```go
models.Config{
    Name:                     "hf.co/menlo/jan-nano-gguf:q4_k_m",
    Temperature:              models.Float64(0.0),
    IncludeNameInToolResponse: models.Bool(true),
}
```

The SDK would then check this flag before injecting `name`:

```go
if agent.ModelConfig.IncludeNameInToolResponse != nil && *agent.ModelConfig.IncludeNameInToolResponse {
    toolMsg.OfTool.SetExtraFields(map[string]any{
        "name": functionName,
    })
}
```

**Pros:**
- Explicit opt-in — no risk of side effects on models that don't need it.
- Clear in the sample code that this model has a special requirement.

**Cons:**
- Adds configuration complexity — users must know which models need this flag.
- Easy to forget, leading to the same cryptic 500 error.
- Since `name` is harmless for APIs that don't use it, the flag provides no real benefit.

## Affected Models

Models known to require this fix:

| Model | Template Requirement |
|-------|---------------------|
| FunctionGemma (`functiongemma-270m-it-gguf`) | `message['name']` in tool response |
| jan-nano (`jan-nano-gguf`) | `message['name']` in tool response |

Any model using a Jinja chat template that accesses `message['name']` on tool response messages will benefit from this fix.

## Files Changed

- `nova-sdk/agents/tools/tools.helpers.go` — Added `createToolResponseMessage()` helper and updated `processToolCalls()` to use it.
