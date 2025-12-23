# Structured Agent with Telemetry

This example demonstrates telemetry tracking for structured data generation using JSON schemas.

## Features

- Define custom Go structs with JSON schema annotations
- Extract structured data from unstructured text
- Track token usage for schema-based generation
- Monitor complexity of structured outputs
- Inspect JSON schema sent in requests

## Key Metrics

Structured data generation telemetry includes:
- **JSON Schema Size**: How much context the schema uses
- **Tokens per Field**: Efficiency of structured extraction
- **Nested Object Tracking**: Complexity metrics
- **Validation Success**: Whether output matches schema

## Structured Data Benefits

Using structured agents with telemetry helps:
1. **Optimize Schema Design**: Track schema complexity vs token usage
2. **Cost Estimation**: Predict costs for different data structures
3. **Performance Monitoring**: Measure extraction speed
4. **Quality Assurance**: Verify all fields are extracted

## Running

```bash
cd samples/61-structured-agent-telemetry
go run main.go
```

## Expected Output

The example:
1. Defines a `Person` struct with nested `Location`
2. Extracts structured data from narrative text
3. Shows the generated JSON schema
4. Displays telemetry for the structured generation
5. Calculates metrics like tokens per field
