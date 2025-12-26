# Generated with skills/nova-agent-builder

✋✋✋ **This directory contains example agents generated using the Nova Agent Builder skill.**

## Examples

### 01-chat-agent
**Prompt:** "In the /generated-with-skills folder, generate a chat agent in a new numbered directory"

A streaming chat agent demonstrating real-time conversational AI capabilities.

### 02-orchestrator-cooking
**Prompt:** "In /generated-with-skills, create an orchestrator agent that will detect various topics about cooking in a numbered subdirectory"

An orchestrator agent that detects and categorizes cooking-related topics across 14 different categories (recipes, techniques, ingredients, etc.).

### 03-structured-actions
**Prompt:** "In /generated-with-skills, create a structured agent using a structure to detect a list of actions to do"

A structured agent that extracts action items from unstructured text (emails, meeting notes, to-do lists) with priority, category, deadlines, and time estimates.

## Usage

Each example is self-contained with:
- Complete source code (`main.go`)
- Dependencies (`go.mod`)
- Comprehensive documentation (`README.md`)
- Additional usage examples where applicable

To run any example:
```bash
cd <example-directory>
go mod tidy
go run main.go
```
