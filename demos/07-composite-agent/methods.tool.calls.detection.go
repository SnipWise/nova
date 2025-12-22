package main

import (
	"github.com/snipwise/nova/nova-sdk/agents/tools"
	"github.com/snipwise/nova/nova-sdk/messages"
	"github.com/snipwise/nova/nova-sdk/messages/roles"
)

func (ca *CompositeAgent) DetectParallelToolCallsWithConfirmation(
	query string,
	toolCallback tools.ToolCallback,
	confirmationCallback tools.ConfirmationCallback,
) (*tools.ToolCallResult, error) {
	return ca.toolsAgent.DetectParallelToolCallsWithConfirmation(
		[]messages.Message{
			{Role: roles.User, Content: query},
		},
		toolCallback,
		confirmationCallback,
	)
}
