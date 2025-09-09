When working with LLMs, managing the context size is really important. [I previously learned](https://dev.to/calvinmclean/how-to-implement-llm-tool-calling-with-go-and-ollama-237g) that a chat session with an LLM or agent is stored as a continuous array of messages and is re-processed by the model for each prompt. As the conversation or coding sessions goes on, the amount of data the model has to process grows. This eventually erodes the responses since they are processing too much information. It is best to keep sessions short, concise, and on-topic.

When you are using an agent to write code, it constantly has to search directories and read files. A lot of this information is not relevant to the goal and contributes to the context rot. This is something that the experimental Model Context Protocol (MCP) server in `gopls` aims to fix. This tool provides direct access to some parts of the language server protocol (LSP) for Go. Some of these tools, like `go_package_api`, which fetches the exported functions, types, and comments for a Go package, can be really useful to learn how to use a package. Additionally, `go_search` provides a way to search a codebase and its dependencies without reading all of the implementation code. `go_symbol_references` allows the agent to search where a function is used so it can understand the impact of a change without searching a large number of files.

I want to take a deeper dive into `gopls` MCP and how (or if) it can improve coding agents. Specifically, I want to see how it can improve an agent's ability to use external packages where the source code is not readily available. One way to get around this is to use `go mod vendor` to get a local copy of dependencies. However, this is not efficient because the agent really only cares about how to use a package. Reading in all of the implementation source code is a waste of the precious context. An abundance of source code in the context can also cause an agent to start copying pieces of code and style choices from dependencies which is not ideal. If you are using a popular package with tons of open source examples in the LLMs training set, you might not notice this issue. However, it is a huge drawback when trying to use new versions of packages, private internal libraries, or more obscure packages.


## The Experiment

In order to learn more about the `gopls` MCP and see how it improves an agent, I will do an experiment. This is not the most scientific experiment and is largely based on user experience and vibes. The experiment will work by attempting to create a new program using a lesser-known library. This library is `babyapi`, which I wrote to simplify REST API development. This package is pretty unknown and has an eclectic feature set based around things I found interesting or useful.

I will start with a simple project with the module initialized and `github.com/calvinmclean/babyapi` imported.

`go.mod`
```
module customers

go 1.24

require (
	github.com/calvinmclean/babyapi v0.31.0
)
```

`main.go`
```go
package main

import _ "github.com/calvinmclean/babyapi"
```

Here is my initial prompt:
> Use github.com/calvinmclean/babyapi to create a REST API for customer relationship management. This should support babyapi's feature for end-dating resources. It should also use babyapi's CLI and support saving to JSON file.

This prompt doesn't provide any specific function names and generally requests features that I know are provided by `babyapi`.

I am using Github Copilot's free tier to access the GPT-4.1 model.

First, I will run the prompt without the `gopls` MCP server. Then, I will start fresh, [setup the MCP server and system instruction](https://dev.to/calvinmclean/how-to-use-gopls-mcp-with-vs-code-11ha), and run the prompt again.


## Attempt 1: VS Code without gopls MCP

I was surprised to see that the first thing the agent did is fetch the documentation from `pkg.go.dev/github.com/calvinmclean/babyapi`. It then implemented this perfectly on its first attempt.

The end-dating and JSON storage features are directly mentioned and used in examples in my project's README file, so this was easy to implement. Interestingly, the agent also enabled babyapi's built-in MCP server option on the customers API. This feature is also mentioned in the README and probably had a high relevance due to the agent's interest in MCP.

This is the first time I saw an agent fetch documentation like this, so I was impressed to see how well it worked. Unfortunately, it doesn't leave much room for improvement provided by the MCP. The project documentation is somewhat large, so maybe the MCP will be able to reduce the context size by only loading relevant information. Fetching the documentation might not be as effective in larger packages or sparsely-documented ones, so the `gopls` MCP is still worth exploring.


## Attempt 2: VS Code with gopls MCP

Next, I start fresh in a new project, setup the MCP server, the default system instruction, and provide the same prompt. The first attempt here did not work well at all.

It started by using `go_package_api` to describe the `babyapi` package and then implemented the base functionality. Then, it spent a lot of time using the `go_search` tool to figure out how to end-date and save to file storage. It used a variety of search queries like "file storage", "storage", "NewFileDB",  and "NewKVStorage". Each got closer and closer to the correct function which is `kv.NewFileDB`. Then, it started searching for a way to end-date the resources. Eventually, it timed out with a nearly-correct implementation. The only problem was using the incorrect type for the `NewFileDB` config input. Althought I cannot see the context size, it seems like this likely used more tokens due to the long iteration.

The agent with the MCP was at a disadvantage since the KV Storage example is explicitly spelled-out in the package README, but the `go_package_api` only gets documentation from the Go file comments. However, it was interesting to see how the agent was able to iterate using the search tools.

I was curious about how the context usage compares between these two experiments, but VS Code does not show the token usage. For the next attempts, I will switch over to my favorite editor: Zed.


## Attempt 3: Zed without gopls MCP

Zed is able to use the same model provided by Github, so the only difference here will be the agent implementation by Zed. This time, the agent did not start by reading the documentation and came up with a totally halucinated implementation:

```go
func main() {
    api := babyapi.NewAPI()

    // Register the Customer resource with end-dating enabled
    api.RegisterResource(babyapi.ResourceConfig[Customer]{
        Name:      "customer",
        Store:     babyapi.NewJSONStore[Customer]("customers.json"),
        EndDating: true, // Enables end-dating (soft delete)
    })

    // Start the API server (default on :8080)
    api.Run()
}
```

When the documentation is not fetched, it's clearly worse than the agent using `gopls` MCP. Since the agents operated by VS Code and Zed have different system prompts and tooling around them, they tackle this problem differently. The `gopls` MCP will enable general functionality that is independent of the agent being used. This means a lesser, and probably cheaper agent can reach similar functionality to more advanced ones.

In order to estimate the size of the context that comes from reading the documentation, the prompt is modified to improve the agent's behavior:
> Use github.com/calvinmclean/babyapi to create a REST API for customer relationship management. This should support babyapi's feature for end-dating resources. It should also use babyapi's CLI and support saving to JSON file. Start by reading the documentation.

With the imprved prompt, it implemented the program nearly identically to the first attempt with VS Code. This used approximately 26k tokens.


## Attempt 4: Zed with gopls MCP

After struggling a bit to get the `gopls` MCP working with Zed, I continued the next step of the experiment. This had a similar problem compared to VS Code and did a lot of iteration using `go_search`. It eventually used 35k tokens to get a mostly correct solution. It is unfortunate to see that the gopls MCP is not reducing the context size so far.

I spent some more time iterating on the prompt and analyzing the agent logs. Eventually, I came up with this addition to the base prompt:
> Start by using Gopls mcp to describe the babyapi package. Instead of searching files by regex, use gopls MCP's go_search.

Before this change, the agent wasn't always using the `go_package_api` tool, so I instruced it to do so. It was using regex searches on the codebase when trying to find `babyapi` functions which was not working. The `go_search` tool will also search dependencies, so it should use that instead. With the new prompt, it provided a working solution with only 23k tokens. Interestingly, instead of using babyapi's built-in KVStorage, the agent fully implemented the `babyapi.Stoage` interface with its own file storage mechanism. Had it used the built-in solution, it might have used even fewer tokens.


## Conclusion

While the gopls MCP sounds really promising for improving an agent's ability to write Go code, it isn't an absolute game changer right now. While working on these experiments, I learned more about general strategies for using agents rather than learning directly about the impacts of the gopls MCP. Some of these lessons are:
- The prompt is really important. Even with a repository instruction/rule describing how to use the `gopls` MCP, it was useful to explicitly tell the agent to prioritize it over built-in tools
- When you know a specific function that you want to use, it's best to spell it out directly instead of describing it generally. With the addition of the `gopls` MCP, this strategy can become even more effective since the agent can effeciently search for this symbol to learn how to use it
- Agents can consume a lot of context. Instructing it to simply read documentation can be really effective if the documentation is well-written. This of course depends on your goal. If you are looking to use one specific function, it might be overkill to consume the whole documentation
- Agents are best at iterating. Rather than attempting to get a perfect one-shot prompt, pay attention to the agent's output. Interrupt it and correct it early in the process if you see something incorrect

In regards to the `gopls` MCP, my example usecase is probably not the most useful one. As demonstrated, simply reading the documentation is sufficient for smaller, well-documented libraries. LLMs are trained on existing code and like to copy, so reading examples is more effective than analyzing function signatures.

I expect that the `gopls` MCP will work better to augment development in large codebases where the agent can benefit from the `go_search` and `go_symbol_references` tools. Also, in a more extended and iterative development example, loading up the whole documentation might pollute the context too much, especially when using multiple packages. I plan to continue using it in my real development flow to observe how it works in a larger real-world usecase. While it didn't crush this example, it is still promising.

Another interesting strategy for working with external libraries is to use a multi-agent workflow. A second agent can be used to browse a variety of relevant documentation. Then, it will extract or summarize the most important pieces and provide that smaller set of data to the main agent.

Ultimately, I learned that good prompting is still the best way to get good results from an agent. The `gopls` MCP is another tool that, when used properly, can improve the workflow, but is not a perfect solution. It can be used along with good prompts to guide an agent to the best solution.


## Next Steps

Before I heard about the `gopls` MCP, I started theorizing about and working on an idea to improve an agent's ability to get relevant context about a library. This idea is to use retrieval augmented generation (RAG) by vectorizing and semantic-searching documentation. This will provide documentation that is directly relevant to what the agent is looking for. My initial idea was to use information from `pkg.go.dev`, but I ended up parsing Go files directly to read documentation comments. This was easier to parse and link directly to the code.

After hearing about the gopls MCP, I felt discouraged about this idea and thought that it might no longer be relevant. This experiment proved otherwise. The agent that read documentation was really effective. While reading the whole documentation worked well, using the vectorized data could reduce overall context since the agent won't need to load the whole documentation into each session. Additionally, it should pair really well with the `gopls` MCP. The semantic search will find relevant symbol and file names, and `gopls` can be used to look them up and get more details if needed. I will continue working on this and rerun the experiments to see if there are any benefits. I learned here that I should also parse a library's README since it can be more useful than the code comments. Keep an eye out for my next post where I will share these results!

Thank you for reading! If you learned anything new or have any advice, please let me know in the comments.
