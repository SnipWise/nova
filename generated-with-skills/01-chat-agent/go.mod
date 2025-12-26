module chat-agent

go 1.25.4

require github.com/snipwise/nova v0.0.0

require (
	github.com/openai/openai-go/v3 v3.10.0 // indirect
	github.com/tidwall/gjson v1.18.0 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	github.com/tidwall/sjson v1.2.5 // indirect
)

replace github.com/snipwise/nova => ../..
