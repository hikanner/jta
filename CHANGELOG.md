# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2025-10-28

### 🎉 First Stable Release

This is the first stable release of Jta (JSON Translation Agent), an intelligent JSON translation tool powered by AI with Agentic reflection capabilities.

#### ✨ Features

**Core Translation**
- Multi-provider AI translation support (OpenAI, Anthropic, Gemini)
- Agentic reflection mechanism following Andrew Ng's Translation Agent approach
- Three-step translation process: Initial translation → LLM reflection → Automated improvement
- Four-dimensional quality evaluation: accuracy, fluency, style, and terminology

**Terminology Management**
- Directory-based terminology system (`.jta/` by default)
- Separate files for definitions and language-specific translations
- Automatic terminology detection using LLM
- Support for preserve terms (technical terms like "API", "SDK") and consistent terms
- Incremental terminology updates with `--redetect-terms`

**Translation Modes**
- Incremental translation: Only translate new or modified keys (`--incremental`)
- Batch processing with configurable concurrency and batch size
- Key filtering with wildcard patterns (`--keys`, `--exclude-keys`)
- Format protection for HTML, Markdown, URLs, and placeholders

**Language Support**
- 40+ supported languages
- Full RTL (Right-to-Left) support for Arabic, Hebrew, Persian, and Urdu
- Automatic punctuation conversion for RTL languages
- Unicode bidirectional text handling

**User Experience**
- Beautiful terminal UI with Lipgloss styling
- Real-time progress indicators for batch processing
- Detailed statistics and timing information
- Verbose mode for debugging (`-v, --verbose`)
- Non-interactive mode for CI/CD (`-y, --yes`)

#### 🧪 Quality

- **Test Coverage**: 51.9% overall
  - Domain module: 98.6%
  - Format protection: 100%
  - RTL support: 100%
  - Incremental translator: 98.2%
  - UI components: 80.0%
- **Test Suite**: 300+ test cases across 16 test files
- **Code Quality**: 100% Go 1.22+ modernized codebase
- **Error Handling**: Custom error types with rich context

#### 📚 Documentation

- Comprehensive README with 970+ lines
- Contributing guide with 530+ lines
- Detailed execution plan and project roadmap
- Code examples and usage patterns

#### 🏗️ Architecture

**Key Components**
- Provider abstraction for multiple AI services
- Modular translator engine with plugin-style architecture
- Terminology manager with dual-repository pattern
- Format protector with pattern-based detection
- Incremental translator for efficient updates
- Key filter with wildcard matching

**Design Highlights**
- Clean separation of concerns
- Dependency injection for testability
- Context-based cancellation support
- Concurrent-safe implementations

#### 🚀 Performance

- Configurable concurrency (1-5 parallel requests)
- Batch processing (10-50 items per API call)
- Efficient incremental updates (only translate changes)
- Smart caching of terminology and format detection

#### 📦 Distribution

- Multi-platform binaries: macOS (amd64, arm64), Linux (amd64, arm64), Windows (amd64)
- GoReleaser integration for automated builds
- GitHub Actions CI/CD pipeline
- Homebrew formula (coming soon)

#### 🙏 Acknowledgments

Inspired by Andrew Ng's [Translation Agent](https://github.com/andrewyng/translation-agent) approach to using reflection for improving LLM translations.

#### 🔗 Links

- [GitHub Repository](https://github.com/hikanner/jta)
- [Documentation](https://github.com/hikanner/jta#readme)
- [Contributing Guide](https://github.com/hikanner/jta/blob/main/CONTRIBUTING.md)
- [Issue Tracker](https://github.com/hikanner/jta/issues)

---

## [Unreleased]

### Planned for v1.1.0
- YAML file format support
- XML file format support
- Properties file format support
- Local model support (Ollama)
- Translation Memory (TMX) integration
- Interactive review mode
- Configuration file support (`.jtarc`)
- Additional progress visualizations

[1.0.0]: https://github.com/hikanner/jta/releases/tag/v1.0.0
[Unreleased]: https://github.com/hikanner/jta/compare/v1.0.0...HEAD
