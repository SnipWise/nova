# Nova Agent Builder Skill - Installation Guide

A Claude Code skill for generating AI agents in Go using the Nova SDK framework.

## 🎯 What is this?

The **Nova Agent Builder** skill enables Claude Code to generate production-ready AI agents in Go using the [Nova SDK](https://github.com/snipwise/nova). It includes:

- 📝 **Code snippets** for all agent types (chat, RAG, tools, orchestrator, crew, etc.)
- 📚 **API references** for the Nova SDK
- 🤖 **Automatic code generation** based on your requests
- ✨ **Best practices** and patterns for AI agent development

## 📦 Installation

Download and extract the skill to your project:

```bash
# Download the latest release
curl -LO https://github.com/snipwise/nova/releases/latest/download/nova-agent-builder-skill.zip

# Extract directly to your project
mkdir -p .claude/skills
unzip nova-agent-builder-skill.zip -d .claude/skills/

# Cleanup
rm nova-agent-builder-skill.zip
```

### From Git Repository

If you're already using the Nova SDK repository:

```bash
git clone https://github.com/snipwise/nova.git
cd nova
# The skill is already in .claude/skills/nova-agent-builder/
```

## ✅ Verification

After installation, verify the skill is available:

```bash
# Check directory structure
ls -la .claude/skills/nova-agent-builder/

# You should see:
# - SKILL.md (main skill documentation)
# - snippets/ (code templates)
# - references/ (API documentation)
# - install.sh (this installer)
```

## 🚀 Usage

Once installed, simply ask Claude Code to generate agents:

```
generate a chat agent with streaming
create a RAG agent for my FAQ system
generate a tools agent to calculate and send emails
create an orchestrator agent for topic detection
```

Claude will automatically:
1. Detect the `nova-agent-builder` skill
2. Select appropriate code snippets
3. Generate Go code using the Nova SDK
4. Follow best practices and patterns

## 📚 Available Agent Types

| Type | Description | Example Request |
|------|-------------|-----------------|
| **Chat** | Conversational agents with streaming | "generate a streaming chat agent" |
| **RAG** | Retrieval-Augmented Generation | "create a RAG agent with vector search" |
| **Tools** | Function calling agents | "generate a tools agent with parallel execution" |
| **Structured** | Typed output agents | "create an agent with structured JSON output" |
| **Orchestrator** | Topic detection & routing | "generate an orchestrator for multi-agent routing" |
| **Crew** | Multi-agent collaboration | "create a crew agent with 3 specialized agents" |
| **Pipeline** | Sequential agent chains | "generate a pipeline for document processing" |

## 🛠️ Configuration

Default configuration (customize in your project):

```yaml
# config.yaml or .env
ENGINE_URL: "http://localhost:12434/engines/llama.cpp/v1"
CHAT_MODEL: "ai/qwen2.5:1.5B-F16"
TOOLS_MODEL: "hf.co/menlo/jan-nano-gguf:q4_k_m"
ORCHESTRATOR_MODEL: "hf.co/menlo/lucy-gguf:q4_k_m"
EMBEDDING_MODEL: "ai/mxbai-embed-large"
```

## 📖 Documentation

- **Skill Documentation**: `.claude/skills/nova-agent-builder/SKILL.md`
- **Project Instructions**: `CLAUDE.md` (in your project root)
- **Code Snippets**: `.claude/skills/nova-agent-builder/snippets/`
- **API References**: `.claude/skills/nova-agent-builder/references/`
- **Nova SDK**: https://github.com/snipwise/nova

## 🔧 Troubleshooting

### Skill not detected by Claude Code

1. Verify directory structure:
   ```bash
   ls .claude/skills/nova-agent-builder/SKILL.md
   ```
2. Restart Claude Code
3. Try explicitly mentioning the skill:
   ```
   use the nova-agent-builder skill to create a chat agent
   ```

### Claude generates wrong language (not Go)

Be explicit in your request:
```
generate a chat agent IN GO using Nova SDK
```

### Missing snippets or references

Re-run the installation script or manually verify all directories:
```bash
ls .claude/skills/nova-agent-builder/snippets/
ls .claude/skills/nova-agent-builder/references/
```

## 🆘 Support

- **Issues**: https://github.com/snipwise/nova/issues
- **Documentation**: https://github.com/snipwise/nova/blob/main/CLAUDE.md
- **Examples**: https://github.com/snipwise/nova/tree/main/samples

## 📄 License

This skill is part of the Nova SDK project. See the main repository for license information.

---

**Happy coding with Nova! 🚀**
