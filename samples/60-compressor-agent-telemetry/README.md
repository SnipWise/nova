# Compressor Agent with Telemetry

This example demonstrates telemetry tracking for context compression operations.

## Features

- Compress long conversation histories
- Track compression efficiency metrics
- Monitor token usage for compression operations
- Calculate compression ratios and savings
- Compare input vs output sizes

## Key Metrics

The telemetry for compression includes:
- **Input Context Length**: Size of original conversation
- **Output Tokens**: Tokens used to generate compressed version
- **Compression Ratio**: Percentage of space saved
- **Processing Time**: Time taken to compress
- **Characters per Token**: Efficiency metric

## Use Cases

Compression telemetry is useful for:
1. **Optimizing Context Windows**: Monitor when to compress
2. **Cost Management**: Track token usage for compression
3. **Performance Tuning**: Measure compression speed
4. **Quality Metrics**: Ratio of compression vs information loss

## Running

```bash
cd samples/60-compressor-agent-telemetry
go run main.go
```

## Expected Output

The example compresses a multi-turn conversation and shows:
1. Original conversation size
2. Compressed output size
3. Compression ratio achieved
4. Token costs for the compression operation
5. Full telemetry data
