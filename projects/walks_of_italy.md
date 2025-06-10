# Walks of Italy Tool

[![walks-of-italy](https://img.shields.io/badge/GitHub-100000?style=for-the-badge&logo=github&logoColor=white)](https://github.com/calvinmclean/walks-of-italy)

This project enables tracking availability of tours provided by [Walks of Italy](https://www.walksofitaly.com).

It has the following features:
- Interact with a local LLM running in Ollama to discuss tour details and availability using provided tools
- Uses [babyapi](https://github.com/calvinmclean/babyapi) to implement a CRUD API for managing tours
- Provides a simple UI to see tracked tours and latest information
- Sends notifications when new availabilities are posted for watched tours
- Uses SQLite to store watched tours and latest availability
