# AZStocker

[![AZStocker](https://img.shields.io/badge/GitHub-100000?style=for-the-badge&logo=github&logoColor=white)](https://github.com/calvinmclean/azstocker)

Check it out at [azstocker.com](https://azstocker.com)!

AZStocker is a web application that provides searchable and on-demand information about fish stocking in Arizona. The AZ Game and Fish Department provides stocking calendars using Google Sheets, which are generally difficult to read and search. AZStocker parses the sheets and provides this data in a more user-friendly and searchable format. Users can search, sort by recent or upcoming stocking, and save their favorite lakes.

It uses `babyapi` to quickly and easily build a JSON API and Client. Then, Go HTML templates are used to render webpages with [HTMX](https://htmx.org) and [Hyperscript](https://hyperscript.org) for user-interactivity.

This is deployed serverlessly on [fly.io](https://fly.io).
