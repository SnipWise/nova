package agents

import "github.com/snipwise/nova/nova-sdk/messages"

type Plan struct {
	Tasks []Task `json:"tasks"`
}

type Task struct {
	ID                    string `json:"id"`
	Description           string `json:"description"`
	Responsible           string `json:"responsible"`
	//SubTasks              []Task `json:"sub_tasks"`
	//AdditionalInformation string `json:"additional_information,omitempty"`
}

// Task represents a single task in a plan with optional subtasks.
// The Task type is exported to allow external packages to work with task structures.
// Each task has a unique hierarchical ID (e.g., "1", "1.1", "1.2") for organization.

type TasksAgent interface {
	// Work in progress - not fully defined yet
	IdentifyPlan(userMessages []messages.Message) (plan *Plan, finishReason string, err error)
	IdentifyPlanFromText(text string) (*Plan, error)
}
