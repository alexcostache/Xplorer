# Contributing to Xplorer

First off, thank you for considering contributing to Xplorer! It's people like you that make Xplorer such a great tool.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [How Can I Contribute?](#how-can-i-contribute)
- [Development Setup](#development-setup)
- [Coding Guidelines](#coding-guidelines)
- [Commit Messages](#commit-messages)
- [Pull Request Process](#pull-request-process)
- [Testing](#testing)

## Code of Conduct

This project and everyone participating in it is governed by respect and professionalism. By participating, you are expected to uphold this standard. Please report unacceptable behavior to the project maintainers.

## Getting Started

1. Fork the repository on GitHub
2. Clone your fork locally
3. Create a new branch for your feature or bugfix
4. Make your changes
5. Test your changes
6. Submit a pull request

## How Can I Contribute?

### Reporting Bugs

Before creating bug reports, please check the existing issues to avoid duplicates. When you create a bug report, include as many details as possible:

- **Use a clear and descriptive title**
- **Describe the exact steps to reproduce the problem**
- **Provide specific examples** to demonstrate the steps
- **Describe the behavior you observed** and what you expected
- **Include screenshots** if applicable
- **Include your environment details**: OS, Go version, terminal emulator

**Bug Report Template:**
```markdown
**Description:**
A clear description of the bug.

**Steps to Reproduce:**
1. Go to '...'
2. Press '...'
3. See error

**Expected Behavior:**
What you expected to happen.

**Actual Behavior:**
What actually happened.

**Environment:**
- OS: [e.g., macOS 14.0, Ubuntu 22.04]
- Go Version: [e.g., 1.24.2]
- Terminal: [e.g., iTerm2, GNOME Terminal]
- Xplorer Version: [e.g., v1.0.0]

**Screenshots:**
If applicable, add screenshots.
```

### Suggesting Enhancements

Enhancement suggestions are tracked as GitHub issues. When creating an enhancement suggestion, include:

- **Use a clear and descriptive title**
- **Provide a detailed description** of the suggested enhancement
- **Explain why this enhancement would be useful**
- **List any alternative solutions** you've considered

**Enhancement Template:**
```markdown
**Feature Description:**
A clear description of the feature.

**Use Case:**
Describe the problem this feature would solve.

**Proposed Solution:**
How you envision this feature working.

**Alternatives Considered:**
Other approaches you've thought about.

**Additional Context:**
Any other context, mockups, or examples.
```

### Your First Code Contribution

Unsure where to begin? Look for issues labeled:
- `good first issue` - Simple issues perfect for newcomers
- `help wanted` - Issues where we need community help
- `documentation` - Documentation improvements

## Development Setup

### Prerequisites
- Go 1.24.2 or higher
- Git
- A terminal emulator

### Setup Steps

```bash
# 1. Fork and clone the repository
git clone https://github.com/YOUR_USERNAME/Xplorer.git
cd Xplorer

# 2. Add upstream remote
git remote add upstream https://github.com/alexcostache/Xplorer.git

# 3. Install dependencies
go mod download

# 4. Create a branch for your changes
git checkout -b feature/your-feature-name

# 5. Make your changes and test
go build -o xp
./xp

# 6. Run tests
go test ./...

# 7. Run with coverage
go test -cover ./...
```

## Coding Guidelines

### Go Style Guide

Follow the official [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments) and [Effective Go](https://golang.org/doc/effective_go.html).

### Project-Specific Guidelines

1. **Module Structure**: Keep the modular architecture intact
   - Each module should have a single responsibility
   - Avoid circular dependencies
   - Use dependency injection

2. **Naming Conventions**:
   - Use descriptive variable names
   - Follow Go naming conventions (camelCase for private, PascalCase for public)
   - Avoid abbreviations unless widely understood

3. **Comments**:
   - Add comments for exported functions and types
   - Explain complex logic
   - Keep comments up-to-date with code changes

4. **Error Handling**:
   - Always handle errors explicitly
   - Provide context in error messages
   - Use meaningful error types

5. **Code Organization**:
   ```go
   // Good
   func (m *Manager) ProcessFile(path string) error {
       if err := m.validate(path); err != nil {
           return fmt.Errorf("validation failed: %w", err)
       }
       // ... rest of logic
   }

   // Avoid
   func ProcessFile(path string) {
       // ... no error handling
   }
   ```

6. **Testing**:
   - Write unit tests for new features
   - Maintain test coverage above 70%
   - Test edge cases and error conditions
   - Use table-driven tests where appropriate

### File Organization

```
internal/
├── app/            # Application orchestration
├── config/         # Configuration management
├── theme/          # Theme system
├── bookmark/       # Bookmark management
├── filesystem/     # File navigation
├── preview/        # File preview
├── fileops/        # File operations
└── ui/             # User interface
```

When adding new features:
- Create a new module in `internal/` if it's a major feature
- Add to existing modules if it extends current functionality
- Update documentation in relevant `.md` files

## Commit Messages

Write clear, concise commit messages following these guidelines:

### Format
```
<type>(<scope>): <subject>

<body>

<footer>
```

### Types
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

### Examples

```bash
# Good commit messages
feat(ui): add file size column to file list
fix(preview): handle binary files correctly
docs(readme): update installation instructions
refactor(filesystem): simplify navigation logic
test(bookmark): add tests for toggle functionality

# Bad commit messages
fixed stuff
update
changes
wip
```

### Detailed Example
```
feat(theme): add custom theme loading from directory

- Add support for loading themes from ~/.xplorer/themes/
- Implement theme validation
- Add error handling for malformed theme files
- Update theme selector to show custom themes

Closes #123
```

## Pull Request Process

1. **Update Documentation**: Ensure all documentation is updated
2. **Add Tests**: Include tests for new features
3. **Update CHANGELOG**: Add entry to CHANGELOG.md (if exists)
4. **Run Tests**: Ensure all tests pass
5. **Format Code**: Run `go fmt ./...`
6. **Lint Code**: Run `go vet ./...`

### PR Template

```markdown
## Description
Brief description of changes.

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
Describe how you tested your changes.

## Checklist
- [ ] Code follows project style guidelines
- [ ] Self-review completed
- [ ] Comments added for complex code
- [ ] Documentation updated
- [ ] Tests added/updated
- [ ] All tests pass
- [ ] No new warnings generated

## Related Issues
Closes #(issue number)
```

### Review Process

1. Maintainers will review your PR
2. Address any requested changes
3. Once approved, your PR will be merged
4. Your contribution will be credited in the release notes

## Testing

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run specific test
go test -v -run TestFunctionName ./...

# Run benchmarks
go test -bench=. ./...
```

### Writing Tests

```go
// Example test structure
func TestFeatureName(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
        wantErr  bool
    }{
        {
            name:     "valid input",
            input:    "test",
            expected: "TEST",
            wantErr:  false,
        },
        {
            name:     "empty input",
            input:    "",
            expected: "",
            wantErr:  true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := YourFunction(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("expected error: %v, got: %v", tt.wantErr, err)
            }
            if result != tt.expected {
                t.Errorf("expected: %v, got: %v", tt.expected, result)
            }
        })
    }
}
```

## Architecture Guidelines

When contributing, maintain the existing architecture:

1. **Separation of Concerns**: Keep modules independent
2. **Dependency Injection**: Pass dependencies through constructors
3. **Single Responsibility**: Each module has one clear purpose
4. **Testability**: Write testable code with clear interfaces

See [ARCHITECTURE.md](ARCHITECTURE.md) for detailed architecture documentation.

## Questions?

Feel free to:
- Open an issue for questions
- Join discussions in existing issues
- Reach out to maintainers

## Recognition

Contributors will be recognized in:
- README.md contributors section
- Release notes
- GitHub contributors page

Thank you for contributing to Xplorer! 🎉