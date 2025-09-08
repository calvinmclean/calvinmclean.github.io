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
	github.com/calvinmclean/babyapi v0.31.0
)
```

```go
package main

import _ "github.com/calvinmclean/babyapi"
```

Then I run `go mod tidy`

Then augment with:
> Now implement a custom handler to send an email to a customer by ID. We don't have a way to send email, so just stub it

Using GPT-4.1 free tier.


## No MCP
- This actually worked perfectly on the first try after I started with a `go.mod` and a `main.go` to initialize the dependency
- It fetched the pkg.go.dev documentation which seemed to be sufficient
- It did enable MCP for some reason with full CRUD permissions

I tried this again with Zed so I could see the number of tokens used. This time, it output a plan and did not read the documentation. The plan showed incorrect/halucinate code. I instructed it to read the documentation first, then it made a perfect implementation. It used about 26k tokens.

Trying again with an improved base prompt:
> Use github.com/calvinmclean/babyapi to create a REST API for customer relationship management. This should support babyapi's feature for end-dating resources. It should also use babyapi's CLI and support saving to JSON file. Start by reading the documentation.

- This time, I had to explicitly tell it to implement after it read the docs. Then, it had a slightly broken implementation due to Bind's method signature.


## Gopls MCP
- Interestingly, this is totally flopping... It's making a lot of requests to the MCP, but isn't doing a good job at using the information correctly
- Tried again using Zed which required some hand-holding and used 35k tokens

I tried again in Zed with a modified prompt:
> Use github.com/calvinmclean/babyapi to create a REST API for customer relationship management. This should support babyapi's feature for end-dating resources. It should also use babyapi's CLI and support saving to JSON file. Start by reading the documentation, then use Gopls MCP for more assistance if needed.

I felt it was unfair that I retried the other one with instruction to use the documentation. This time, it used 31k tokens and had a slightly better implementation which used CreatedAt and UpdatedAt fields correctly.

I noticed it mostly was searching by regex, so I am trying again with a new prompt:
> Use github.com/calvinmclean/babyapi to create a REST API for customer relationship management. This should support babyapi's feature for end-dating resources. It should also use babyapi's CLI and support saving to JSON file. Start by reading the documentation, then use Gopls MCP for more assistance if needed. Instead of searching files by regex, use gopls MCP.

This time, it still did a lot of searching and didn't have a very good solution. 36k tokens.

> Use github.com/calvinmclean/babyapi to create a REST API for customer relationship management. This should support babyapi's feature for end-dating resources. It should also use babyapi's CLI and support saving to JSON file. Start by using Gopls mcp to describe the babyapi package. Instead of searching files by regex, use gopls MCP.

The application froze and MCP is not working after restart. It says it's not in a go workspace.

> Use github.com/calvinmclean/babyapi to create a REST API for customer relationship management. This should support babyapi's feature for end-dating resources. It should also use babyapi's CLI and support saving to JSON file. Start by using Gopls mcp to describe the babyapi package. Instead of searching files by regex, use gopls MCP's go_search.

- This time, I got a really complete implementation including a file storage implementation of the Storage interface. All of this with only 23k tokens.
- It's interesting to see how this is inconsistent and very dependent on the prompt and how it chooses to being the project
- It looks like the MCP `go_package_api` function responds with all function signatures and doc comments, similar to my RAG example. When this tool is used early on, it works really well, similar to the documentation. I think getting the pkg.go.dev document is better because it includes some examples of KVStorage, EndDateable, and MCP. This is probably why it insists on using an MCP.
- I think this works better in a large project where there are many smaller packages that can be described, or to use the `go_symbol_references` tool to find uses of a function or type


## Godoc RAG
- This time I will enable an MCP that allows semantic searching a package's comments for usage details
