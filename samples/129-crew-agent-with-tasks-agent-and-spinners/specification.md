SPECIFICATION:
I need a simple golang http server with:
- the root endpoint that serves a home.html page
- a `/hello-text` endpoint that returns a "hello world" response in text mode
- a `/hello-json` endpoint that returns a "hello world" response in json mode
- a `/hello` endpoint (POST method) with json payload (name is the argument) that returns in json mode a greeting to `name`
- use only base packages from the Go standard library
- the server should listen on port 8080
- add comments to explain the code