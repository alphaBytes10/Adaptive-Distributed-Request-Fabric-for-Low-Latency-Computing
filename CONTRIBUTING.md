# Contributing to Adaptive Distributed Request Fabric

First off, thank you for considering contributing to the Adaptive Distributed Request Fabric! It's people like you that make open-source a great community.

## How Can I Contribute?

### Reporting Bugs
This section guides you through submitting a bug report. Following these guidelines helps maintainers and the community understand your report, reproduce the behavior, and find related reports.

- **Use a clear and descriptive title** for the issue to identify the problem.
- **Describe the exact steps which reproduce the problem** in as many details as possible.
- **Provide specific examples to demonstrate the steps**. Include links to files or GitHub projects, or copy/paste reproducible code snippets.

### Suggesting Enhancements
This section guides you through submitting an enhancement suggestion, including completely new features and minor improvements to existing functionality.

- **Use a clear and descriptive title** for the issue to identify the suggestion.
- **Provide a step-by-step description of the suggested enhancement** in as many details as possible.
- **Explain why this enhancement would be useful** to most users.

### Pull Requests
- Fill in the required template
- Do not include issue numbers in the PR title
- Include screenshots and animated GIFs in your pull request whenever possible.
- Follow the Go and Rust styleguides.
- Document new code based on the Documentation Styleguide.
- End all files with a newline.

## Development Setup
1. Ensure you have Go 1.21+ and Rust (Cargo) installed.
2. Ensure you have Docker and docker-compose installed to spin up the local Postgres and Redis environments.
3. Run `go mod tidy` in the root folder.
4. Run `docker-compose up -d` to start the infrastructure.
