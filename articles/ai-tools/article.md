Tool-calling is a feature of LLMs and LLM APIs that enables the model to access external information and perform actions. Instead of responding to a prompt directly, the model generates inputs to a function, calls that function (more on this later), and then generates a response based on the function's output. This significantly improves the functionality of LLMs by accessing real-time information like weather, sports scores, or recent events. Additionally, the model can interact with external systems like databases, APIs, or even execute scripts.

LLMs and tool-calling introduced a new type of program where the operation and flow is controlled by the LLM rather than rigid code. This creates highly flexible and dynamic programs, but they can also be less reliable and consistent. Like always, it's important to balance functionality and reliability.


## How it works

LLMs are only able to generate text. In order to use tools, the model needs to be prompted _through_ another program or controller. Instead of actually "calling" a function directly, the the model generates structured data to describe the function to use and which inputs to use with it. The controller program is responsible for parsing the input data, finding the function, and executing it. Then, it provides the output to the LLM (often as JSON). The LLM is finally able to respond to the user's prompt using the output of the function in addition to the data it was trained on.

Before using tools, the LLM needs to know about the tools that it can use. It needs the function's name, a description of what it does, and a schema defining the function's inputs. This can be done with system prompts, but many model APIs now include a separate input field for tool definitions.

Interacting with the model directly, with only a thin chat abstraction around it, effectively demonstrates how this works. Ollama allows creating a Modelfile to extend a model with a specific system prompt or other configurations. Here is a system prompt that instructs a model how to use "tools" (but really it will just be me responding with JSON):
```
FROM gemma3:4b

SYSTEM """
A program is used to interact with you and the user. The program allows you to use functions/tools
to get external information. When you use a tool, the program will respond with JSON data. Anything else is
coming from the user and should be treated as human interaction.

To use a tool, respond with a JSON object with the following structure:
{
	"tool": <name of the called tool>,
	"tool_input": <parameters for the tool matching the above JSON schema>
}
Do not include any other text in your response.

If the user's prompt requires a tool to get a valid answer, use the above format to use the tool.
After receiving a tool response, continue to answer the user's prompt using the tool's response.
If you don't have a relevant tool for the prompt, answer it normally. Be fun and interesting.

Here are your available tools:

getCurrentWeather
Get the current weather in a given location
{
	"type": "object",
	"properties": {
		"location": {"type": "string", "description": "The city and state, e.g. San Francisco, CA"},
		"unit": {"type": "string", "enum": ["celsius", "fahrenheit"]}
	},
	"required": ["location", "unit"]
}
"""
```

Now that the model is ready, let's chat:

1. Input:
    > what is the weather like in new york city?

1. Model's response:
    ```json
    {
    	"tool": "getCurrentWeather",
    	"tool_input": {
    		"location": "New York City",
    		"unit": "fahrenheit"
    	}
    }
    ```

1. Input:
    ```json
    {"temperature": 78, "unit": "F"}
    ```

2. Model's response:
    >  The temperature in New York City is currently 78°F! It’s a beautiful, warm day – perfect for a stroll through Central Park or grabbing a hot dog.

Instead of attempting to hallucinate a response, the model generates tool input using the specified format. Then, I respond with JSON emulating the tool output. With this new information, the model is able to accurately answer the prompt with a friendly flair.

Although this isn't real tool calling, it is interesting to see how impactful the system prompt is on the model's behavior. This is helpful to demystify the LLM's behavior and show what is really happening behind the scenes. This foundational understanding beneficial when developing more complex tools and building applications around LLMs.


## Implement Tool-calling with Ollama + Go

The next step is to integrate tool calling into a real program instead of chatting with the model. In order to stay at a low-level and learn more about LLM interactions, the program uses Ollama's API and Go client instead of a more generalized library.

Here is the request used to prompt a model with a tool defined:
```go
messages := []api.Message{
	api.Message{
		Role:    "user",
		Content: "What is the weather like in New York City?",
	},
}

req := &api.ChatRequest{
		Model: model,
		Messages: messages,
		Tools: api.Tools{
			api.Tool{
				Type: "function",
				Function: api.ToolFunction{
					Name:        "getCurrentWeather",
					Description: "Get the current weather in a given location",
					Parameters: api.ToolFunctionParameters{
						Type:     "object",
						Required: []string{"location"},
						Properties: map[string]api.ToolFunctionProperty{
							"location": {
								Type:        api.PropertyType{"string"},
								Description: "The city and state, e.g. San Francisco, CA",
							},
						},
					},
				},
			},
		},
	}
```

The client program sends the ChatRequest and handles the ToolCall in the model's response:
```go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/ollama/ollama/api"
)

const model = "llama3.2:3b"

func main() {
	client, err := api.ClientFromEnvironment()
	if err != nil {
		log.Fatal(err)
	}

	messages := []api.Message{
		api.Message{
			Role:    "user",
			Content: "What is the weather like in New York City?",
		},
	}

	ctx := context.Background()
	req := &api.ChatRequest{
		// ... request from above
	}

	handler := func(resp api.ChatResponse) error {
		// If there is no tool call, just print the message content
		if len(resp.Message.ToolCalls) == 0 {
			fmt.Print(resp.Message.Content)
			return nil
		}

		// Otherwise, process the tool call
		tc := resp.Message.ToolCalls[0].Function
		switch tc.Name {
		case "getCurrentWeather":
			output, err := getCurrentWeather(tc.Arguments)
			if err != nil {
				log.Fatal(err)
			}

			messages = append(messages, api.Message{
				Role:    "tool",
				Content: output,
			})
		default:
			log.Fatal(fmt.Errorf("invalid function: %q", tc.Name))
		}

		return nil
	}

	// Send initial chat message
	err = client.Chat(ctx, req, handler)
	if err != nil {
		log.Fatal(err)
	}

	// The model should have responded with a tool call
	// and the handler would have appended a response to messages.
	// Now, call the tool again with the response
	req.Messages = messages
	err = client.Chat(ctx, req, handler)
	if err != nil {
		log.Fatal(err)
	}
}

func getCurrentWeather(input map[string]any) (string, error) {
	location, ok := input["location"].(string)
	if !ok {
		log.Fatalf("bad args: %v", input)
	}

	weatherInfo := map[string]any{
		"location":    location,
		"temperature": "80",
		"unit":        "fahrenheit",
		"forecast":    []string{"sunny", "windy"},
	}

	b, err := json.Marshal(weatherInfo)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
```

The output should be something like:

> The current weather in New York City is mostly sunny, with a windy condition. The temperature is at 80°F.


## Going a step further

I happen to have a program that is perfect for demonstrating simple LLM integration using tools. The [`walks-of-italy`](https://github.com/calvinmclean/walks-of-italy) was originally created to watch for new tour availability from [Walks of Italy](https://www.walksofitaly.com) tour company. I started with a grander vision for this program: a tool that could watch, search, and observe tour availability. After implementing the base feature of notifying on new availability, I decided it was good enough for my use-case.

Now, with LLM tools, this program can be extended in any number of ways. Here are the tools it currently has:
- `getAllTours`: list the tour names and their UUIDs from the database
- `getTourDetails`: use the API to access the tour descriptions
- `getTourAvailability`: use the API to check availability, prices, and capacity at certain times

With just a few tools, the LLM has these features:
- Simply query a date range with natural language
  - "what is the availability of the Pristine Sistine tour next week?"
- Ask a more general question about multiple tours
  - "what is the availability of vatican tours on July 5th?"
- Get travel recommendations
  - "I am going to be in Rome on August 3rd and 4th. Which tours can I go on?"
  - "If I can only do one tour in Rome, which should I do?"

When asked "what is the availability of the Pristine Sistine tour next week?" the model responds:
> Here's an overview of the availability for the tours:
>
> - **June 20, 2025**:
>   - There are limited spots available for the 7:15 AM and 7:45 AM tours.
>   - The 7:15 AM tour has 1 vacancy out of a capacity of 20.
>   - The 7:45 AM tour also has 1 vacancy out of a capacity of 48.
>   - The 6:45 AM tour is fully booked.
>
> - **June 21, 2025**:
>   - The 6:45 AM tour has 10 available spots out of 10 total.
>   - The 7:15 AM tour has 3 vacancies out of a capacity of 34.
>   - The 7:45 AM tour has 1 vacancy out of a capacity of 28.
>
> - **Other dates**:
>   - On June 19, 2025, the 7:45 AM tour is fully booked.
>   - On June 18, 2025, the 7:15 AM tour is fully booked, and the 7:45 AM tour has limited availability.
>   - On June 17, 2025, the 7:15 AM tour is fully booked, and the 7:45 AM tour has limited availability.
>   - On June 16, 2025, the 7:45 AM tour is fully booked.

If you're looking for availability, the best options are on June 20 and June 21, with limited spots available for the morning tours. Let me know if you'd like help booking!


These are just a few examples that I can think of. A huge benefit to this type of program is that its functionality isn't only limited by the features added by a developer. Instead, it's limited by the LLM and the tools that it can use. The user might come up with a usecase that works out of the box due to the program's flexibility. In this example, I justed wanted to check specific tours and dates for availability, but can now do a lot more.

The effectiveness and accuracy of the responses here of course varies and is dependent on the model that is used. More complex questions that require more tool calls will generally decrease the accuracy of responses. When a lot of information is required, larger and more capable models are required.

In this example, the functionality could be made even more flexible. Currently, it just is able to use tours that I manually add. If it has a new function to get tours from the Walks of Italy website, it could be even more dynamic (but less predictable).


## MCP

Model Context Protocol, or MCP, is the next evolution of tool calling. Instead of requiring a custom program to provide tools, it moves this to the server-side. The model is used with a generic client that supports the protocol and accesses functions provided by any number of external servers. If this previous example used MCP, Walks of Italy would provide an MCP server any model would be able to access it when used with an MCP client. It would eliminate almost all of the work done to implement [my `walks-of-italy` program](https://github.com/calvinmclean/walks-of-italy).


## Thoughts

Until recently, I primarily interacted with LLMs using ChatGPT or Gemini to answer questions, discuss ideas, and sometimes generate SQL queries or code. As AI Agents become more popular and integrated into all parts of our lives and jobs, it starts to seem like LLMs can do anything. It was refreshing to start from the barebones model interactions and really see how tool-calling works. In reality, LLMs are still just reading input and generating output. It's the code around them, built by software engineers, that enables them to actually _do_ things.

Building an application around an LLM and tools can be really beneficial. If the model has access to a set of tools with general functionality like querying APIs and DBs (with pre-defined SQL of course), the program's functionality can change and expand on-demand. This is especially useful for non-technical users. Instead of asking engineers to build new features for a specific usecase, they can just adjust their prompts and interactions with the LLM. In the previous example, the `walks-of-italy` program was originally designed to watch and alert for new tour availability or run specific date range queries with the CLI. The introduction of a few tools and an LLM drastically expanded the scope of the program and makes it more useful for non-technical users. Now that the LLM can query any date range and get descriptions about tours, you can:
  - Get help deciding which tours to go based on real pricing, descriptions, and availability
  - Ask about general availability next month for multiple tours
  - Compare availability and prices for weekdays or weekends

It is important to remember that the model can behave unpredictably and the tools should not provide unlimited access. For example, instead of providing a tool to execute any generated SQL, engineers should expose a few functions like "createResource" and "updateResource" with pre-defined queries.

System prompts and tool definitions provide more structure and control over how an LLM operates. An engineer can build a relatively reliable and useful program with these techniques. However, as complexity of the system grows, the potential for the LLM to do unpredictable things also grows.

Normally, code is very predictable. A set of functions exists and engineers define exactly when and how they are used. Now, programs become a collection of tools and the actual control is being relinquished to LLMs. Engineers always have to balance complexity, maintainability, and functionality. Using LLMs just makes this balance more delicate.
