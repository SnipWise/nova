# Demo Questions for Nova Crew Server

Test the different agents and features with these example questions.

## 🔵 Coder Agent (Programming Questions)

These questions will be routed to the **coder agent** specialized in programming:

```
Write a Go function that reads a JSON file and returns a struct
```

```
How do I implement a binary search tree in Python?
```

```
Explain the difference between mutex and channels in Go
```

```
Debug this code: for i in range(10) print(i)
```

## 🟢 Thinker Agent (Philosophy, Science, Math)

These questions will be routed to the **thinker agent** for deep thinking:

```
What is the relationship between consciousness and free will?
```

```
Explain the Monty Hall problem step by step
```

```
What are the implications of Gödel's incompleteness theorems?
```

```
How does quantum entanglement work?
```

## 🟠 Cook Agent (Food & Recipes)

These questions will be routed to the **cook agent** for culinary expertise:

```
Give me a recipe for homemade pizza dough
```

```
How do I make perfect scrambled eggs?
```

```
What can I substitute for eggs in a cake recipe?
```

```
Give me a meal plan for a week of healthy dinners
```

## ⚪ Generic Agent (Everything Else)

These questions will be handled by the **generic agent**:

```
What's the capital of France?
```

```
Tell me about the history of the internet
```

```
What are the best practices for remote work?
```

## 🔧 Tool Calling (Function Calling)

These commands will trigger the **tools agent** and show operation controls:

```
Say hello to Bob
```

```
Calculate the sum of 123 and 456
```

```
Say exit
```

**Expected behavior**:
- A notification will appear asking you to validate or cancel the operation
- Click "Validate" to execute the tool
- Click "Cancel" to reject it

## 📚 RAG (Document Retrieval)

If you have documents in the `data/` folder, these questions will use **RAG**:

```
What information do you have about [topic in your documents]?
```

```
Search for details about [specific term]
```

## 🧪 Context Compression

Try long conversations to see **compression** in action:

1. Ask several questions in a row
2. Watch the context size grow
3. When it reaches ~8500 tokens, the compressor activates
4. Context gets compressed while maintaining key information

## 💬 Memory Management

Test memory controls:

1. Have a conversation with several messages
2. Click "View Messages" to see all messages in console
3. Click "Clear Memory" to reset
4. Verify the conversation starts fresh

## 📊 Model Information

Click "View Models" to see:
- Chat model: `hf.co/qwen/qwen2.5-coder-3b-instruct-gguf:q4_k_m`
- Tools model: `hf.co/menlo/jan-nano-gguf:q4_k_m`
- RAG embedding model: `ai/mxbai-embed-large:latest`
- Compressor model: `ai/qwen2.5:0.5B-F16`

## 🎨 Markdown Rendering

Test markdown features:

```
Explain quicksort with:
1. A bullet list of steps
2. Code examples in Python
3. Big O notation in a blockquote
4. A final summary
```

Expected output:
- ✅ Numbered/bulleted lists
- ✅ Code blocks with syntax highlighting
- ✅ Blockquotes
- ✅ Headers
- ✅ Inline code

## ⚡ Streaming

All responses stream token-by-token in real-time. You can:

- Click "Stop" mid-stream to interrupt
- Watch markdown render progressively
- See code blocks highlight as they complete

## 🔀 Agent Switching

The orchestrator automatically detects topics and switches agents:

```
First, write a function to sort an array
```
*(Routes to coder)*

```
Now, what's the philosophy behind functional programming?
```
*(Routes to thinker)*

```
Finally, give me a recipe that uses sorted ingredients
```
*(Routes to cook)*

## Advanced Scenarios

### Multi-turn Conversation

```
User: Write a function to calculate fibonacci
[AI responds]

User: Now optimize it with memoization
[AI responds with improved version]

User: Explain the time complexity difference
[AI analyzes both versions]
```

### Tool Chain

```
Say hello to Alice
[Validate]

Calculate the sum of 10 and 20
[Validate]

Say exit
[This returns an error - test error handling]
```

### Context Overflow Test

Send very long messages or many messages to trigger compression:

```
Explain in detail the entire history of computer programming from the 1800s to today, including all major languages, paradigms, and influential people
```

Happy testing! 🚀
