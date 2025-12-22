# Quick Start Guide

## 1. Start the Server

In one terminal:
```bash
cd samples/50-server-agent-with-tools
go run main.go
```

You should see:
```
🚀 Server starting on http://localhost:8080
📡 Endpoints:
  POST   /completion
  POST   /completion/stop
  ...
```

## 2. Run the Client

In another terminal:
```bash
cd samples/51-remote-agent-stream
go run main.go
```

The client will:
1. Check if the server is healthy
2. Retrieve models information from the server
3. Execute the examples

## 3. Handle Tool Calls

When the client detects a tool call, it will display:

```
🔔 Tool Call Detected: Tool call detected: say_hello
📝 Operation ID: op_0x14000126020
✅ To validate: curl -X POST http://localhost:8080/operation/validate -d '{"operation_id":"op_0x14000126020"}'
⛔️ To cancel:   curl -X POST http://localhost:8080/operation/cancel -d '{"operation_id":"op_0x14000126020"}'
```

### Option A: Use Helper Scripts (Recommended)

In a **third terminal**:
```bash
cd samples/51-remote-agent-stream

# Copy the operation_id from the client output and run:
./validate-operation.sh op_0x14000126020

# Or to cancel:
./cancel-operation.sh op_0x14000126020
```

### Option B: Use curl Directly

Copy the curl command from the client output and run it in another terminal.

## Example Flow

1. **Client sends**: "Say hello to Alice and Bob"
2. **Server detects** tool calls: `say_hello(name="Alice")` and `say_hello(name="Bob")`
3. **Client displays** two operation IDs with validation commands
4. **You validate** each operation using the scripts or curl
5. **Server executes** the tools and continues with the response
6. **Client receives** the final response with tool results

## Testing Scripts

- `run-demo.sh` - Checks if server is running and starts the client
- `validate-operation.sh <op_id>` - Validate a pending operation
- `cancel-operation.sh <op_id>` - Cancel a pending operation

## Health Check

The client automatically checks server health at startup. If the server is not available, the client will exit with an error message.

You can also manually check server health:
```bash
curl http://localhost:8080/health
```

Expected response:
```json
{"status":"ok"}
```

## Tips

- Keep 3 terminals open: server, client, and validation
- The client shows the exact commands to run for validation
- Operations timeout if not validated within a reasonable time
- You can validate operations in any order (parallel tool calls)
- The client will display detailed models information at startup
