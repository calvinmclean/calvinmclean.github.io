## Introduction

Git is a familiar tool for most, if not all, software engineers. While it is generally used for managing code files, it works with any plain text files. Continuous integration and delivery are processes built around git repositories and can be useful for more than just testing and deploying software. After all, publishing a draft article is similar to merging from a development branch into main.

I was considering hosting my articles on a GitHub Pages site that is independent of any other platform. This would of course mean I need to store articles in git and also post them on DEV (and potentially other sites). Being a software engineer, I saw this as a perfect opportunity for building an automation tool. A tool that can synchronize articles from a file-based structure to one or more sites, provides a few benefits:
  - Use your preferred editor for Markdown files without having to copy/paste to web editors
  - Invite others to review your PR and add comments
  - Potential to host GitHub Pages for presenting articles with your own branding
  - A single "source of truth" is especially useful when trying keep the article consistent across multiple platforms
  - Integrate with GitHub Actions to automate the entire process

## The Software
Luckily, Forem/DEV is open source and provides great [API documentation and specification](https://developers.forem.com/api). I used [`oapi-codegen`](https://github.com/deepmap/oapi-codegen/) to automatically generate a Go API client. Then, I simply had to walk the root articles directory and:
  1. Parse a metadata file (`article.json`) to check if it has an ID
  2. If it has an ID, compare to the published article and update if there is any difference
  3. If there is no ID, post a new article and save the ID back to the `article.json` file

This relies on a simple file structure where each article is in its own directory under the root:
```
articles/
├── my-article
│   ├── article.json
│   └── article.md
└── my-other-article
    ├── article.json
    └── article.md
```

Then, you just need an API key to run:
```shell
# dry-run to see logs of what would happen (update or post)
go run -mod=mod github.com/calvinmclean/article-sync@v1.0.1 \
  --api-key $API_KEY \
  --dry-run

# run for real to synchronize
go run -mod=mod github.com/calvinmclean/article-sync@v1.0.1 \
  --api-key $API_KEY
```

## GitHub Action
Use the [`calvinmclean/article-sync`](https://github.com/marketplace/actions/article-sync) GitHub Action to:

- Comment a summary of changes when opening a PR
    ```yaml
    name: Synchronization summary
    on:
      pull_request:
        branches:
          - main
    jobs:
      summary:
        runs-on: ubuntu-latest
        steps:
          - uses: actions/checkout@v3
          - uses: calvinmclean/article-sync@v1.0.1
            with:
              type: summary
              api_key: ${{ secrets.DEV_TO_API_KEY }}
    ```

- After pushing to main, post/update articles and make a commit with new IDs if articles are created
    ```yaml
    name: Synchronize and commit
    on:
      push:
        branches:
          - main
    jobs:
      synchronize:
        runs-on: ubuntu-latest
        steps:
          - uses: actions/checkout@v3
          - uses: calvinmclean/article-sync@v1.0.1
            with:
              type: synchronize
              api_key: ${{ secrets.DEV_TO_API_KEY }}
    ```

Check out the [Action Marketplace page](https://github.com/marketplace/actions/article-sync) to see the latest setup instructions.

You can see the PR that posted this article [here](https://github.com/calvinmclean/articles/pull/1).
The PR that added this new line is [here](https://github.com/calvinmclean/articles/pull/2).

## Roadmap
I plan to add a few more features to the CLI and GitHub Action:
- Attach cover images to articles
- Manage tags on articles
- Support other article platforms and GitHub Pages site

Check out my [GitHub repository](https://github.com/calvinmclean/article-sync) and give it a Star if you want to keep up with these changes! Feel free to leave a comment here or open an issue in the repository if you have any questions about using it. Of course contributions are always welcome if you have any improvements or features you would like to add. Thanks for reading!
