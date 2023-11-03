### Introduction

GNU Make originated as a dependency-tracking build tool that automates the build process of source files and libraries. Although many modern languages provide built-in dependency management and build tools, a Makefile still finds its place in these projects. In these cases, Makefiles are used as a collection of aliases for actions like test, build, and lint. In this article, I will show you that there is a better option for these use-cases: [`task`](https://taskfile.dev).

### But Make is standard practice!

Yes, Make has been continuously developed over nearly 50 years, leading to a robust set of features and a rich history in software engineering. It will likely be around forever, but these deep roots are the same thing preventing it from being user-friendly and intuitive.

- Makefiles often become a mess with lots of environment variables, macros, and special symbols
- You MUST use tabs (ugh)
- You can't pass arguments without using environment variables
- An annoying convention in Makefiles is to string together a lot of environment variables to construct commands
- `.PHONY` is required when using non-file targets as command aliases
- Many features are designed around building files, which often isn't relevant in these modern scenarios

When we aren't using it for the core features Make is built and designed around, we have to ask ourselves if there is an alternative.

### Introduction to Task

This is where the Taskfile comes in. The YAML format creates a self-documenting file that states how each "task" will behave. 
The verbose format is a welcome feature in comparison to the Makefile and will result in a more approachable file. When integrated in an open-source project, it can make it easier for new contributors to get started.

After following the [simple install process](https://taskfile.dev/installation/), you just need to run `task --init` to create a simple example file. This example shows how to use environment variables and execute commands.

```yml
version: '3'

vars:
  GREETING: Hello, World!

tasks:
  default:
    cmds:
      - echo "{{.GREETING}}"
    silent: true
```

At this point, you have all the information needed to cover the basic use-cases. Rather than providing a tutorial here, I encourage you to check out the [official documentation](https://taskfile.dev/usage/). The well-written guide starts by showing the basics with concrete examples and graduates into more complex functionality as you scroll. This is a welcome sight in comparison to the large plain HTML page provided by Make.

### Rewriting a Makefile to Taskfile

I recently read the [Guide to using Makefile with Go](https://levelup.gitconnected.com/a-comprehensive-guide-for-using-makefile-in-golang-projects-c89edebcbe6e) by Ramseyjiang. This is an interesting read and makes good points about providing a consistent interface for common tasks and builds. It left me thinking about how the developer experience could be further improved by using a Taskfile instead.

This is the example Makefile created by the article's author:
```Makefile
APP_NAME = myapp
GO_FILES = $(wildcard *.go)
GO_CMD = go
GO_BUILD = $(GO_CMD) build
GO_TEST = $(GO_CMD) test

all: build

executable
build: $(APP_NAME)

$(APP_NAME): $(GO_FILES)
   $(GO_BUILD) -o $(APP_NAME) $(GO_FILES)

test:
   $(GO_TEST) -v ./... -cover

.PHONY: all build test
```

Here is my translation to Taskfile:
```yml
version: "3"

vars:
  APP_NAME: myapp

tasks:
  default:
    cmds:
      - task: build

  build:
    cmds:
      - go build -o {{.APP_NAME}} *.go
    sources:
      - "*.go"
    generates:
      - "{{.APP_NAME}}"

  test:
    cmds:
      - go test -v ./... -cover
```

If you're counting lines, you might notice the Taskfile has a few more. This could be equalized by using `cmd` with a string instead of `cmds` and the array, but my priority here is to create something easy to read and build upon.

### Task Features

Task covers all of the main features of Make:

- Intelligent tracking of input/output files to skip unnecessary builds
- Dependencies between tasks or targets
- Include other files
- Access environment variables

In addition, here are a few of my favorite Task features:

- Automatic CLI usage output and autocompletion
- Run multiple tasks in parallel. I use this one to start a Go backend service and `npm run dev` for the frontend with a single command
- Control of output syntax which is useful for grouping output in CI environments
- Forward CLI arguments to task commands. I use this to run `task up -d` which will start Docker containers in detached mode
- Global Taskfile: `task` will walk up the filesystem tree until it finds a Taskfile, or use `-g`/`--global` to search your home directory

I like to add a `docs` task to my global Taskfile that will open the usage guide in my default browser:

```yml
version: '3'

tasks:
  docs:
    cmd: open https://taskfile.dev/usage/
```

### Conclusion

Although I am arguing here that a Taskfile is better than a Makefile for many modern projects, I am not saying that Taskfile is a replacement for Makefile in all cases. The two tools have a lot of overlap, but are ultimately designed for different use-cases. There will always be a place for Makefiles, but your modern project is probably better off with a Taskfile. I hope your next step from here is the [installation docs](https://taskfile.dev/installation/) for Task so you can give it a try!

### Links
- Photo by <a href="https://unsplash.com/@ffstop?utm_source=unsplash&utm_medium=referral&utm_content=creditCopyText">Fotis Fotopoulos</a> on <a href="https://unsplash.com/photos/DuHKoV44prg?utm_source=unsplash&utm_medium=referral&utm_content=creditCopyText">Unsplash</a>
- [Task](https://taskfile.dev)
- [Guide to using Makefile with Go](https://levelup.gitconnected.com/a-comprehensive-guide-for-using-makefile-in-golang-projects-c89edebcbe6e)
- [GNU Make](https://www.gnu.org/software/make/)
- ["Simple Makefile"](https://www.gnu.org/software/make/manual/html_node/Simple-Makefile.html)
