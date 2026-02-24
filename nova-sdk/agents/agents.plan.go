package agents

import "github.com/snipwise/nova/nova-sdk/messages"

type Plan struct {
	Tasks []Task `json:"tasks"`
}

type Task struct {
	ID          string   `json:"id"`
	Description string   `json:"description"`
	Responsible string   `json:"responsible"`
	ToolName    string   `json:"tool_name,omitempty"`
	Arguments   map[string]string `json:"arguments,omitempty"`
	DependsOn   []string `json:"depends_on,omitempty"`
	Complexity  string   `json:"complexity,omitempty"`
}

type TasksAgent interface {
	// Work in progress - not fully defined yet
	IdentifyPlan(userMessages []messages.Message) (plan *Plan, finishReason string, err error)
	IdentifyPlanFromText(text string) (*Plan, error)
}
