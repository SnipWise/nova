# Nova Agent Builder Skill for Claude Code

Generate production-ready Nova agents with Claude Code! The **nova-agent-builder** skill enables automatic Go code generation for all agent types (chat, RAG, tools, orchestrator, crew, etc.).

## Installing the Skill

```bash
# Download and extract the skill
curl -LO https://github.com/snipwise/nova/releases/latest/download/nova-agent-builder-skill.zip
mkdir -p .claude/skills
unzip nova-agent-builder-skill.zip -d .claude/skills/
rm nova-agent-builder-skill.zip
```

If you've cloned this repository, the skill is already available in `.claude/skills/nova-agent-builder/`

## Using the Skill

Once installed, ask Claude Code to generate agents:

```
generate a chat agent with streaming
create a RAG agent for my FAQ system
generate a tools agent with parallel execution
create an orchestrator agent for topic detection
```

See [.claude/skills/nova-agent-builder/INSTALL.md](.claude/skills/nova-agent-builder/INSTALL.md) for complete documentation.
