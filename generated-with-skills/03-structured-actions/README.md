# 03 - Structured Agent for Action Item Extraction

> Generated with Nova Agent Builder skill

## Description

A structured agent that extracts and organizes action items from unstructured text. Automatically identifies tasks, priorities, categories, deadlines, and estimates from emails, meeting notes, project documents, and personal notes.

## Features

- **Complete Action Extraction**: Identifies all explicit and implicit action items
- **Rich Metadata**: Priority, category, due dates, time estimates, and tags
- **Smart Classification**: Automatically categorizes actions (work, personal, shopping, etc.)
- **Structured JSON Output**: Type-safe Go structs with validation
- **Multiple Scenarios**: Handles emails, meeting notes, project plans, personal to-dos
- **Statistics & Analytics**: Priority distribution, category breakdown

## Use Cases

- **Email Processing**: Extract action items from team emails
- **Meeting Notes**: Convert meeting minutes into actionable tasks
- **Project Management**: Parse project documents for tasks and deadlines
- **Personal Productivity**: Organize daily/weekly to-do lists
- **Team Coordination**: Identify and assign tasks from discussions
- **Task Automation**: Feed extracted actions into task management systems

## Prerequisites

- Model `ai/qwen2.5:1.5B-F16` available

## Installation

```bash
cd generated-with-skills/03-structured-actions
go mod init structured-actions
go mod tidy
```

## Usage

```bash
go run main.go
```

## Data Structures

### Action

```go
type Action struct {
    ID          string   `json:"id"`              // Unique identifier
    Description string   `json:"description"`
    Priority    string   `json:"priority"`        // high, medium, low
    Category    string   `json:"category"`        // work, personal, etc.
    DueDate     string   `json:"due_date"`        // YYYY-MM-DD
    Tags        []string `json:"tags"`
    Estimated   string   `json:"estimated_time"`  // "30min", "2h"
}
```

### ActionList

```go
type ActionList struct {
    Actions     []Action `json:"actions"`
    TotalCount  int      `json:"total_count"`
    Summary     string   `json:"summary"`
    HasDeadline bool     `json:"has_deadline"`
}
```

## Example Output

```
📋 Action Item Extraction Demo
================================================================================

[Scenario 1] Email with Multiple Tasks
--------------------------------------------------------------------------------
Input:
Hi team,

After today's meeting, here are the action items:

1. John needs to update the documentation by Friday
2. Sarah will schedule a follow-up meeting with the client for next week
3. I'll review the code changes and provide feedback by EOD tomorrow
4. Don't forget to submit your timesheets before the end of the month
5. We should order new office supplies - we're running low on printer paper

Also, can someone pick up coffee for the break room?

Thanks!
--------------------------------------------------------------------------------

✅ Extracted 6 actions (finish_reason: stop)

📊 Summary: Team coordination tasks including documentation, meetings, code review,
           timesheets, office supplies, and coffee

📝 Action Items:

   [1] Update the documentation
       Priority: 🔴 High | Category: work | Due: 2024-01-05 | Time: 2h

   [2] Schedule follow-up meeting with client
       Priority: 🟡 Medium | Category: work | Due: 2024-01-08
       Tags: [meeting, client]

   [3] Review code changes and provide feedback
       Priority: 🔴 High | Category: work | Due: 2024-01-02 | Time: 1h

   [4] Submit timesheets
       Priority: 🟡 Medium | Category: administrative | Due: 2024-01-31

   [5] Order office supplies (printer paper)
       Priority: 🟢 Low | Category: office | Time: 15min
       Tags: [supplies, procurement]

   [6] Pick up coffee for break room
       Priority: 🟢 Low | Category: office | Time: 10min

📈 Statistics:
   Priorities: High: 2 | Medium: 2 | Low: 2
   Categories: work: 3 | administrative: 1 | office: 2 |
   ⏰ Contains time-sensitive actions!
```

## Configuration

### Agent Settings

```go
agents.Config{
    Name:      "action-extractor",
    EngineURL: "http://localhost:12434/engines/llama.cpp/v1",
    SystemInstructions: `Extract all actionable tasks and to-dos.
                        Identify priority, category, deadlines, and estimates.`,
}
```

### Model Settings

```go
models.Config{
    Name:        "hf.co/menlo/jan-nano-gguf:q4_k_m",
    Temperature: models.Float64(0.0),  // Critical for consistent JSON
}
```

## Integration Examples

### With Task Management System

```go
import "your-task-system/api"

// Extract actions
actionList, _, _ := agent.GenerateStructuredData(messages)

// Create tasks in your system
for _, action := range actionList.Actions {
    task := api.Task{
        Title:       action.Description,
        Priority:    action.Priority,
        Category:    action.Category,
        DueDate:     action.DueDate,
        Estimated:   action.Estimated,
        Tags:        action.Tags,
    }
    api.CreateTask(task)
}
```

### With Calendar Integration

```go
import "google.golang.org/api/calendar/v3"

// Extract actions with deadlines
for _, action := range actionList.Actions {
    if action.DueDate != "" {
        event := &calendar.Event{
            Summary:     action.Description,
            Description: fmt.Sprintf("Priority: %s\nCategory: %s",
                                    action.Priority, action.Category),
            Start: &calendar.EventDateTime{
                Date: action.DueDate,
            },
        }
        calendarService.Events.Insert("primary", event).Do()
    }
}
```

### Email Processing Pipeline

```go
// Process incoming emails
emails := fetchUnreadEmails()

for _, email := range emails {
    // Extract actions from email body
    actionList, _, err := agent.GenerateStructuredData([]messages.Message{
        {
            Role:    roles.User,
            Content: "Extract actions from: " + email.Body,
        },
    })

    if err != nil || len(actionList.Actions) == 0 {
        continue // No actions found
    }

    // Filter high-priority actions
    for _, action := range actionList.Actions {
        if action.Priority == "high" {
            sendNotification("High priority action detected!", action.Description)
        }
    }

    // Auto-create tasks
    createTasksFromActions(actionList.Actions)
}
```

### Batch Processing Multiple Documents

```go
documents := []string{
    readFile("meeting-notes.txt"),
    readFile("project-plan.md"),
    readFile("sprint-review.txt"),
}

allActions := []Action{}

for _, doc := range documents {
    actionList, _, _ := agent.GenerateStructuredData([]messages.Message{
        {Role: roles.User, Content: "Extract actions: " + doc},
    })
    allActions = append(allActions, actionList.Actions...)
}

// Generate consolidated report
generateActionReport(allActions)
```

## Priority Guidelines

| Priority | Use When | Examples |
|----------|----------|----------|
| **High** | Urgent, time-sensitive, blocking others | Hotfix deployment, client emergency |
| **Medium** | Important but not urgent | Code review, documentation update |
| **Low** | Nice to have, can be postponed | Cleanup tasks, optional improvements |

## Category Suggestions

- **work**: Professional tasks, projects, deliverables
- **personal**: Personal errands, self-care, hobbies
- **shopping**: Groceries, supplies, purchases
- **health**: Medical appointments, exercise, wellness
- **administrative**: Paperwork, bills, bureaucracy
- **meeting**: Scheduling, attending meetings
- **communication**: Emails, calls, follow-ups
- **learning**: Study, courses, research

## Customization

### Add Custom Fields

```go
type Action struct {
    ID          int
    Description string
    Priority    string
    Category    string

    // Custom fields
    AssignedTo  string   `json:"assigned_to"`
    Status      string   `json:"status"`      // todo, in-progress, done
    Dependencies []int   `json:"dependencies"` // IDs of dependent actions
    Location    string   `json:"location"`
    Recurring   bool     `json:"recurring"`
    RecurPattern string  `json:"recur_pattern"` // daily, weekly, monthly
}
```

### Filter by Priority

```go
func getHighPriorityActions(actionList ActionList) []Action {
    var highPriority []Action
    for _, action := range actionList.Actions {
        if strings.ToLower(action.Priority) == "high" {
            highPriority = append(highPriority, action)
        }
    }
    return highPriority
}
```

### Export to Different Formats

```go
// Export to Markdown
func exportToMarkdown(actionList ActionList) string {
    var md strings.Builder
    md.WriteString("# Action Items\n\n")

    for _, action := range actionList.Actions {
        md.WriteString(fmt.Sprintf("- [ ] **%s**\n", action.Description))
        md.WriteString(fmt.Sprintf("  - Priority: %s\n", action.Priority))
        md.WriteString(fmt.Sprintf("  - Category: %s\n", action.Category))
        if action.DueDate != "" {
            md.WriteString(fmt.Sprintf("  - Due: %s\n", action.DueDate))
        }
        md.WriteString("\n")
    }

    return md.String()
}

// Export to CSV
func exportToCSV(actionList ActionList) string {
    var csv strings.Builder
    csv.WriteString("ID,Description,Priority,Category,DueDate,Estimated\n")

    for _, action := range actionList.Actions {
        csv.WriteString(fmt.Sprintf("%d,\"%s\",%s,%s,%s,%s\n",
            action.ID, action.Description, action.Priority,
            action.Category, action.DueDate, action.Estimated))
    }

    return csv.String()
}
```

## Performance

- **Average latency**: 1-3 seconds per extraction
- **Accuracy**: High with temperature=0.0
- **Scalability**: Can process multiple documents in batch
- **Model**: Works well with 1.5B+ parameter models

## Related Examples

- **structured/structured-output**: Basic structured extraction (sample 23)
- **structured/structured-validation**: With validation and retry
- See [CLAUDE.md](../../CLAUDE.md) for all snippets

## Reference

- Snippet: `.claude/skills/nova-agent-builder/snippets/structured/structured-output.md`
- Category: `structured`
- Complexity: `intermediate`
- Pattern: Structured data extraction with custom schemas
