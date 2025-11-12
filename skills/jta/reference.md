# Jta Translation Skill - Detailed Reference

This document provides comprehensive technical details about the Jta translation skill.

## Table of Contents

- [How Jta Works](#how-jta-works)
- [Agentic Reflection Mechanism](#agentic-reflection-mechanism)
- [Terminology Management System](#terminology-management-system)
- [Format Protection](#format-protection)
- [Incremental Translation](#incremental-translation)
- [Supported Languages](#supported-languages)
- [AI Providers](#ai-providers)
- [Command-Line Reference](#command-line-reference)
- [API Cost Estimation](#api-cost-estimation)
- [Troubleshooting](#troubleshooting)

## How Jta Works

Jta follows a multi-phase translation workflow:

```
Phase 1: Preparation
  ├─ Load source JSON
  ├─ Detect/load terminology
  ├─ Apply key filters (if specified)
  └─ Create batches

Phase 2: Translation
  ├─ Translate batches (parallel)
  └─ Return initial translations

Phase 3: Agentic Reflection ⭐
  ├─ Step 1: LLM evaluates quality (accuracy, fluency, style, terminology)
  ├─ Step 2: LLM generates improvement suggestions
  └─ Step 3: LLM applies improvements

Phase 4: Finalization
  ├─ Process RTL (if needed)
  ├─ Merge results
  └─ Save output
```

## Agentic Reflection Mechanism

Jta implements a three-step Agentic reflection process inspired by Andrew Ng's Translation Agent:

### Step 1: Initial Translation (1 API call)

```
Input: "Welcome to {app_name}"
↓ LLM Translation
Output: "欢迎使用 {app_name}"
```

### Step 2: Quality Reflection (1 API call)

The AI acts as an expert reviewer, evaluating its own translation:

```
Evaluation Dimensions:
(i)   Accuracy: No errors, mistranslations, or omissions?
(ii)  Fluency: Natural grammar and punctuation?
(iii) Style: Appropriate tone and cultural context?
(iv)  Terminology: Consistent and correct domain terms?

AI Self-Critique Example:
"[welcome.message] The translation '欢迎使用 {app_name}' is accurate
but could be more natural. Consider '欢迎来到' which conveys a warmer,
more inviting tone that better matches 'Welcome to'."
```

### Step 3: Self-Improvement (1 API call)

The AI refines the translation based on its expert analysis:

```
AI Improvement:
Original: Welcome to {app_name}
Initial Translation: 欢迎使用 {app_name}
Suggestion: Use '欢迎来到' for warmer tone
↓
Improved Translation: 欢迎来到 {app_name}
```

### Why Agentic Reflection Works

1. **AI as Expert Reviewer**: The same AI that translated understands the context and nuances
2. **Beyond Static Rules**: Dynamically identifies issues specific to each translation
3. **Contextual Improvements**: Generates tailored suggestions, not generic fixes
4. **Iterative Quality**: Complete review-and-refine cycle catches subtle issues

### Cost Structure

- **Total API calls per batch**: 3x (translate → reflect → improve)
- **Example**: 100 keys with batch-size 20 = 15 API calls (5 + 5 + 5)
- **Trade-off**: 3x API cost for significantly higher quality

## Terminology Management System

### Automatic Detection

Jta uses LLM to identify two types of terms:

1. **Preserve Terms**: Never translate (e.g., API, OAuth, GitHub)
2. **Consistent Terms**: Translate uniformly (e.g., credits → 积分)

### File Structure

```
.jta/
├── terminology.json       # Source language terms
├── terminology.zh.json    # Chinese translations
├── terminology.ja.json    # Japanese translations
└── terminology.ko.json    # Korean translations
```

### terminology.json Format

```json
{
  "version": "1.0",
  "sourceLanguage": "en",
  "detectedAt": "2025-01-26T10:30:00Z",
  "preserveTerms": ["API", "OAuth", "JSON", "GitHub"],
  "consistentTerms": ["credits", "workspace", "prompt", "template"]
}
```

### terminology.{lang}.json Format

```json
{
  "version": "1.0",
  "sourceLanguage": "en",
  "targetLanguage": "zh",
  "translatedAt": "2025-01-26T10:31:00Z",
  "translations": {
    "credits": "积分",
    "workspace": "工作空间",
    "prompt": "提示词",
    "template": "模板"
  }
}
```

### Workflow

1. **First run**: Detects terms → saves to `terminology.json` → translates to target language
2. **Subsequent runs**: Loads existing terms → translates missing terms only
3. **New language**: Uses existing `terminology.json` → creates new `terminology.{lang}.json`

### Custom Terminology

Users can manually edit terminology files:

```json
{
  "version": "1.0",
  "sourceLanguage": "en",
  "preserveTerms": ["MyBrand", "ProductName", "API"],
  "consistentTerms": ["feature", "dashboard", "analytics"]
}
```

Then run with `--skip-terminology` to use custom terms:

```bash
jta en.json --to zh --skip-terminology
```

## Format Protection

Jta automatically detects and preserves various format elements:

### Variables and Placeholders

- i18next style: `{variable}`, `{{count}}`
- printf style: `%s`, `%d`, `%(name)s`, `%(count)d`
- ICU MessageFormat: `{name, number}`, `{count, plural, one{#} other{#}}`

### HTML Tags

- Simple tags: `<b>`, `<i>`, `<span>`
- Tags with attributes: `<span class="highlight">`, `<a href="...">`
- Self-closing tags: `<br/>`, `<img src="..."/>`

### URLs

- HTTP/HTTPS: `https://example.com`, `http://api.example.com/v1`
- Relative: `/path/to/resource`
- With query params: `https://example.com?key=value&lang=en`

### Markdown

- Bold: `**text**`, `__text__`
- Italic: `*text*`, `_text_`
- Links: `[text](url)`
- Code: `` `code` ``

### Examples

```json
{
  "welcome": "Welcome to {appName}!",
  "description": "We have <b>{count}</b> users",
  "link": "Visit <a href=\"https://example.com\">our site</a>",
  "markdown": "Read the **documentation** for more info",
  "printf": "Hello %(name)s, you have %d credits"
}
```

All format elements are preserved in the translated output.

## Incremental Translation

### How It Works

Jta uses content hashing to detect changes:

```
Phase 1: Load existing translation
Phase 2: Compare with source
  ├─ New keys: Not in previous translation
  ├─ Modified keys: Content hash changed
  ├─ Unchanged keys: Content hash same
  └─ Deleted keys: In previous but not in source
Phase 3: Translate only new + modified
Phase 4: Merge with unchanged translations
Phase 5: Remove deleted keys
```

### Usage

```bash
# First time: Full translation
jta en.json --to zh

# After updates: Incremental translation
jta en.json --to zh --incremental
```

### Benefits

- **Cost savings**: 80-90% reduction on updates
- **Time savings**: Only processes changed content
- **Quality preservation**: Keeps existing high-quality translations

### Example Output

```
🔍 Analyzing changes...
   New: 5 keys
   Modified: 2 keys
   Unchanged: 93 keys
   Deleted: 1 key

Continue? [Y/n] y

🤖 Translating 7 items...
✓ Translation completed
```

## Supported Languages

### Complete List (27 Languages)

#### Left-to-Right (LTR)

**European:**
- English (en)
- Spanish (es)
- French (fr)
- German (de)
- Italian (it)
- Portuguese (pt)
- Russian (ru)
- Dutch (nl)
- Polish (pl)
- Turkish (tr)

**Asian:**
- Chinese Simplified (zh)
- Chinese Traditional (zh-TW)
- Japanese (ja)
- Korean (ko)
- Thai (th)
- Vietnamese (vi)
- Indonesian (id)
- Malay (ms)
- Hindi (hi)
- Bengali (bn)
- Sinhala (si)
- Nepali (ne)
- Burmese (my)

#### Right-to-Left (RTL)

**Middle Eastern:**
- Arabic (ar)
- Persian (fa)
- Hebrew (he)
- Urdu (ur)

### RTL Language Support

Special handling for Right-to-Left languages:

1. **Automatic bidirectional markers**: `\u202B` (RTL embedding), `\u202C` (pop directional formatting)
2. **Smart punctuation conversion**: For Arabic-script languages
3. **Proper LTR content embedding**: URLs, numbers, code in RTL context

Example:

```json
{
  "welcome": "مرحبا بك في {appName}!",
  "link": "زر <a href=\"https://example.com\">موقعنا</a>"
}
```

### View All Languages

```bash
jta --list-languages
```

Output includes:
- Language code
- Native name
- English name
- Flag emoji
- Script type (LTR/RTL)

## AI Providers

### OpenAI

**Environment Variable:** `OPENAI_API_KEY`

**Default Model:** `gpt-5`

**Available Models:**
- GPT-5 (latest, most capable)
- GPT-5 mini (faster, cost-effective)
- GPT-5 nano (fastest, lowest cost)
- GPT-4o (previous generation)
- GPT-3.5 Turbo (fast, economical)

**Usage:**
```bash
export OPENAI_API_KEY=sk-...
jta en.json --to zh

# Or specify model
jta en.json --to zh --provider openai --model gpt-5-mini
```

### Anthropic

**Environment Variable:** `ANTHROPIC_API_KEY`

**Default Model:** `claude-sonnet-4-5`

**Available Models:**
- Claude Sonnet 4.5 (balanced, recommended)
- Claude Opus 4.1 (highest quality)
- Claude Haiku 4.5 (fastest)

**Usage:**
```bash
export ANTHROPIC_API_KEY=sk-ant-...
jta en.json --to zh --provider anthropic

# For highest quality
jta en.json --to zh --provider anthropic --model claude-opus-4-1
```

### Google Gemini

**Environment Variable:** `GEMINI_API_KEY`

**Default Model:** `gemini-2.5-flash`

**Available Models:**
- Gemini 2.5 Pro (highest quality)
- Gemini 2.5 Flash (fast, cost-effective)

**Usage:**
```bash
export GEMINI_API_KEY=...
jta en.json --to zh --provider gemini

# For highest quality
jta en.json --to zh --provider gemini --model gemini-2.5-pro
```

## Command-Line Reference

### Basic Options

```bash
jta <source> --to <languages>

Arguments:
  <source>              Source JSON file path (required)

Flags:
  --to string          Target language(s), comma-separated (required)
                       Example: zh,ja,ko
```

### AI Provider Options

```bash
  --provider string    AI provider (openai, anthropic, gemini)
                       Default: openai

  --model string       Model name
                       Default: Provider-specific default

  --api-key string     API key (or use environment variable)
```

### Output Options

```bash
  -o, --output string  Output file path
                       Default: <target-lang>.json in source directory
```

### Translation Options

```bash
  --incremental        Incremental translation (only new/modified)

  --keys string        Only translate specified keys (glob patterns)
                       Example: "settings.*,user.*"

  --exclude-keys       Exclude specified keys (glob patterns)
                       Example: "admin.*,internal.*"

  --batch-size int     Batch size for translation (default: 20)

  --concurrency int    Concurrency for batch processing (default: 3)
```

### Terminology Options

```bash
  --terminology-dir    Terminology directory (default: .jta/)

  --skip-terminology   Skip term detection (use existing terminology)

  --no-terminology     Disable terminology management completely

  --redetect-terms     Re-detect terminology (when source language changes)
```

### Other Options

```bash
  --source-lang        Source language (auto-detected from filename)

  -y, --yes            Non-interactive mode

  -v, --verbose        Verbose output

  --list-languages     List all supported languages and exit

  --version            Print version information and exit
```

### Complete Example

```bash
jta locales/en.json \
  --to zh,ja,ko \
  --provider anthropic \
  --model claude-sonnet-4-5 \
  --incremental \
  --keys "settings.*,user.*" \
  --exclude-keys "internal.*" \
  --batch-size 15 \
  --concurrency 2 \
  --terminology-dir ../shared-terms/ \
  --output ./locales/ \
  --verbose \
  -y
```

## API Cost Estimation

### Cost Structure

For 100-key file with Agentic reflection:

**OpenAI GPT-4o:**
- First translation: ~$0.15-0.30
- Incremental update (10 keys): ~$0.03-0.06
- Without reflection: ~$0.05-0.10

**Anthropic Claude Sonnet 4.5:**
- First translation: ~$0.20-0.40
- Incremental update (10 keys): ~$0.04-0.08
- Without reflection: ~$0.07-0.13

**Google Gemini 2.5 Flash:**
- First translation: ~$0.10-0.20
- Incremental update (10 keys): ~$0.02-0.04
- Without reflection: ~$0.03-0.07

### Cost Optimization Tips

1. **Use incremental mode** for updates (saves 80-90%)
2. **Larger batch sizes** reduce total API calls
3. **Choose provider wisely**: Gemini Flash for cost, Claude for quality
4. **Filter keys** if only translating specific sections

### Example Calculation

100-key file, batch-size 20, with reflection:
- Batches: 100 ÷ 20 = 5 batches
- API calls per batch: 3 (translate + reflect + improve)
- Total API calls: 5 × 3 = 15 calls

## Troubleshooting

### Installation Issues

**Problem:** `brew install jta` fails

**Solution:**
```bash
# Try manual installation
OS="$(uname -s)"
if [[ "$OS" == "Darwin" ]]; then
  curl -L https://github.com/hikanner/jta/releases/latest/download/jta-darwin-arm64 -o jta
else
  curl -L https://github.com/hikanner/jta/releases/latest/download/jta-linux-amd64 -o jta
fi
chmod +x jta
sudo mv jta /usr/local/bin/
```

### API Issues

**Problem:** Rate limit exceeded

**Solution:**
```bash
# Reduce batch size and concurrency
jta en.json --to zh --batch-size 10 --concurrency 1

# Add delays between requests (wait a minute, then retry)
```

**Problem:** API key not found

**Solution:**
```bash
# Set environment variable
export OPENAI_API_KEY=sk-...

# Or pass directly
jta en.json --to zh --api-key sk-...
```

### Translation Quality Issues

**Problem:** Translations are not natural

**Solution:**
1. Use a better model:
   ```bash
   jta en.json --to zh --provider anthropic --model claude-sonnet-4-5
   ```

2. Check and edit terminology:
   ```bash
   vim .jta/terminology.json
   vim .jta/terminology.zh.json
   ```

3. Verify Agentic reflection is working:
   ```bash
   jta en.json --to zh --verbose
   # Look for: "Step 2: Reflection", "Step 3: Improvement"
   ```

### Format Issues

**Problem:** Variables or HTML tags lost in translation

**Solution:**
1. Verify format protection is working:
   ```bash
   jta en.json --to zh --verbose
   ```

2. Check if placeholders follow standard patterns:
   - Supported: `{var}`, `{{var}}`, `%s`, `%d`
   - Not supported: Custom placeholder formats

3. Report non-standard formats as issues

### File Issues

**Problem:** Invalid JSON error

**Solution:**
```bash
# Validate source file
jq . source.json

# Fix JSON syntax errors
```

**Problem:** Output file not created

**Solution:**
```bash
# Check permissions
ls -la output-directory/

# Specify explicit output path
jta en.json --to zh --output ./locales/zh.json
```

### Performance Issues

**Problem:** Translation is slow

**Solution:**
```bash
# For large files (>500 keys)
jta large.json --to zh --batch-size 50 --concurrency 5

# Use faster provider
jta en.json --to zh --provider gemini --model gemini-2.5-flash
```

**Problem:** High API costs

**Solution:**
1. Always use incremental mode for updates:
   ```bash
   jta en.json --to zh --incremental
   ```

2. Translate only needed sections:
   ```bash
   jta en.json --to zh --keys "public.*"
   ```

3. Use larger batch sizes:
   ```bash
   jta en.json --to zh --batch-size 50
   ```
