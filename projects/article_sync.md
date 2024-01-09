# Article Sync

[![article-sync](https://img.shields.io/badge/GitHub-100000?style=for-the-badge&logo=github&logoColor=white)](https://github.com/calvinmclean/article-sync)

Article Sync is a simple CLI written in Go + a GitHub Action to easily run it in GitHub. This enables synchronizing markdown articles (like the ones on this site!) to blog/article platforms like [dev.to](https://dev.to/calvinmclean/manage-dev-articles-with-git-and-github-actions-13md).

The GitHub Action will leave comments on PRs with an explanation of changes after merging. Then, once merged to `main`, it will synchronize with the platforms and commit any new article IDs and links to the `article.json` files.

A fun feature of `article-sync` is that you can provide a link to an image from [gopherize.me](https://gopherize.me) to automatically create an article banner with the title and the generated gopher image.
