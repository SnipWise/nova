package messages

import "github.com/snipwise/nova/nova-sdk/messages/roles"

type Message struct {
	Role    roles.Role
	Content string
}
