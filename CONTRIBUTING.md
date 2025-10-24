# Contributing to Jta

Thank you for your interest in contributing to Jta! This document provides guidelines and instructions for contributing.

## 🌟 Ways to Contribute

There are many ways to contribute to Jta:

- 🐛 **Report bugs**: Found an issue? Let us know!
- 💡 **Suggest features**: Have an idea? We'd love to hear it!
- 📖 **Improve documentation**: Help others understand Jta better
- 🔧 **Fix bugs**: Pick an issue and submit a PR
- ✨ **Add features**: Implement new capabilities
- 🧪 **Write tests**: Improve test coverage
- 🌍 **Translate**: Help translate Jta's own messages

## 🚀 Getting Started

### Prerequisites

- **Go 1.25+**: [Install Go](https://go.dev/doc/install)
- **Git**: [Install Git](https://git-scm.com/downloads)
- **AI Provider API Key**: For testing (OpenAI, Anthropic, or Google)

### Development Setup

1. **Fork the repository**

   Click the "Fork" button on GitHub to create your own copy.

2. **Clone your fork**

   ```bash
   git clone https://github.com/YOUR_USERNAME/jta.git
   cd jta
   ```

3. **Add upstream remote**

   ```bash
   git remote add upstream https://github.com/hikanner/jta.git
   ```

4. **Install dependencies**

   ```bash
   go mod download
   ```

5. **Set up environment variables**

   ```bash
   export OPENAI_API_KEY=sk-...
   # Or use .env file (not tracked in git)
   echo "OPENAI_API_KEY=sk-..." > .env
   ```

6. **Run tests**

   ```bash
   go test ./...
   ```

7. **Build the project**

   ```bash
   go build -o jta cmd/jta/main.go
   ```

8. **Try it out**

   ```bash
   ./jta examples/en.json --to zh
   ```

## 📝 Development Workflow

### Creating a Feature Branch

Always create a new branch for your work:

```bash
git checkout -b feature/your-feature-name
# or
git checkout -b fix/your-bug-fix
```

Branch naming conventions:
- `feature/` - New features
- `fix/` - Bug fixes
- `docs/` - Documentation changes
- `test/` - Test additions or modifications
- `refactor/` - Code refactoring
- `perf/` - Performance improvements

### Making Changes

1. **Write clean, readable code**
   - Follow Go conventions and best practices
   - Use meaningful variable and function names
   - Add comments for complex logic
   - Keep functions focused and small

2. **Add tests**
   - Write unit tests for new functions
   - Ensure existing tests still pass
   - Aim for high test coverage (target: 60%+)

3. **Update documentation**
   - Update README.md if adding user-facing features
   - Add godoc comments for exported functions
   - Update CHANGELOG.md (if exists)

### Code Style

We follow standard Go formatting:

```bash
# Format your code
gofmt -w .

# Run linter
golangci-lint run

# Check for common issues
go vet ./...
```

### Testing

#### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with detailed coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run specific package tests
go test ./internal/translator/...

# Run with verbose output
go test -v ./...
```

#### Writing Tests

Follow these guidelines:

1. **Test file naming**: `*_test.go`
2. **Test function naming**: `Test<FunctionName>`
3. **Table-driven tests** for multiple scenarios:

```go
func TestTranslate(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
        wantErr  bool
    }{
        {
            name:     "simple translation",
            input:    "Hello",
            expected: "你好",
            wantErr:  false,
        },
        // More test cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := Translate(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("Translate() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if result != tt.expected {
                t.Errorf("Translate() = %v, want %v", result, tt.expected)
            }
        })
    }
}
```

### Committing Changes

We follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `test`: Adding or updating tests
- `refactor`: Code refactoring
- `perf`: Performance improvements
- `chore`: Maintenance tasks
- `ci`: CI/CD changes

**Examples:**

```bash
# Feature
git commit -m "feat(translator): add support for custom prompts"

# Bug fix
git commit -m "fix(format): preserve markdown bold syntax in translations"

# Documentation
git commit -m "docs(readme): add troubleshooting section"

# Multiple changes
git commit -m "feat(reflection): optimize quality check algorithm

- Reduce API calls by 30%
- Add batch processing for issues
- Improve error handling

Closes #123"
```

### Submitting a Pull Request

1. **Update your branch**

   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

2. **Push to your fork**

   ```bash
   git push origin feature/your-feature-name
   ```

3. **Create Pull Request**

   - Go to the original repository on GitHub
   - Click "New Pull Request"
   - Select your fork and branch
   - Fill in the PR template

4. **PR Checklist**

   - [ ] Tests pass (`go test ./...`)
   - [ ] Code is formatted (`gofmt -w .`)
   - [ ] Documentation is updated
   - [ ] Commit messages follow conventions
   - [ ] PR description explains changes clearly
   - [ ] Linked to related issues (if any)

### PR Review Process

1. **Automated checks** run (tests, linting)
2. **Maintainers review** your code
3. **Address feedback** if requested
4. **Approval** by at least one maintainer
5. **Merge** by maintainer

## 🏗️ Project Structure

Understanding the codebase:

```
jta/
├── cmd/
│   └── jta/
│       └── main.go          # CLI entry point
├── internal/
│   ├── cli/                 # Command-line interface
│   │   ├── app.go           # Main application logic
│   │   └── root.go          # Cobra root command
│   ├── domain/              # Domain models
│   │   ├── language.go      # Language definitions
│   │   ├── terminology.go   # Terminology models
│   │   └── translation.go   # Translation models
│   ├── provider/            # AI provider implementations
│   │   ├── openai.go        # OpenAI integration
│   │   ├── anthropic.go     # Anthropic integration
│   │   └── google.go        # Google Gemini integration
│   ├── translator/          # Translation engine
│   │   ├── engine.go        # Core translation logic
│   │   ├── batch.go         # Batch processor
│   │   └── reflection.go    # Reflection mechanism ⭐
│   ├── terminology/         # Terminology management
│   │   ├── manager.go       # Terminology manager
│   │   ├── detector.go      # LLM-based detection
│   │   └── repository.go    # JSON storage
│   ├── format/              # Format protection
│   │   └── protector.go     # Format element preservation
│   ├── incremental/         # Incremental translation
│   │   └── translator.go    # Diff analysis
│   ├── keyfilter/           # Key filtering
│   │   ├── filter.go        # Filter logic
│   │   └── matcher.go       # Pattern matching
│   ├── rtl/                 # RTL language support
│   │   └── processor.go     # Bidirectional text handling
│   ├── ui/                  # Terminal UI
│   │   ├── styles.go        # Lipgloss styles
│   │   └── printer.go       # Styled output
│   └── utils/               # Utilities
│       └── json.go          # JSON helpers
├── examples/                # Example files
│   └── en.json              # Sample source file
├── go.mod                   # Go modules
├── go.sum                   # Dependencies
├── README.md                # Main documentation
├── CONTRIBUTING.md          # This file
├── LICENSE                  # MIT License
└── EXECUTION_PLAN.md        # Development roadmap
```

## 🎯 Areas Needing Help

Current priorities:

### High Priority

- [ ] **Provider tests**: Add unit tests for AI provider implementations
- [ ] **Terminology tests**: Test terminology detection and management
- [ ] **Integration tests**: End-to-end translation workflows
- [ ] **Error handling**: Improve error messages and recovery

### Medium Priority

- [ ] **Performance optimization**: Profile and optimize hot paths
- [ ] **Memory efficiency**: Reduce memory usage for large files
- [ ] **Better progress indicators**: Real-time translation progress
- [ ] **Configuration file support**: YAML/TOML config files

### Low Priority

- [ ] **Additional providers**: Azure OpenAI, local models
- [ ] **Alternative formats**: YAML, XML, PO file support
- [ ] **Translation memory**: TMX integration
- [ ] **Web UI**: Browser-based interface

## 🐛 Bug Reports

### Before Submitting

1. **Search existing issues** to avoid duplicates
2. **Try the latest version** to see if it's already fixed
3. **Collect information**:
   - Jta version (`jta --version`)
   - Go version (`go version`)
   - Operating system
   - Steps to reproduce
   - Expected vs actual behavior
   - Error messages and logs

### Bug Report Template

```markdown
## Bug Description
A clear description of what the bug is.

## To Reproduce
Steps to reproduce the behavior:
1. Run command '...'
2. With file '...'
3. See error

## Expected Behavior
What you expected to happen.

## Actual Behavior
What actually happened.

## Environment
- Jta version: x.y.z
- Go version: 1.25.0
- OS: macOS 14.0
- Provider: OpenAI GPT-4o

## Logs
```
Paste relevant error messages or logs here
```

## Additional Context
Any other context about the problem.
```

## 💡 Feature Requests

### Before Requesting

1. **Search existing issues** to see if it's already requested
2. **Consider if it fits** the project's goals
3. **Think about implementation** complexity

### Feature Request Template

```markdown
## Feature Description
A clear description of the feature you'd like.

## Use Case
Explain why this feature would be useful.
Example scenarios where you'd use it.

## Proposed Solution
How you envision this working.

## Alternatives Considered
Other approaches you've thought about.

## Additional Context
Mockups, examples, or references.
```

## 📚 Documentation

Documentation improvements are always welcome!

### Types of Documentation

- **README.md**: User-facing documentation
- **Code comments**: Inline documentation (godoc)
- **Wiki**: Detailed guides and tutorials
- **Examples**: Sample code and usage patterns

### Writing Guidelines

- Use clear, simple language
- Include code examples
- Add screenshots/GIFs for UI features
- Keep it up-to-date with code changes

## 🤝 Code Review

### For Contributors

- Be patient and respectful
- Be open to feedback
- Ask questions if unclear
- Don't take criticism personally
- Learn from the review process

### For Reviewers

- Be constructive and helpful
- Explain why changes are needed
- Suggest alternatives
- Acknowledge good work
- Be timely in responses

## 📜 Code of Conduct

### Our Standards

- **Be respectful**: Treat everyone with respect
- **Be inclusive**: Welcome diverse perspectives
- **Be collaborative**: Work together constructively
- **Be professional**: Maintain professionalism

### Unacceptable Behavior

- Harassment or discrimination
- Trolling or insulting comments
- Personal or political attacks
- Publishing private information
- Other unprofessional conduct

### Enforcement

Violations may result in:
1. Warning
2. Temporary ban
3. Permanent ban

Report issues to: conduct@example.com

## 🎓 Learning Resources

### Go Programming

- [A Tour of Go](https://go.dev/tour/)
- [Effective Go](https://go.dev/doc/effective_go)
- [Go by Example](https://gobyexample.com/)

### AI/LLM Integration

- [OpenAI API Documentation](https://platform.openai.com/docs)
- [Anthropic Claude Documentation](https://docs.anthropic.com/)
- [Andrew Ng's Translation Agent](https://github.com/andrewyng/translation-agent)

### Testing

- [Go Testing Package](https://pkg.go.dev/testing)
- [Table Driven Tests](https://go.dev/wiki/TableDrivenTests)

## 💬 Communication

### Channels

- **GitHub Issues**: Bug reports and feature requests
- **GitHub Discussions**: General questions and ideas
- **Pull Requests**: Code contributions and reviews

### Getting Help

- Check the [README](README.md) first
- Search [existing issues](https://github.com/hikanner/jta/issues)
- Ask in [Discussions](https://github.com/hikanner/jta/discussions)

## 🙏 Thank You!

Every contribution, no matter how small, helps make Jta better. We appreciate your time and effort!

---

**Questions?** Feel free to ask in [Discussions](https://github.com/hikanner/jta/discussions)

**Happy Contributing! 🎉**
