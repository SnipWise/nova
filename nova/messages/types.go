package messages

import "github.com/snipwise/nova/nova/messages/roles"

type Message struct {
	Role    roles.Role
	Content string
}
