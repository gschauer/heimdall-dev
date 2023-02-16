# Heimdall (proof of concept)

## Disclaimer

The content of this repository was a one-time feasibility study and won't be continued.

## Introduction

This project is an experimental project, which should run checks against software releases.
It collects facts from Git repos, Jira, etc. and verifies that it meets the requirements, including, but not limited to

* code coverage,
* Git PRs approved, 
* tickets approved and closed,
* etc.

## Configuration

Configuration is mostly done by values provided in the `cfg` package and environment variables:

* `GIT_USERNAME`
* `GIT_PASSWORD`
* `GITHUB_API_URL`
* `GITHUB_TOKEN`
* `JIRA_BASE_URL`
* `JIRA_TOKEN`
