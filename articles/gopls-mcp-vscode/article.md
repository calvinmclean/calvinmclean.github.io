The Go team recently added an official [Model Context Protocol (MCP) server to `gopls`](https://tip.golang.org/gopls/features/mcp), the language server protocol (LSP) implementation for Go. This allows IDEs and other AI coding agents to directly access some features of the LSP. While most of the features provided by the MCP already exist in AI coding tools, this can improve the efficiency of token use by avoiding searching and reading whole files.

### Why MCP?

One of the most impactful features is the ability to use the `go_diagnostics` tool. Besides the usual information about diagnostic issues, this also allows the agent to automatically review and apply fixes suggested by the LSP. This greatly improves consistency and accuracy of code fixes since it uses deterministic analysis provided directly by the Go maintainers.

If you are regularly programming with newer versions of Go, you might be familiar with the "modernize" suggestions. I mostly-commonly see these with typical `for i := 0; i < 10; i++` loops that could use the new `for i := range 10` syntax. This feature is a big win for the language in the world of AI coding. By definition, LLMs always output "old" code since they are trained on existing codebases. This means any new language features won't be used by LLMs until enough human coders have introduced them into the training data.

The new [analyze and modernize](https://pkg.go.dev/golang.org/x/tools/gopls/internal/analysis/modernize) tools allow coding agents to automatically update to new language features that they don't even know about yet. This improves the overall quality of a codebase even when using AI agents. This is an important feature helping to keep Go relevant as software development is rapidly changing.


## Enable gopls MCP in VS Code

Now, let's get to the reason you're here! With a few simple steps, we can start using this new MCP feature in VS Code:

### 1. Install/update `gopls`
```shell
go install golang.org/x/tools/gopls@latest
```
- I currently have Go `1.24.0` with `v0.20.0` of `gopls`, but newer versions exist at the time of writing this

### 2. Enable VS Code to run the server
VS Code automatically runs the `gopls` server for Go projects. A simple update to the config will also start the MCP server.

- Navigate to Settings and find `Go: Language Server Flags`
- Click `Edit in settings.json` and add:
  ```json
  "go.languageServerFlags": [
      "-mcp.listen=localhost:8092",
  ]
  ```
- Restart/reload VS Code

### 3. Enable the MCP server for agents
- Use the `MCP: Add Server...` command
- Select `HTTP`
- Input the address: `http://localhost:8092`
- Alternatively, just edit `.vscode/settings.json` with this content:
  ```json
  {
	  "servers": {
  		"gopls": {
	  		"url": "http://localhost:8092",
		  	"type": "http"
      }
	  }
  }
  ```

### 4. Setup instructions for using the tools

The Go team provides a base instruction to improve how agents use the feature. Setting up VS Code to use this feature will greatly improve how it uses the MCP tools.

- First, create the instructions directory
  ```shell
  mkdir -p.github/instructions
  ```
- Then, start by telling VS Code to use this for Go files:
  ```shell
  cat <<EOF > .github/instructions/gopls.instructions.md
  ---
  applyTo: "**/*.go"
  ---
  EOF
  ```
- Finally, add the `gopls` instruction:
  ```shell
  gopls mcp -instructions >> .github/instructions/gopls.instructions.md
  ```

## Use it!

Now you should have `gopls` MCP setup in your Go project! Open up the agent chat and start coding!

The simplest place to start is to ask about references for a function or type in your code: "Describe where MyFunction is used."

This should use the `go_symbol_references` tool to find references to this function.

A more interesting example uses the `go_diagnostics` tool that I mentioned earlier. To do this, write a simple function using a for loop:

```go
func Example() {
	for i := 0; i < 10; i++ {
		println("Hello, World!")
	}
}
```

Then, ask the agent to run diagnostics on it! It should look something like this:
![Go Diagnostic Agent Example](https://raw.githubusercontent.com/calvinmclean/calvinmclean.github.io/main/articles/gopls-mcp-vscode/go_diagnostics_example.png "Go Diagnostic Agent Example")
