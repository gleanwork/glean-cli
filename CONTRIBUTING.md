# Contributing to Glean CLI

> Thank you for your interest in contributing to the Glean CLI! This document provides guidelines and instructions for contributing.

## Development Setup

### Prerequisites

- Go 1.26 or higher
- [mise](https://mise.jdx.dev/getting-started.html) (tool and task runner)
- Git

### Getting Started

1. Fork and clone the repository:
```bash snippet=contributing/snippet-01.sh
git clone https://github.com/YOUR-USERNAME/glean-cli.git
cd glean-cli
```

2. Install development dependencies:
```bash snippet=contributing/snippet-02.sh
mise run setup
```

3. Build the project:
```bash snippet=contributing/snippet-03.sh
mise run build
```

## Development Workflow

1. Create a new branch for your changes:
```bash snippet=contributing/snippet-04.sh
git checkout -b feature/your-feature-name
```

2. Make your changes, following our coding standards and practices.

3. Run tests:
```bash snippet=contributing/snippet-05.sh
# Run tests with verbose output
mise run test

# Run tests with colorized summary
mise run test:summary

# Run all checks (lint + test + build)
mise run test:all
```

4. Run linters:
```bash snippet=contributing/snippet-06.sh
# Run linter
mise run lint

# Run linter with auto-fix
mise run lint:fix
```

5. Install locally to test your changes:
```bash snippet=contributing/snippet-07.sh
mise run install
```

6. Commit your changes using conventional commit messages:
```bash snippet=contributing/snippet-08.sh
git commit -m "feat: add new feature"
git commit -m "fix: resolve issue with X"
```

## Available Tasks

Run `mise tasks` to see all available tasks. Common tasks include:

- `mise run setup`: Install required development dependencies
- `mise run build`: Build the CLI
- `mise run test:all`: Run all checks (used in CI)
- `mise run lint`: Run linters
- `mise run lint:fix`: Run linters with auto-fix
- `mise run install`: Install the CLI locally
- `mise run clean`: Clean build artifacts
- `mise run docs:check`: Verify README code blocks match snippet files
- `mise run docs:sync`: Sync README code blocks from snippet files

## Pull Request Process

1. Update documentation (README.md, code comments) if you're changing functionality. README code examples live in `snippets/` — edit the `.sh` files there, then run `mise run docs:sync` to update the README.
2. Add tests for any new features.
3. Ensure all tests pass and linters are clean.
4. Push your changes and create a pull request.
5. Fill out the pull request template with all required information.

## Code Style

- Follow standard Go conventions and idioms
- Use meaningful variable and function names
- Add comments for public APIs and complex logic
- Keep functions focused and concise
- Write tests for new functionality

## Testing

- Write unit tests for new features
- Use table-driven tests where appropriate
- Mock external dependencies
- Aim for high test coverage of critical paths
- Test both success and error cases

## Documentation

- Update README.md for user-facing changes
- Add godoc comments for exported functions and types
- Include examples in documentation where helpful
- Keep documentation up to date with code changes

## Release Process

1. Update version numbers in relevant files
2. Create a new release tag
3. Build and publish new artifacts

## Getting Help

- Open an issue for bugs or feature requests
- Ask questions in pull requests
- Be respectful and constructive in discussions

## License

By contributing to Glean CLI, you agree that your contributions will be licensed under the MIT License.