# Contributing to ADX

First off, thank you for considering contributing to ADX! It's people like you that make ADX such a great tool for the Go community.

## Where to start?

1. **Bug Reports**: If you find a bug, please open an issue describing the bug, how to reproduce it, and your environment.
2. **Feature Requests**: If you have an idea for a new feature, open an issue to discuss it before you start coding.
3. **Pull Requests**: We welcome PRs! Please ensure your code adheres to our guidelines.

## Development Setup

To get started with development, you'll need Go 1.26 or higher installed on your machine, as declared in `go.mod`.

1. Fork the repository
2. Clone your fork: `git clone https://github.com/your-username/adx.git`
3. Enter the directory: `cd adx`
4. Run tests to ensure everything works: `make test`

## Project Structure

- `adx.go`: Contains the core logic for resolving Android device codes.
- `brand.go`: Contains brand shape and prefix logic.
- `rules.go`: Fallback custom device matching logic.
- `names.go`: Dependency-free accessor for consumers that want only the strings.
- `internal/catalog/`: Contains the generated catalogue `catalog_gen.go` which provides the exhaustive Android device mappings.
- `gen/`: The importer that produces `internal/catalog/catalog_gen.go`. Edit the transformations here, never the generated file — see `make catalog`.

## Writing Tests

We maintain a strict 100% test coverage requirement for the core library. 

When you add a new feature or fix a bug, please write tests for it.

To run the test suite and check coverage:

```bash
make test
make cover
```

## Pull Request Process

1. Create a new branch for your feature or bugfix: `git checkout -b feature-name`
2. Write your code and tests.
3. Ensure the code is formatted using `gofmt`.
4. Run `make lint` and ensure there are no errors.
5. Commit your changes with a descriptive commit message.
6. Push to your fork and submit a Pull Request against the `main` branch.

## Code of Conduct

Please note that this project is released with a [Contributor Code of Conduct](CODE_OF_CONDUCT.md). By participating in this project you agree to abide by its terms.
