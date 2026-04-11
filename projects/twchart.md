# TWChart

[![twchart](https://img.shields.io/badge/GitHub-100000?style=for-the-badge&logo=github&logoColor=white)](https://github.com/calvinmclean/twchart)

TWChart is used to document and display sessions using a Thermoworks thermometer that is recorded in the cloud. Using the exported CSV of thermometer data, this program uses Apache ECharts to display other data on top of the graph.

The original inspiration and use-case for this was tracking bread-making since timing and temperature are important aspects of the process. On top of the timeseries temperature data, this is used to track different stages and ad-hoc notes throughout the process.

TWChart parses a flexible notes format that supports timestamps or elapsed durations, probe assignments, and chronological stages. This enables visualization of processes like coffee roasting, bread making, or BBQ with temperature data overlaid with custom events and annotations.

This project is integrated with [Auto-Roast](/projects/auto_roast.html) for coffee roasting visualization.
