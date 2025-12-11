package roles

type Role string

const (
	System    Role = "system"
	User      Role = "user"
	Assistant Role = "assistant"
	Developer Role = "developer"
	Tool	  Role = "tool"
)