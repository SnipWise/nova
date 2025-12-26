package agents

import "github.com/snipwise/nova/nova-sdk/messages"

type Intent struct {
	TopicDiscussion string `json:"topic_discussion"`
}

// OrchestratorAgent is an interface for agents that can identify intents/topics from user input
type OrchestratorAgent interface {
	// IdentifyIntent sends messages and returns the identified intent
	IdentifyIntent(userMessages []messages.Message) (intent *Intent, finishReason string, err error)

	// IdentifyTopicFromText is a convenience method that takes a text string and returns the topic
	IdentifyTopicFromText(text string) (string, error)
}

