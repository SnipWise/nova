package serverbase

import (
	"fmt"
	"strings"

	"github.com/snipwise/nova/nova-sdk/agents"
)

// BuildTaskContext creates a context string from accumulated task results.
func BuildTaskContext(results []string) string {
	if len(results) == 0 {
		return ""
	}
	return strings.Join(results, "\n---\n")
}

// FormatPlanSummary creates a human-readable summary of the plan.
func FormatPlanSummary(plan *agents.Plan) string {
	var sb strings.Builder
	sb.WriteString("**Plan identified:**\n")
	for _, task := range plan.Tasks {
		dependsOn := ""
		if len(task.DependsOn) > 0 {
			dependsOn = fmt.Sprintf(" (depends on: %s)", strings.Join(task.DependsOn, ", "))
		}
		fmt.Fprintf(&sb, "- **%s.** [%s] %s%s\n", task.ID, task.Responsible, task.Description, dependsOn)
	}
	return sb.String()
}
