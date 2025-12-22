# Remote Agent - Interactive Operation Management

This example demonstrates interactive command-line control of tool call operations using a remote agent.

## Features

- Interactive CLI for managing operations
- Real-time validation/cancellation of tool calls
- Commands: validate, cancel, validate-all, reset, quit
- Operation ID tracking

## Prerequisites

- Server agent running on `http://localhost:8080` (see example 50)

## Running the Example

First, start the server in another terminal:
```bash
cd ../50-server-agent-with-tools
go run main.go
```

Then run this interactive client:
```bash
go run main.go
```

## Commands

Once the client is running and tool calls are detected, use these commands:

- `v <operation_id>` - Validate a specific operation
- `c <operation_id>` - Cancel a specific operation
- `va` - Validate all pending operations
- `r` - Reset (cancel) all pending operations
- `q` - Quit the program

## Example Session

```
🌐 Connected to remote agent at http://localhost:8080
Agent: Interactive Remote Bob Client
Model: hf.co/menlo/jan-nano-gguf:q4_k_m

=== Interactive Remote Agent Demo ===
This example demonstrates operation management with the remote agent.

Sending question that triggers multiple tool calls...

🔔 Tool Call Detected: Tool call detected: say_hello
📝 Operation ID: op_0x14000126020

=== Operation Management ===
Commands:
  v <operation_id>  - Validate an operation
  c <operation_id>  - Cancel an operation
  va                - Validate all pending operations
  r                 - Reset (cancel) all pending operations
  q                 - Quit

> v op_0x14000126020
✅ Operation op_0x14000126020 validated successfully

> r
🔄 All pending operations reset successfully

> q
Quitting...
```

## Use Cases

### Interactive Testing
Best for development and testing where you want real-time control without switching terminals.

### Manual Review
Review each operation before approving, useful for security-sensitive operations.

### Learning & Debugging
Understand the tool call flow and see exactly what operations are being requested.

## Related Examples

- [50-server-agent-with-tools](../50-server-agent-with-tools) - Server that this client connects to
- [51-remote-agent-stream](../51-remote-agent-stream) - Basic remote agent usage
- [53-remote-programmatic](../53-remote-programmatic) - Automated operation handling
