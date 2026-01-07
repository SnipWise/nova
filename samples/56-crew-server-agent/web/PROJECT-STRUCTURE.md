# Project Structure

Complete web interface for Nova Crew Server Agent built with Vue.js 3.

## File Organization

```
web/
├── 📄 index.html                    # Main entry point (HTML + inline CSS)
├── 📁 js/                           # JavaScript modules
│   ├── api.js                       # API service layer (CrewServerAPI class)
│   ├── markdown.js                  # Markdown rendering utilities
│   ├── app.js                       # Main Vue.js application
│   └── components/                  # Vue components
│       ├── ChatMessage.js           # Message display component
│       ├── InputBar.js              # User input and action buttons
│       ├── StatusBar.js             # Context size and model info
│       └── OperationControls.js     # Tool call validation controls
├── 📄 README.md                     # Full documentation
├── 📄 QUICKSTART.md                 # Quick start guide
├── 📄 PROJECT-STRUCTURE.md          # This file
├── 📄 demo-questions.md             # Example questions for testing
├── 🔧 start.sh                      # Launch script (macOS/Linux)
└── 🔧 start.bat                     # Launch script (Windows)
```

## Component Architecture

```
App (app.js)
├── Header
│   ├── Title
│   └── StatusBar (StatusBar.js)
│       ├── Agent name
│       ├── Context size
│       └── Model information
├── Chat Container
│   ├── ChatMessage[] (ChatMessage.js)
│   │   ├── Role display
│   │   └── Markdown content
│   └── OperationControls[] (OperationControls.js)
│       ├── Operation info
│       └── Validate/Cancel buttons
└── InputBar (InputBar.js)
    ├── Textarea (user input)
    └── Action Buttons
        ├── Send
        ├── Stop
        ├── Clear Memory
        ├── View Messages
        ├── View Models
        └── Reset Operations
```

## Data Flow

```
User Input
    ↓
InputBar Component (emits 'send' event)
    ↓
App Component (handles send)
    ↓
CrewServerAPI.sendMessage()
    ↓
HTTP POST /completion (SSE Stream)
    ↓
Server-Sent Events ← Go Server
    ↓
API callbacks:
    ├── onChunk → Update message content
    ├── onNotification → Show operation controls
    └── onError → Display error
    ↓
Vue Reactivity Updates UI
    ↓
ChatMessage renders markdown
```

## State Management

All state is managed in the main App component using Vue 3 Composition API:

| State Variable | Type | Purpose |
|---|---|---|
| `messages` | `ref([])` | Conversation history |
| `contextSize` | `ref(0)` | Current context size |
| `models` | `ref({})` | Model information |
| `selectedAgent` | `ref('generic')` | Active agent name |
| `isLoading` | `ref(false)` | Streaming status |
| `error` | `ref(null)` | Error messages |
| `pendingOperations` | `ref([])` | Tool call operations |
| `streamingMessageIndex` | `ref(-1)` | Currently streaming message |

## API Endpoints

| Endpoint | Method | Purpose | Component |
|---|---|---|---|
| `/completion` | POST | Send message & stream response | App → API |
| `/completion/stop` | POST | Stop current stream | InputBar → API |
| `/memory/reset` | POST | Clear conversation | InputBar → API |
| `/memory/messages/list` | GET | Get all messages | InputBar → API |
| `/memory/messages/context-size` | GET | Get context size | App → API (polling) |
| `/operation/validate` | POST | Approve tool call | OperationControls → API |
| `/operation/cancel` | POST | Reject tool call | OperationControls → API |
| `/operation/reset` | POST | Clear operations | InputBar → API |
| `/models` | GET | Get model info | App → API |
| `/health` | GET | Health check | App → API |

## External Dependencies (CDN)

| Library | Version | Purpose | Size |
|---|---|---|---|
| Vue.js 3 | 3.4.15 | Reactive UI framework | ~150KB |
| Marked.js | 11.1.1 | Markdown parser | ~50KB |
| Highlight.js | 11.9.0 | Syntax highlighting | ~100KB |
| **Total** | | | **~300KB** |

## Styling Approach

- **No CSS Framework**: Custom vanilla CSS for minimal bundle size
- **Inline Styles**: All CSS in `index.html` `<style>` tag
- **Dark Theme**: Modern dark color scheme optimized for readability
- **Responsive**: Mobile-first design with media queries
- **CSS Variables**: Not used (for broader browser support)

## Key Features Implementation

### 1. Streaming (SSE)

```javascript
// api.js - sendMessage()
const reader = response.body.getReader();
while (!done) {
    const { value, done } = await reader.read();
    // Parse SSE format: "data: {...}\n\n"
    onChunk(content, isComplete);
}
```

### 2. Markdown Rendering

```javascript
// markdown.js - render()
marked.setOptions({
    highlight: (code, lang) => hljs.highlight(code, { language: lang })
});
return marked.parse(markdownText);
```

### 3. Progressive Code Highlighting

```javascript
// markdown.js - renderStreaming()
if (hasIncompleteCodeBlock) {
    // Temporarily close for rendering
    const textWithClosedBlock = markdownText + '\n```';
    return marked.parse(textWithClosedBlock);
}
```

### 4. Auto-scroll

```javascript
// app.js - watch messages
Vue.watch(() => messages.value.length, () => {
    Vue.nextTick(() => scrollToBottom());
});
```

### 5. Context Size Polling

```javascript
// app.js - startContextSizePolling()
setInterval(async () => {
    const size = await api.getContextSize();
    contextSize.value = size;
}, 2000); // Every 2 seconds
```

## Performance Optimizations

1. **CDN Loading**: All dependencies loaded from fast CDNs
2. **No Build Step**: Instant development, no webpack/vite needed
3. **Lazy Rendering**: Only visible messages rendered
4. **Efficient Reactivity**: Vue 3 Proxy-based reactivity
5. **Stream Processing**: Chunk-by-chunk, no buffering entire response
6. **DOM Updates**: Batched via Vue.nextTick()

## Browser Compatibility

| Browser | Version | Support |
|---|---|---|
| Chrome | 90+ | ✅ Full |
| Firefox | 88+ | ✅ Full |
| Safari | 14+ | ✅ Full |
| Edge | 90+ | ✅ Full |
| Mobile Chrome | Latest | ✅ Full |
| Mobile Safari | Latest | ✅ Full |

## Security Considerations

⚠️ **Current Setup**: Development only

For production deployment:

1. **CORS**: Restrict allowed origins
2. **HTTPS**: Use TLS for all connections
3. **Authentication**: Add user authentication
4. **Rate Limiting**: Prevent API abuse
5. **Input Validation**: Sanitize user input
6. **CSP Headers**: Content Security Policy
7. **XSS Protection**: Already handled by Vue's text interpolation

## Customization Guide

### Change Theme Colors

Edit `index.html` `<style>` section:

```css
/* Background colors */
body { background: #1a1a1a; }
.header { background: #2d2d2d; }

/* Accent colors */
.status-value { color: #4fc3f7; } /* Primary blue */
button.success { background: #43a047; } /* Green */
button.danger { background: #e53935; } /* Red */
```

### Change API URL

Edit `js/api.js`:

```javascript
const API_BASE_URL = 'http://your-server:port';
```

### Add Custom Actions

1. Add button in `components/InputBar.js`
2. Emit custom event
3. Handle in `app.js`
4. Call API method

### Modify Markdown Rendering

Edit `js/markdown.js`:

```javascript
marked.setOptions({
    // Your custom options
});
```

## Testing Checklist

- [ ] Send message and receive streaming response
- [ ] Markdown renders correctly (headers, lists, code)
- [ ] Code blocks have syntax highlighting
- [ ] Can stop streaming mid-response
- [ ] Clear memory resets conversation
- [ ] Context size updates in real-time
- [ ] View messages shows all messages
- [ ] View models displays model info
- [ ] Tool call notifications appear
- [ ] Can validate tool calls
- [ ] Can cancel tool calls
- [ ] Reset operations clears pending ops
- [ ] Agent routing works (coder/thinker/cook/generic)
- [ ] Error messages display correctly
- [ ] Mobile responsive layout works
- [ ] Browser refresh preserves no state (expected)

## Development Workflow

1. **Start Go Server**:
   ```bash
   cd samples/56-crew-server-agent
   go run main.go
   ```

2. **Start Web Server**:
   ```bash
   cd web
   ./start.sh  # or start.bat on Windows
   ```

3. **Open Browser**:
   ```
   http://localhost:3000
   ```

4. **Edit Code**:
   - Edit any `.js` file
   - Refresh browser (no build needed)

5. **Debug**:
   - Open DevTools (F12)
   - Check Console for logs
   - Check Network tab for API calls

## Future Enhancements

Potential improvements (not implemented):

- [ ] Persistent conversation history (localStorage)
- [ ] Export conversation (JSON/Markdown)
- [ ] Dark/Light theme toggle
- [ ] Voice input via Web Speech API
- [ ] Copy code blocks to clipboard
- [ ] Message search and filter
- [ ] Manual agent selection
- [ ] File upload for RAG
- [ ] Multi-tab conversations
- [ ] Keyboard shortcuts
- [ ] Notification sounds
- [ ] PWA support (offline mode)

## Troubleshooting Tips

| Problem | Solution |
|---|---|
| Blank page | Check console for JavaScript errors |
| 404 errors | Verify file paths are correct |
| CORS errors | Ensure Go server has CORS enabled |
| No streaming | Check SSE support in browser |
| Markdown not rendering | Verify marked.js loaded |
| No syntax highlighting | Verify highlight.js loaded |
| Slow performance | Check network tab for slow CDN |

## License

Same as Nova SDK project.

---

Built with ❤️ using Vue.js 3, Marked.js, and Highlight.js
