package messages

import "github.com/snipwise/nova/nova/roles"

type Message struct {
	Role    roles.Role
	Content string
}
