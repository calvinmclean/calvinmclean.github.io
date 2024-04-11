<img alt="sorry darkmode users" src="https://raw.githubusercontent.com/calvinmclean/calvinmclean.github.io/main/articles/babyapi-htmx/ui-demo.gif"/>

> This UI and backend are implemented with only 150 lines of code, including the HTML!
> The full example code for this tutorial is available in the [`babyapi` GitHub repository](https://github.com/calvinmclean/babyapi/blob/main/examples/todo-htmx/main.go) if you're eager to get into it.

In my recent article, [The Easiest Way to Create a REST API With Go](https://dev.to/calvinmclean/the-easiest-way-to-create-a-rest-api-with-go-20bo), I demonstrated how [`babyapi`](https://github.com/calvinmclean/babyapi) can jumpstart REST API creation. This time, I will walk you through some additional `babyapi` features and show how to create an easy and dynamic frontend using [HTMX](https://htmx.org). If you're not already familiar, HTMX is a library that basically extends HTML with functionality that normally requires Javascript.

HTMX was designed with REST backends in mind, which makes it a perfect companion for `babyapi`. HTMX provides a snappy SPA-like feel by replacing and rendering individual components from server responses instead of the entire page.

Since `babyapi` uses `chi/render` package for requests and responses, it automatically supports input from HTML forms instead of only JSON. Additionally, `babyapi` defines an `HTMLer` interface which allows resources to define custom HTML string responses when the client requests `text/html`. We can take advantage of these features to easily create and serve an HTMX frontend.


## Create the TODOs API

In the last article, we created a simple TODO resource. This struct allows `babyapi` to serve the necessary HTTP endpoints for interacting with the data:

```go
package main

import "github.com/calvinmclean/babyapi"

type TODO struct {
    babyapi.DefaultResource

    Title       string
    Description string
    Completed   bool
}

func main() {
    api := babyapi.NewAPI[*TODO](
        "TODOs", "/todos",
        func() *TODO { return &TODO{} },
    )

    api.RunCLI()
}
```

In order to add an HTMX UI on top of this, all we need to do is:
  1. Implement `babyapi.HTMLer` interface for the `TODO` resource
  2. Use `api.SetGetAllResponseWrapper` with a new `babyapi.HTMLer` type to render HTML for the `/todos` response
  3. Write templates for HTMX frontend with these features:
    - List all TODOs in a table
    - Buttons to mark items as completed and delete
    - Server-sent events automatically append new TODOs
    - Form to create TODOs


## Respond With HTML Instead of JSON

By default, `babyapi` is designed to marshal structs to the requested response type (usually JSON). HTML responses are not as straightforward, so `babyapi` provides an `HTMLer` interface that enables creating HTML responses from resource structs:

```go
type HTMLer interface {
	HTML(*http.Request) string
}
```

The `HTML(*http.Request)` method can be implemented as follows:

```go
const todoRowTemplate = `...`

func (t *TODO) HTML(*http.Request) string {
	tmpl := template.Must(
        template.New("todoRow").Parse(todoRowTemplate),
    )
	return babyapi.MustRenderHTML(tmpl, t)
}
```

This method simply renders a template from a string and uses `babyapi.MustRenderHTML` to execute it with the TODO data.

The HTML template looks like this:
```html
<tr hx-target="this" hx-swap="outerHTML">
	<td>{{ .Title }}</td>
	<td>{{ .Description }}</td>
	<td>
		{{- $disabled := "" }}
		{{- if .Completed }}
			{{- $disabled = "disabled" }}
		{{- end -}}

		<button
			hx-put="/todos/{{ .ID }}"
			hx-headers='{"Accept": "text/html"}'
			hx-include="this"
			{{ $disabled }}>

            <!-- Include entire TODO item for idempotent PUT -->
			<input type="hidden" name="Completed" value="true">
			<input type="hidden" name="Title" value="{{ .Title }}">
			<input type="hidden" name="Description" value="{{ .Description }}">
			<input type="hidden" name="ID" value="{{ .ID }}">
			Complete
		</button>

		<button hx-delete="/todos/{{ .ID }}" hx-swap="swap:1s">
			Delete
		</button>
	</td>
</tr>
```

> NOTE: some minor details, like CSS classes, are excluded from the example for simplicity. The full example can be found [on GitHub](https://github.com/calvinmclean/babyapi/blob/main/examples/todo-htmx/main.go)

This template creates an HTML table row (`<tr>`) to display the TODO's title and description. The row also contains buttons to mark the item as complete or to delete it.

On top of the regular HTML, we use HTMX attributes to control interactions with the backend:
  - `hx-target` and `hx-swap` tell HTMX to replace the entire row with the contents of a successful response. These attributes are inherited by the buttons
  - The "Complete" button uses `hx-put` to make a `PUT` request to `/todos/{{ .ID }}`. The request sets `Accept: text/html` to request an HTML response from the server. `hx-include="this"` uses the child `input` fields to make the request body. This sets `Completed=true` and leaves the other fields unchanged
  - The "Delete" button uses `hx-delete` to send a `DELETE` request to `/todos/{{ .ID }}`. Then, `hx-swap="swap:1s"` tells HTMX to swap new contents over 1 second instead of immediately. Since the response content will be empty, this creates a fade-out effect

Now that we have a simple HTML implementation, run the server, create a TODO item, and fetch it:
```shell
go build -o todo-app

# start the server in a separate terminal
./todo-app serve

# create new TODO
./todo-app post TODOs '{"title": "use babyapi!"}'

# get the previously-created TODO by ID with HTML response
./todo-app -H "Accept: text/html" get TODOs clnvnt5o402av6j1oal0
```

Loading `http://localhost:8080/todos/{ID}` in the web browser will show the browser-rendering of the HTML, but it's not very impressive yet without style or HTMX working. The next step will be rendering the full HTML page to display all TODOs and enable HTMX.


## Create the All TODOs Page

Logically, the UI view for all TODOs will come from our base endpoint, `/todos`. However, `babyapi` is setup to use the default response type of `*babyapi.ResourceList[*TODO]` to create a JSON-compatible response. As you probably guessed, `babyapi` has a way to deal with this.

All we have to do is use `api.SetGetAllResponseWrapper` to set a function that accepts `[]*TODO` and returns a new `render.Renderer`. For this example, we create the `AllTODOs` type to satisfy the `render.Renderer` and `babyapi.HTMLer` interfaces:

```go
type AllTODOs []*TODO

func (at AllTODOs) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

const allTODOsTemplate = `...`

func (at AllTODOs) HTML(*http.Request) string {
	tmpl := template.Must(
        template.New("todoRow").Parse(todoRowTemplate),
    )
	tmpl = template.Must(
        tmpl.New("allTODOs").Parse(allTODOsTemplate),
    )
	return babyapi.MustRenderHTML(tmpl, at)
}

func main() {
	api := babyapi.NewAPI[*TODO]("TODOs", "/todos", func() *TODO { return &TODO{} })

	api.SetGetAllResponseWrapper(
        func(todos []*TODO) render.Renderer {
		    return AllTODOs(todos)
	    },
    )

    // ...
}
```

The `AllTODOs` type implements `babyapi.HTMLer` in the same way as the previous example: just parse and execute HTML templates. Then,  `api.SetGetAllResponseWrapper` function simply returns `AllTODOs` from the provided slice of TODOs.

The `allTODOs` template contains the full HTML page which imports HTMX scripts and UIKit CSS. The `<body>` portion of this new template looks like:

```html
<body>
    <table>
        <thead>
            <tr>
                <th>Title</th>
                <th>Description</th>
                <th></th>
            </tr>
        </thead>

        <tbody>
            <form hx-post="/todos" hx-swap="none" hx-on::after-request="this.reset()">
                <td>
                    <input name="Title" type="text">
                </td>
                <td>
                    <input name="Description" type="text">
                </td>
                <td>
                    <button type="submit">Add TODO</button>
                </td>
            </form>

            {{ range .Items }}
            {{ template "todoRow" . }}
            {{ end }}
        </tbody>
    </table>
</body>
```

> NOTE: some styling and other details that are not important to the example are excluded for brevity. The full example can be found [on GitHub](https://github.com/calvinmclean/babyapi/blob/main/examples/todo-htmx/main.go)

This template creates the structure of an HTML table and populates the first row with a simple HTML form for creating new TODO items. Then, the `todoRow` template from the previous section is used to create rows for all existing TODOs.

The form uses `hx-post="/todos"` to send a `POST` request with the form contents to the API. `hx-swap="none"` disables swapping the response contents since we will use server-sent events in the next section to append new rows. Then we use `hx-on::after-request="this.reset()"` to reset the form.


## Implement Server-Sent Events

Server-sent events allow one-way communication from the backend to the frontend. In this case, we will use the feature to push new TODO rows to the UI, even if they are created from the CLI or other sources. Luckily, both `babyapi` and HTMX make it super easy to use server-sent events.

HTMX has an [SSE extension](https://htmx.org/extensions/server-sent-events/) which can be used in this UI by replacing the plain `<tbody>` with the following:

```html
<tbody
    hx-ext="sse"
    sse-connect="/todos/listen"
    sse-swap="newTODO"
    hx-swap="beforeend">
    ...
</tbody>
```

This new `tbody` works with server-sent events using the new HTMX attributes:
  - `hx-ext="sse"` enables the extension
  - `sse-connect="/todos/listen"` opens the connection to the server's SSE endpoint (not implemented yet)
  - `sse-swap="newTODO"` instructs HTMX to only use events with the type `newTODO`
  - `hx-swap="beforeend"` will append rows before the `</tbody>` closing tag

Now that the frontend is ready to receive new TODO rows, the backend just needs to send them! `babyapi` makes it unbelievably simple to handle SSE connections, even if this concept is completely new to you. The `api.AddServerSentEventHandler` modifier adds a new route to handle SSE connections and returns a channel for sending events.

Since we want to push new TODOs after they are created, we can use `api.SetOnCreateOrUpdate` to push events to the channel after successful `POST` requests:

```go
todoChan := api.AddServerSentEventHandler("/listen")

api.SetOnCreateOrUpdate(func(r *http.Request, t *TODO) *babyapi.ErrResponse {
    if r.Method != http.MethodPost {
        return nil
    }

    select {
    case todoChan <- &babyapi.ServerSentEvent{
        Event: "newTODO",
        Data: t.HTML(r),
    }:
    default:
    }
    return nil
})
```

This function first makes sure the request method is `POST` and then pushes a new event on the channel. The event has the `newTODO` type that is expected by the frontend and uses the TODO's `HTML` method to create the row. Additionally, it uses `select` with a `default` case instead of simply pushing to the channel so the function will not block if there is no frontend receiving from the channel.


## Final Touches

With the main structure and functionality in place, only a few minor tweaks are required to finish the seamless integration between `babyapi` and HTMX.

By default, `babyapi` is setup to respond to successful `DELETE` requests with `204 No Content`, but HTMX treats this response as the server requesting to [disable the swap](https://htmx.org/docs/#requests). This means our deleted rows will not disappear until the page is refreshed, which is not okay in a modern web application. In order to enable swaps, HTMX needs a `200 OK` response. Once again, `babyapi` provides the necessary flexibility to achieve this with the `api.SetCustomResponseCode` modifier:

```go
api.SetCustomResponseCode(http.MethodDelete, http.StatusOK)
```

This simply instructs the default `DELETE` handler to respond with `http.StatusOK`.

All that's left is adding a little style. I won't claim to be a frontend engineer or a UI designer, so I just used [UIKit](https://getuikit.com) to easily add modern-looking style to the HTML table and buttons. As mentioned throughout the article, the CSS classes and other small details are excluded since they are not directly relevant to the tutorial. See the full example [on GitHub](https://github.com/calvinmclean/babyapi/blob/main/examples/todo-htmx/main.go) to try running it for yourself.


## Storage Layer

At this point, we have a fully functioning web app for managing TODOs. If you want to take it one step further and use it for a real TODO tracker instead of a demo, you just need to add persistent storage. `babyapi` also simplifies this with the `babyapi/storage` package. This package provides a generic implementation of the `Storage` interface with helpers for setting up local file or Redis storage. Add the following to the `main` function to save the TODOs in a JSON file:

```go
db, err := storage.NewFileDB(hashmap.Config{
    Filename: "todos.json",
})
if err != nil {
    panic(err)
}

api.SetStorage(storage.NewClient[*TODO](db, "TODO"))
```


## Conclusion

In this tutorial, we extended the super simple `babyapi` introduction example to implement an HTMX frontend without adding too much complexity. This shows how `babyapi`'s RESTful API and HTMX, a frontend library designed for RESTful backends, are a perfect fit together. It's easier than ever to create a responsive and dynamic web application.

The steps here also demonstrate the flexibility of `babyapi`. Although the default API provided is sufficient for all CRUD functionality, our real-world usecases often introduce more requirements and variance which `babyapi` is able to handle gracefully. Hopefully this has revealed how `babyapi` can be used to implement your next API-driven application!

I encourage you to continue experimenting with `babyapi` and HTMX to learn more! Here are some ideas to get you started with extending this example:
  - Add a toggle to show complete, incomplete, or all TODOs
  - Allow "un-completing" a TODO
  - Add a `CreatedAt` field to the `TODO` struct and sort by this date (you can implement `render.Binder` for `TODO` to automatically set `CreatedAt` on new items)

Thanks for reading!
