# Jta - JSON Translation Agent

[![Go Version](https://img.shields.io/badge/Go-1.25%2B-blue)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Test Coverage](https://img.shields.io/badge/coverage-34.2%25-yellow)](coverage.out)
[![Go Report Card](https://img.shields.io/badge/go%20report-A+-brightgreen)](https://goreportcard.com/)

> AI-powered Agentic JSON Translation tool with intelligent quality optimization

Jta is a production-ready command-line tool that uses AI to translate JSON internationalization files with **exceptional accuracy and consistency**. Inspired by Andrew Ng's Translation Agent, it features automatic terminology detection, format preservation, and a **lightweight reflection mechanism** for quality self-optimization.

## ✨ Key Features

### 🤖 Agentic Translation with Self-Optimization

- **Lightweight Reflection Engine**: Inspired by Andrew Ng's approach, optimized to 1.2-1.5x API calls (vs 3x for full reflection)
- **Quality Auto-Check**: Automatically validates format integrity, terminology consistency, and completeness
- **Selective Improvement**: Only fixes Critical/High severity issues for optimal cost-efficiency
- **Batch Optimization**: Single API call handles multiple improvements

### 📚 Intelligent Terminology Management

- **Automatic Detection**: Uses LLM to identify important terms in your content
- **Preserve Terms**: Brand names, technical terms that should never be translated
- **Consistent Terms**: Domain-specific terms translated uniformly across all content
- **Editable Dictionary**: Saved to `.jta-terminology.json` for manual refinement

### 🔒 Robust Format Protection

Automatically preserves:
- **Placeholders**: `{variable}`, `{{count}}`, `%s`, `%(name)d`
- **HTML Tags**: `<b>`, `<span class="highlight">`, `<a href="...">`
- **URLs**: `https://example.com`, `http://api.example.com/v1`
- **Markdown**: `**bold**`, `*italic*`, `[link](url)`

### ⚡ Smart Incremental Translation

- Only translates new or modified content
- Preserves existing high-quality translations
- Automatically removes obsolete keys
- Saves time and API costs (typically 80-90% reduction on updates)

### 🎯 Flexible Key Filtering

- **Glob Patterns**: `settings.*`, `user.**`, `*.title`
- **Precise Control**: Include or exclude specific sections
- **Recursive Wildcards**: Translate entire subsections with `**`

### 🌍 RTL Language Support

- Proper bidirectional text handling for Arabic, Hebrew, Persian, Urdu
- Automatic direction markers for LTR content in RTL context
- Smart punctuation conversion for Arabic-script languages

### 🚀 Production-Ready Performance

- Batch processing with configurable concurrency
- Retry logic with exponential backoff
- Graceful error handling and recovery
- Progress indicators and detailed statistics

### 🎨 Multi-Provider Support

- **OpenAI**: GPT-4o, GPT-4 Turbo, GPT-3.5
- **Anthropic**: Claude 3.5 Sonnet, Claude 3 Opus/Haiku
- **Google**: Gemini 2.0 Flash (experimental)

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

## 🏗️ Architecture

Jta follows a clean, modular architecture:

```
┌─────────────────────────────────────────────────────────────┐
│                        CLI Layer                             │
│  (Command parsing, user interaction, progress display)      │
└─────────────────────┬───────────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────────┐
│                  Translation Engine                          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │   Batch      │  │  Key Filter  │  │   RTL        │     │
│  │  Processor   │  │   Matcher    │  │  Processor   │     │
│  └──────────────┘  └──────────────┘  └──────────────┘     │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │  Reflection  │  │  Incremental │  │   Format     │     │
│  │   Engine     │  │  Translator  │  │  Protector   │     │
│  └──────────────┘  └──────────────┘  └──────────────┘     │
└─────────────────────┬───────────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────────┐
│              Terminology Manager                             │
│  (Automatic detection, dictionary management)               │
└─────────────────────┬───────────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────────┐
│                  AI Provider Layer                           │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │   OpenAI     │  │  Anthropic   │  │   Google     │     │
│  │   Provider   │  │   Provider   │  │   Provider   │     │
│  └──────────────┘  └──────────────┘  └──────────────┘     │
└─────────────────────────────────────────────────────────────┘
```

### Translation Workflow

1. **Load & Analyze**: Load source JSON, detect changes (incremental mode)
2. **Terminology**: Auto-detect or load terminology dictionary
3. **Filter**: Apply key filters if specified
4. **Batch**: Split into batches for efficient processing
5. **Translate**: Send to AI provider with format instructions
6. **Reflect** ⭐: Quality check and selective improvement (Agentic!)
7. **Process RTL**: Apply bidirectional text handling if needed
8. **Merge**: Combine with unchanged translations
9. **Save**: Write final output with pretty formatting

## 💡 Examples

### Example 1: First-time Translation

```bash
$ jta en.json --to zh

📄 Loading source file...
✓ Source file loaded

📚 Loading terminology...
🔍 Detecting terminology...
✓ Detected 8 terms

🤖 Translating...
✓ Translation completed

💾 Saving translation...
✓ Saved to zh.json

📊 Translation Statistics
   Total items     100
   Success         100
   Failed          0
   Duration        30s
   API calls       5 (includes 1 reflection call)
```

**Generated `.jta-terminology.json`:**
```json
{
  "source_language": "en",
  "preserve_terms": ["GitHub", "API", "OAuth"],
  "consistent_terms": {
    "en": ["repository", "commit", "pull request"]
  }
}
```

### Example 2: Incremental Update

```bash
$ jta en.json --to zh

📄 Loading source file...
✓ Source file loaded

🔍 Analyzing changes...
   New: 5 keys
   Modified: 2 keys
   Unchanged: 93 keys

Continue? [Y/n] y

🤖 Translating...
✓ Translation completed

📊 Translation Statistics
   Total items     7
   Success         7
   Filtered        93 included, 0 excluded (of 100 total)
   Duration        3s
   API calls       1
```

### Example 3: Key Filtering

```bash
# Translate only settings and user sections
$ jta en.json --to ja --keys "settings.**,user.**"

📊 Translation Statistics
   Filtered        45 included, 55 excluded (of 100 total)
   Total items     45
   Success         45
```

### Example 4: Multi-language Batch

```bash
# Translate to multiple languages at once
$ jta en.json --to zh,ja,ko,es,fr -y

Processing: zh ━━━━━━━━━━━━━━━━━━━━ 100% (100/100) ✓
Processing: ja ━━━━━━━━━━━━━━━━━━━━ 100% (100/100) ✓
Processing: ko ━━━━━━━━━━━━━━━━━━━━ 100% (100/100) ✓
Processing: es ━━━━━━━━━━━━━━━━━━━━ 100% (100/100) ✓
Processing: fr ━━━━━━━━━━━━━━━━━━━━ 100% (100/100) ✓

✓ Successfully created 5 translation files
```

### Example 5: CI/CD Integration

```yaml
# .github/workflows/translate.yml
name: Auto-translate i18n files

on:
  push:
    paths:
      - 'locales/en.json'

jobs:
  translate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Install Jta
        run: go install github.com/hikanner/jta/cmd/jta@latest
      
      - name: Translate
        env:
          OPENAI_API_KEY: ${{ secrets.OPENAI_API_KEY }}
        run: |
          jta locales/en.json --to zh,ja,ko -y
      
      - name: Commit translations
        run: |
          git config user.name "Translation Bot"
          git config user.email "bot@example.com"
          git add locales/*.json
          git commit -m "chore: update translations" || exit 0
          git push
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

## 🔧 Troubleshooting

### Common Issues

#### API Key Not Found

```bash
Error: OPENAI_API_KEY environment variable not set
```

**Solution**: Set the API key as an environment variable or pass it directly:
```bash
export OPENAI_API_KEY=sk-...
# Or
jta en.json --to zh --api-key sk-...
```

#### Translation Quality Issues

If translations are not meeting quality expectations:

1. **Use a better model**: GPT-4o > GPT-4 Turbo > GPT-3.5
   ```bash
   jta en.json --to zh --model gpt-4o
   ```

2. **Check terminology**: Review and refine `.jta-terminology.json`
   ```json
   {
     "preserve_terms": ["YourBrand", "ProductName"],
     "consistent_terms": {
       "en": ["important", "domain", "terms"]
     }
   }
   ```

3. **Enable reflection**: The reflection mechanism should auto-fix issues, but verify it's running:
   ```bash
   jta en.json --to zh --verbose
   ```

#### Format Elements Lost in Translation

The format protector should automatically preserve placeholders, but if you notice issues:

1. Check the format instructions in verbose mode
2. Verify your placeholders follow standard patterns: `{var}`, `{{var}}`, `%s`, `%d`
3. Report non-standard formats as issues

#### Rate Limit Errors

```bash
Error: Rate limit exceeded
```

**Solution**: Reduce concurrency and batch size:
```bash
jta en.json --to zh --concurrency 1 --batch-size 10
```

#### Large File Handling

For files with 1000+ keys:

```bash
# Process in smaller batches with lower concurrency
jta large.json --to zh --batch-size 10 --concurrency 2

# Or filter by sections
jta large.json --to zh --keys "section1.**"
jta large.json --to zh --keys "section2.**"
```

### Performance Tips

1. **Batch Size**: Larger batches (20-50) are more efficient but use more tokens per request
2. **Concurrency**: Higher concurrency (3-5) speeds up translation but may hit rate limits
3. **Incremental Mode**: Always use incremental translation for updates (automatic)
4. **Provider Selection**: 
   - OpenAI GPT-4o: Best quality, moderate speed
   - Anthropic Claude 3.5: Great quality, good speed
   - Google Gemini: Experimental, fastest but may need more reflection passes

### Debug Mode

Enable verbose output to see detailed execution:

```bash
jta en.json --to zh --verbose

# You'll see:
# - Provider initialization
# - Batch processing details
# - Reflection engine decisions
# - API call statistics
# - Format validation reports
```

## ❓ FAQ

**Q: How much does it cost to translate a typical i18n file?**

A: For a 100-key file using OpenAI GPT-4o:
- First translation: ~$0.05-0.10 (depending on content length)
- Incremental updates: ~$0.01-0.02 (only new/modified keys)
- The reflection mechanism adds ~20-50% to cost but significantly improves quality

**Q: Can I translate offline or use my own models?**

A: Currently, Jta requires an internet connection and uses cloud AI providers. Local model support is planned for a future release.

**Q: Does Jta support variables inside translated strings?**

A: Yes! All standard placeholder formats are automatically preserved:
- `{variable}`, `{{count}}` (i18next, Vue I18n)
- `%s`, `%d`, `%(name)s` (printf-style)
- `<b>`, `<span>` (HTML tags)

**Q: How do I handle custom terminology?**

A: Edit `.jta-terminology.json` manually:
```json
{
  "source_language": "en",
  "preserve_terms": ["MyApp", "SpecialFeature"],
  "consistent_terms": {
    "en": ["user", "account", "settings"]
  }
}
```

Then run translation with `--skip-terms` to use your custom dictionary.

**Q: Can I review translations before saving?**

A: Currently, translations are saved automatically. For manual review:
1. Use `--output` to save to a separate file
2. Review and edit the output
3. Copy to your actual locale file when satisfied

**Q: What languages are supported?**

A: Jta supports any language that your chosen AI provider supports. Common languages include:
- European: EN, ES, FR, DE, IT, PT, NL, PL, RU
- Asian: ZH, JA, KO, TH, VI, ID
- Middle Eastern: AR, HE, FA (with RTL support)
- And many more...

**Q: How is this different from other translation tools?**

A: Jta's Agentic approach with reflection mechanism sets it apart:
1. **Self-optimizing**: Automatically checks and improves translation quality
2. **Context-aware**: Understands domain terminology and preserves it
3. **Format-safe**: Never breaks your placeholders or markup
4. **Cost-efficient**: Incremental translation saves 80-90% on updates
5. **Production-ready**: Built with Go for reliability and performance

## 🤝 Contributing

Contributions are welcome! Please read our [Contributing Guide](CONTRIBUTING.md) for details.

### Development Setup

```bash
# Clone the repository
git clone https://github.com/hikanner/jta.git
cd jta

# Install dependencies
go mod download

# Run tests
go test ./...

# Build
go build -o jta cmd/jta/main.go

# Run locally
./jta examples/en.json --to zh
```

## 📄 License

MIT License - see [LICENSE](LICENSE) for details.

## 🙏 Acknowledgments

- Inspired by [Andrew Ng's Translation Agent](https://github.com/andrewyng/translation-agent)
- Built with official AI provider SDKs:
  - [OpenAI Go SDK](https://github.com/openai/openai-go)
  - [Anthropic Go SDK](https://github.com/anthropics/anthropic-sdk-go)
  - [Google GenAI Go SDK](https://github.com/google/generative-ai-go)
- Powered by:
  - [Cobra](https://github.com/spf13/cobra) for CLI
  - [Lipgloss](https://github.com/charmbracelet/lipgloss) for beautiful terminal output
  - [Sonic](https://github.com/bytedance/sonic) for fast JSON parsing

## 📞 Support

- 🐛 **Bug Reports**: [GitHub Issues](https://github.com/hikanner/jta/issues)
- 💬 **Discussions**: [GitHub Discussions](https://github.com/hikanner/jta/discussions)
- 📖 **Documentation**: [Wiki](https://github.com/hikanner/jta/wiki)
- ⭐ **Star us**: If you find Jta useful, give us a star on GitHub!

## 🗺️ Roadmap

- [ ] Support for local/self-hosted LLMs
- [ ] Interactive translation review mode
- [ ] Translation memory (TMX) integration
- [ ] Custom prompt templates
- [ ] Web UI for terminology management
- [ ] Support for additional file formats (YAML, XML, PO)
- [ ] Translation statistics and analytics
- [ ] A/B testing for translation quality

---

**Made with ❤️ by the Jta team**

*Jta - Making i18n translation intelligent, reliable, and effortless.*
