# Example: Basic Translation

This example demonstrates the most common use case: translating a JSON i18n file to multiple languages.

## Scenario

User has `locales/en.json` containing English translations and wants to add Chinese, Japanese, and Korean support.

## User Request

> "Translate my en.json to Chinese, Japanese, and Korean"

## Claude's Workflow

### Step 1: Understand the Request

Claude recognizes:
- **Task**: Translation
- **Source file**: en.json (implied to be in locales/)
- **Target languages**: Chinese (zh), Japanese (ja), Korean (ko)
- **Mode**: First-time translation (full, not incremental)

### Step 2: Locate the Source File

```bash
# Find JSON files in locale directories
find . -type f -name "en.json" \( -path "*/locales/*" -o -path "*/locale/*" -o -path "*/i18n/*" \)
```

Output:
```
./locales/en.json
```

### Step 3: Check Environment

```bash
# Verify jta is installed
jta --version
```

If not installed:
```bash
# Install via Homebrew (macOS)
brew tap hikanner/jta && brew install jta
```

Check API key:
```bash
# Check if API key exists
if [[ -n "$OPENAI_API_KEY" ]]; then
  echo "✓ API key found"
fi
```

### Step 4: Preview Source File

```bash
# Show file structure
jq '.' locales/en.json | head -20
```

Output:
```json
{
  "app": {
    "name": "My App",
    "welcome": "Welcome to {appName}!"
  },
  "settings": {
    "title": "Settings",
    "language": "Language"
  }
}
```

### Step 5: Execute Translation

```bash
# Translate to Chinese, Japanese, and Korean
jta locales/en.json --to zh,ja,ko -y
```

**What happens:**
1. **Terminology Detection**: Jta analyzes content and detects important terms
2. **Batch Processing**: Splits into batches (default 20 keys per batch)
3. **Translation**: Sends each batch to AI provider
4. **Agentic Reflection**: AI evaluates and improves translations
5. **Output**: Creates zh.json, ja.json, ko.json

### Step 6: Monitor Progress

Console output:
```
📄 Loading source file...
✓ Source file loaded (100 keys)

📚 Loading terminology...
🔍 Detecting terminology...
✓ Detected 8 terms (3 preserve, 5 consistent)

🤖 Translating to Chinese (zh)...
Processing ━━━━━━━━━━━━━━━━━━━━ 100% (100/100)
✓ Chinese translation completed

🤖 Translating to Japanese (ja)...
Processing ━━━━━━━━━━━━━━━━━━━━ 100% (100/100)
✓ Japanese translation completed

🤖 Translating to Korean (ko)...
Processing ━━━━━━━━━━━━━━━━━━━━ 100% (100/100)
✓ Korean translation completed

💾 Saving translations...
✓ Saved to locales/zh.json
✓ Saved to locales/ja.json
✓ Saved to locales/ko.json
```

### Step 7: Verify Results

```bash
# Check created files
ls -lh locales/*.json
```

Output:
```
-rw-r--r--  1 user  staff   2.3K  en.json
-rw-r--r--  1 user  staff   2.8K  zh.json
-rw-r--r--  1 user  staff   3.1K  ja.json
-rw-r--r--  1 user  staff   2.9K  ko.json
```

Validate JSON structure:
```bash
for file in locales/{zh,ja,ko}.json; do
  if jq empty "$file" 2>/dev/null; then
    echo "✓ $file is valid JSON"
  fi
done
```

### Step 8: Show Statistics

```
📊 Translation Statistics

Chinese (zh):
  Total items:     100
  Success:         100
  Failed:          0
  Duration:        45s
  API calls:       15 (5 translate + 5 reflect + 5 improve)

Japanese (ja):
  Total items:     100
  Success:         100
  Failed:          0
  Duration:        48s
  API calls:       15

Korean (ko):
  Total items:     100
  Success:         100
  Failed:          0
  Duration:        43s
  API calls:       15

Total Duration:    2m 16s
Total API calls:   45
Estimated cost:    ~$0.90 (using GPT-4o)

Terminology saved to .jta/ directory for future consistency.
```

## Claude's Response to User

```
✅ Translation completed successfully!

I've translated your en.json to Chinese, Japanese, and Korean.

**Files created:**
- `locales/zh.json` (Chinese Simplified)
- `locales/ja.json` (Japanese)
- `locales/ko.json` (Korean)

**Statistics:**
- Total items: 100 per language
- Success rate: 100%
- Duration: 2 minutes 16 seconds
- API calls: 45 total (15 per language with Agentic reflection)
- Estimated cost: ~$0.90

**Terminology management:**
Jta detected 8 important terms and saved them to `.jta/` directory:
- Preserve terms (never translate): API, OAuth, JSON
- Consistent terms (translate uniformly): credits, workspace, settings, profile, template

All placeholders ({appName}), HTML tags, and URLs have been preserved in the translations.

Would you like me to show you a sample of the translations or make any adjustments?
```

## Sample Translations

### English (source)
```json
{
  "app": {
    "welcome": "Welcome to {appName}!",
    "description": "Manage your <b>workspace</b> settings"
  },
  "user": {
    "credits": "You have {count} credits remaining"
  }
}
```

### Chinese (zh.json)
```json
{
  "app": {
    "welcome": "欢迎来到 {appName}!",
    "description": "管理您的<b>工作空间</b>设置"
  },
  "user": {
    "credits": "您还有 {count} 积分"
  }
}
```

### Japanese (ja.json)
```json
{
  "app": {
    "welcome": "{appName}へようこそ!",
    "description": "<b>ワークスペース</b>の設定を管理"
  },
  "user": {
    "credits": "残り {count} クレジットがあります"
  }
}
```

### Korean (ko.json)
```json
{
  "app": {
    "welcome": "{appName}에 오신 것을 환영합니다!",
    "description": "<b>워크스페이스</b> 설정 관리"
  },
  "user": {
    "credits": "{count}개의 크레딧이 남아 있습니다"
  }
}
```

## Notes

- All placeholders (`{appName}`, `{count}`) preserved
- HTML tags (`<b>`) preserved
- Terminology consistency maintained ("workspace" → "工作空间", "ワークスペース", "워크스페이스")
- Agentic reflection ensured natural, fluent translations
