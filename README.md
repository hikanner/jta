# Jta - JSON Translation Agent

> AI-powered Agentic JSON Translation tool with intelligent terminology management

Jta is a command-line tool that uses AI to translate JSON internationalization files with high accuracy and consistency. It features automatic terminology detection, format preservation, and incremental translation capabilities.

## ✨ Features

- 🤖 **Agentic Translation**: AI intelligently manages the entire translation process
- 📚 **Smart Terminology Management**: Automatic detection and consistent translation
- 🔒 **Format Protection**: Preserves placeholders, HTML tags, and special markers
- ⚡ **Incremental Translation**: Only translates new or changed content
- 🎯 **Key Filtering**: Selectively translate specific sections
- 🌍 **RTL Language Support**: Proper handling of right-to-left languages
- 🚀 **High Performance**: Batch processing with concurrency control
- 🎨 **Multiple AI Providers**: OpenAI, Anthropic Claude, Google Gemini

## 📦 Installation

### Using Go Install

```bash
go install github.com/hikanner/jta/cmd/jta@latest
```

### From Source

```bash
git clone https://github.com/hikanner/jta.git
cd jta
go build -o jta cmd/jta/main.go
```

## 🚀 Quick Start

### Basic Usage

```bash
# Translate to a single language
jta en.json --to zh

# Translate to multiple languages
jta en.json --to zh,ja,ko

# Specify output directory
jta en.json --to zh --output ./locales/
```

### With AI Provider Configuration

```bash
# Using environment variables (recommended)
export OPENAI_API_KEY=sk-...
jta en.json --to zh

# Or specify directly
jta en.json --to zh --provider anthropic --api-key sk-ant-...
```

### Advanced Usage

```bash
# Skip terminology detection (use existing)
jta en.json --to zh --skip-terms

# Disable terminology management completely
jta en.json --to zh --no-terminology

# Translate specific keys only
jta en.json --to zh --keys "settings.*,user.*"

# Exclude certain keys
jta en.json --to zh --exclude-keys "admin.*,internal.*"

# Force complete re-translation
jta en.json --to zh --force

# Non-interactive mode (for CI/CD)
jta en.json --to zh,ja,ko -y
```

## 📖 Documentation

### Terminology Management

Jta automatically detects important terminology in your source file and ensures consistent translation:

- **Preserve Terms**: Brand names, technical terms that should never be translated
- **Consistent Terms**: Domain terms that must be translated uniformly

The terminology is saved to `.jta-terminology.json` and can be manually edited.

### Incremental Translation

When you run Jta on an existing translation, it intelligently:

1. Detects new keys
2. Identifies modified content
3. Preserves unchanged translations
4. Removes deleted keys

This saves time and API costs by only translating what's necessary.

### Format Protection

Jta automatically protects:

- Variables: `{variable}`, `{{count}}`, `%s`
- HTML tags: `<b>`, `<span class="highlight">`
- URLs: `https://example.com`
- Markdown: `**bold**`, `*italic*`

## 🎯 Supported AI Providers

| Provider | Models | Environment Variable |
|----------|--------|---------------------|
| OpenAI | GPT-4o, GPT-4 Turbo | `OPENAI_API_KEY` |
| Anthropic | Claude 3.5 Sonnet | `ANTHROPIC_API_KEY` |
| Google | Gemini 2.0 Flash | `GEMINI_API_KEY` |

## 🌍 Supported Languages

Jta supports 25+ languages including:

- English, Chinese (Simplified/Traditional), Japanese, Korean
- Spanish, French, German, Italian, Portuguese
- Arabic, Hebrew (with RTL support)
- And many more...

## 💡 Examples

### Example 1: First-time Translation

```bash
$ jta en.json --to zh

📖 Loading source file...
🔍 Detecting terminology...
✨ Detected 8 terms
Save terminology to .jta-terminology.json? [Y/n] y
🤖 Translating...
💾 Saving translation...

📊 Statistics:
   Total items: 100
   Success: 100
   Failed: 0
   Duration: 30s
   API calls: 5
   Output: zh.json
```

### Example 2: Incremental Update

```bash
$ jta en.json --to zh

📖 Loading source file...
🔍 Analyzing changes...
   ✨ New: 5 keys
   🔄 Modified: 2 keys
   ✅ Unchanged: 93 keys

💡 Will translate 7 keys, keep 93 unchanged. Continue? [Y/n] y
🤖 Translating...

📊 Statistics:
   Total items: 7
   Success: 7
   Duration: 3s
   API calls: 1
   Output: zh.json
```

## 🛠 Configuration

### Environment Variables

```bash
# AI Provider API Keys
export OPENAI_API_KEY=sk-...
export ANTHROPIC_API_KEY=sk-ant-...
export GEMINI_API_KEY=...
```

### Command-line Options

```
Flags:
  --to string              Target language(s), comma-separated (required)
  --provider string        AI provider (openai, anthropic, google) (default "openai")
  --model string           Model name (uses default if not specified)
  --api-key string         API key (or use environment variable)
  -o, --output string      Output file or directory
  --terminology string     Terminology file path (default ".jta-terminology.json")
  --skip-terms            Skip term detection (still translates missing terms)
  --no-terminology        Disable terminology management completely
  --keys string           Only translate specified keys (glob patterns)
  --exclude-keys string   Exclude specified keys (glob patterns)
  --force                 Force complete re-translation
  --batch-size int        Batch size for translation (default 20)
  --concurrency int       Concurrency for batch processing (default 3)
  -y, --yes               Non-interactive mode
  -v, --verbose           Verbose output
```

## 🤝 Contributing

Contributions are welcome! Please read our [Contributing Guide](CONTRIBUTING.md) for details.

## 📄 License

MIT License - see [LICENSE](LICENSE) for details.

## 🙏 Acknowledgments

- Inspired by Andrew Ng's Translation Agent
- Built with official AI provider SDKs
- Thanks to the Go community for excellent tools and libraries

## 📞 Support

- 🐛 [Report Issues](https://github.com/hikanner/jta/issues)
- 💬 [Discussions](https://github.com/hikanner/jta/discussions)
- 📧 Email: support@example.com

---

**Made with ❤️ by the Jta team**
