# Nova Crew Server - Web Chat Interface

A modern, responsive web interface for interacting with the Nova Crew Server Agent.

## Features

- **Real-time Streaming**: SSE (Server-Sent Events) based streaming for instant responses
- **Markdown Rendering**: Full markdown support with syntax highlighting for code blocks
- **Multi-Agent Support**: Automatic routing between specialized agents (coder, thinker, cook, generic)
- **Function Calling**: Visual controls for validating/canceling tool calls (human-in-the-loop)
- **RAG Integration**: Context retrieval from document embeddings
- **Context Management**: Real-time context size monitoring
- **Memory Controls**: Reset conversation, view messages, manage operations

## Technology Stack

- **Vue.js 3**: Progressive JavaScript framework (CDN-based, no build required)
- **Marked.js**: Markdown parsing
- **Highlight.js**: Syntax highlighting for code blocks
- **Vanilla CSS**: Custom responsive styles

## Project Structure

```
web/
├── index.html                      # Main HTML entry point
├── js/
│   ├── api.js                      # API service layer
│   ├── markdown.js                 # Markdown rendering utilities
│   ├── app.js                      # Main Vue application
│   └── components/
│       ├── ChatMessage.js          # Message component
│       ├── InputBar.js             # Input and action buttons
│       ├── StatusBar.js            # Context and model info
│       └── OperationControls.js    # Tool call validation controls
└── README.md                       # This file
```

## Setup

### Prerequisites

1. **Go Server Running**: Make sure the Nova Crew Server Agent is running:
   ```bash
   cd samples/56-crew-server-agent
   go run main.go
   ```

   The server should start on `http://localhost:8080`

### Option 1: Serve via Go Server (Recommended)

Update your Go server to serve static files (see main.go modifications below).

### Option 2: Simple HTTP Server

Use any static file server:

```bash
# Python 3
cd web
python3 -m http.server 3000

# Node.js (http-server)
npx http-server web -p 3000

# PHP
cd web
php -S localhost:3000
```

Then open `http://localhost:3000` in your browser.

## Usage

### Sending Messages

1. Type your message in the input area
2. Press **Enter** to send (or click "Send" button)
3. Use **Shift+Enter** for new lines
4. Watch responses stream in real-time with markdown formatting

### Agent Routing

The system automatically routes your question to specialized agents:

- **Coder Agent**: Programming, coding, debugging questions
- **Thinker Agent**: Philosophy, math, science, psychology
- **Cook Agent**: Cooking, recipes, food-related queries
- **Generic Agent**: Everything else

### Function Calling

When the agent wants to call a tool:

1. A notification appears with the operation details
2. Click **Validate** to approve the operation
3. Click **Cancel** to reject it
4. The agent proceeds based on your choice

### Action Buttons

- **📤 Send**: Send your message
- **⏹ Stop**: Stop current streaming response
- **🗑 Clear Memory**: Reset conversation (keeps system instruction)
- **💬 View Messages**: Show all messages in console
- **🤖 View Models**: Display model information
- **🔄 Reset Operations**: Clear all pending operations

### Status Bar

Real-time information display:

- **Agent**: Currently active agent
- **Context Size**: Current conversation context size
- **Chat Model**: Model used for chat
- **Tools**: Model used for function calling
- **RAG**: Embedding model for retrieval

## API Endpoints Used

The web interface communicates with these endpoints:

- `POST /completion` - Send message and receive streaming response
- `POST /completion/stop` - Stop current streaming
- `POST /memory/reset` - Clear conversation memory
- `GET /memory/messages/list` - Get all messages
- `GET /memory/messages/context-size` - Get context size
- `POST /operation/validate` - Approve tool call
- `POST /operation/cancel` - Reject tool call
- `POST /operation/reset` - Clear all pending operations
- `GET /models` - Get model information
- `GET /health` - Health check

## Customization

### Styling

All styles are in `index.html` within the `<style>` tag. Key CSS classes:

- `.message.user` - User messages
- `.message.assistant` - AI responses
- `.message.system` - System messages
- `.operation-controls` - Tool call notifications

### Colors

Current color scheme (dark theme):

- Background: `#1a1a1a`
- Cards: `#2d2d2d`
- Primary: `#4fc3f7` (blue)
- Success: `#43a047` (green)
- Danger: `#e53935` (red)
- Warning: `#fb8c00` (orange)

### API Configuration

Change the API base URL in `js/api.js`:

```javascript
const API_BASE_URL = 'http://localhost:8080'; // Change to your server URL
```

## Troubleshooting

### Connection Failed

**Error**: "Failed to connect to server"

**Solution**:
1. Ensure Go server is running: `go run main.go`
2. Check server is on port 8080
3. Verify no CORS issues (Go server should allow CORS)

### Streaming Not Working

**Error**: Messages not streaming

**Solution**:
1. Check browser console for errors
2. Ensure SSE is supported (all modern browsers)
3. Verify `/completion` endpoint is working

### Code Blocks Not Highlighting

**Error**: Code appears without syntax highlighting

**Solution**:
1. Check highlight.js CDN is loaded
2. Verify language is specified in code fence (```go, ```python, etc.)
3. Check browser console for errors

### Markdown Not Rendering

**Error**: Markdown appears as plain text

**Solution**:
1. Check marked.js CDN is loaded
2. Verify `js/markdown.js` is loaded before `js/app.js`
3. Check browser console for errors

## Browser Support

- Chrome/Edge: ✅ Full support
- Firefox: ✅ Full support
- Safari: ✅ Full support
- Mobile browsers: ✅ Responsive design

## Performance

- **Bundle Size**: ~250KB (all dependencies via CDN)
- **Initial Load**: < 1 second
- **Streaming Latency**: Real-time (SSE)
- **Memory**: Efficient Vue.js 3 reactivity

## Security Notes

- **Local Development**: This is designed for local development
- **Production**: Add authentication, rate limiting, and HTTPS for production use
- **CORS**: Ensure proper CORS configuration on the Go server

## Future Enhancements

Potential improvements:

- [ ] Dark/Light theme toggle
- [ ] Message export (JSON, Markdown)
- [ ] Multi-session support
- [ ] Voice input
- [ ] Copy code blocks to clipboard
- [ ] Message search and filtering
- [ ] Agent selection override
- [ ] File upload support
- [ ] Custom system instructions

## License

Same as Nova SDK project.
