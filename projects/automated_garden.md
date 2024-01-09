# Automated Garden

[![automated-garden](https://img.shields.io/badge/GitHub-100000?style=for-the-badge&logo=github&logoColor=white)](https://github.com/calvinmclean/automated-garden)

Automated Garden is an open source irrigation controller. The repository contains code for ESP32 microcontrollers to toggle valves and report sensor data as well as a Go server which provides the API, scheduling, and serves the UI. Currently, the UI is created with Svelte but I am re-writing to use HTMX. It is integrated with the Netatmo API to allow watering based on real measurements from your Netatmo weather sensors.

The backend communicates with multiple devices using the MQTT message protocol. The individual controllers then push data and watering logs to InfluxDB via MQTT and Telegraf. Grafana is used to visualize sensor data and Prometheus metrics from the application.

The architecture of the project uses InfluxDB to store time-series data from sensors and logs to confirm that the controller successfully completed watering.

This project also was the starting point for the API structure that was used to create `babyapi`.
