# Stream Completion with Hugging Face Models

You need a token from Hugging Face to use their models. You can get one by signing up at [Hugging Face](https://huggingface.co/join) and creating an access token in your account settings.

✋ it works only for chat completion and tool calls for now. (no embeddings yet), see https://huggingface.co/docs/inference-providers/en/index:

```
Note: This OpenAI-compatible endpoint is currently available for chat completion tasks only. For other tasks like text-to-image, embeddings, or speech processing, use the Hugging Face inference clients shown above.
```

Create a file named `huggingface.env` in the root directory of the project with the following content:

```env
TOKEN=your_hugging_face_token_here
NOVA_LOG_LEVEL="INFO"
ENGINE_URL="https://router.huggingface.co/v1"
CHAT_MODEL="Qwen/Qwen2.5-7B-Instruct"
```