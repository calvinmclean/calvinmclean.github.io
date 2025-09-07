# Gopls MCP Demo

This project will have 3 different implementations:
- No Gopls
- Gopls
- Gopls + RAG tool

I will use the same prompt for each and attempt to record the agent's process.

Prompt:
> Use github.com/calvinmclean/babyapi to create a REST API for customer relationship management. This should support babyapi's feature for end-dating resources. It should also use babyapi's CLI and support saving to JSON file.

I should start with a go.mod
```
module customers

go 1.24

require (
	github.com/calvinmclean/babyapi v0.30.0
)
```

```go
package main

import _ "github.com/calvinmclean/babyapi"
```

Then I run `go mod tidy`

Then augment with:
> Now implement a custom handler to send an email to a customer by ID. We don't have a way to send email, so just stub it

## No MCP
- This actually worked perfectly on the first try after I started with a `go.mod` and a `main.go` to initialize the dependency
- It fetched the pkg.go.dev documentation which seemed to be sufficient
- It did enable MCP for some reason with full CRUD permissions

## Gopls MCP
- Interestingly, this is totally flopping... It's making a lot of requests to the MCP, but isn't doing a good job at using the information correctly
- Tried again using Zed which required some hand-holding and used 35k tokens

## Godoc RAG
- This time I will enable an MCP that allows semantic searching a package's comments for usage details
