# Chat Agent with Telemetry Callbacks

This example demonstrates how to use real-time telemetry callbacks to monitor LLM interactions as they happen.

## Features Demonstrated

- **Real-time Monitoring**: Get notified immediately when requests are sent and responses are received
- **Custom Telemetry Logger**: Implement your own telemetry handler
- **Streaming Support**: Monitor individual chunks in streaming responses
- **Error Tracking**: Capture and log errors as they occur
- **Session Analytics**: Track cumulative metrics across the entire session

## How It Works

The example implements a custom `TelemetryCallback` interface:

```go
type TelemetryCallback interface {
    OnRequestSent(metadata base.RequestMetadata, requestJSON string)
    OnResponseReceived(metadata base.ResponseMetadata, responseJSON string)
    OnStreamChunk(chunk string, index int)
    OnError(err error, context string)
}
```

## Custom Telemetry Logger

The example includes a complete custom telemetry logger that:

- Counts requests, responses, and errors
- Tracks total token usage across all requests
- Calculates average tokens per response
- Measures session duration
- Displays real-time notifications for each event

## Key Implementation

```go
// Create custom telemetry logger
telemetryLogger := NewCustomTelemetryLogger()

// Create agent
agent, _ := chat.NewAgent(ctx, agentConfig, modelConfig)

// Register the callback
agent.SetTelemetryCallback(telemetryLogger)

// Now all requests/responses will trigger the callback automatically
result, _ := agent.GenerateCompletion(messages)
```

## Use Cases

This callback-based telemetry is ideal for:

1. **External Monitoring Systems**: Send metrics to Prometheus, Datadog, etc.
2. **Real-time Dashboards**: Update UI with live request/response data
3. **Cost Tracking**: Calculate running costs in real-time
4. **Performance Alerting**: Trigger alerts when response times exceed thresholds
5. **Debug Logging**: Write detailed logs to external systems
6. **Analytics Pipelines**: Feed data into analytics platforms
7. **Compliance/Audit**: Maintain detailed audit trails

## Integration Examples

### Prometheus Metrics
```go
func (logger *TelemetryLogger) OnResponseReceived(metadata base.ResponseMetadata, responseJSON string) {
    prometheusTokenCounter.Add(float64(metadata.TotalTokens))
    prometheusLatencyHistogram.Observe(float64(metadata.ResponseTime))
}
```

### JSON Logging
```go
func (logger *TelemetryLogger) OnRequestSent(metadata base.RequestMetadata, requestJSON string) {
    log.Printf("REQUEST: %s", requestJSON)
}
```

### Error Tracking
```go
func (logger *TelemetryLogger) OnError(err error, context string) {
    sentry.CaptureException(err)
}
```

## Running the Example

```bash
# Make sure you have a compatible LLM server running
cd samples/58-chat-agent-telemetry-callback
go run main.go
```

## Expected Output

The example will:
1. Register a custom telemetry callback
2. Send three questions to the LLM
3. Display real-time telemetry events as they occur:
   - Request sent notifications with metadata
   - Response received notifications with token counts
   - Running totals
4. Show a final summary with:
   - Session duration
   - Total requests/responses
   - Total tokens used
   - Average tokens per response

## Advanced Usage

You can implement multiple callbacks for different purposes:

```go
agent.SetTelemetryCallback(&MultiCallback{
    callbacks: []base.TelemetryCallback{
        &MetricsCollector{},
        &AuditLogger{},
        &CostTracker{},
    },
})
```

## Next Steps

- Integrate with your monitoring infrastructure
- Add custom business logic to track specific metrics
- Implement streaming telemetry for long-running requests
- Build real-time dashboards using the callback data
