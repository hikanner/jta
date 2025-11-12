# Example: Incremental Translation

This example demonstrates how to update existing translations when you add new content to your source file.

## Scenario

User has already translated `en.json` to Chinese (`zh.json`). Now they've added 5 new keys and modified 2 existing keys in `en.json`. They want to update `zh.json` without re-translating everything.

## User Request

> "I added new keys to my en.json, please update the Chinese translation"

## Claude's Workflow

### Step 1: Understand the Context

Claude recognizes:
- **Task**: Update existing translation
- **Source file**: en.json
- **Target**: Chinese (zh.json already exists)
- **Mode**: Incremental (only translate changes)

### Step 2: Check Existing Translation

```bash
# Verify both files exist
ls -lh locales/en.json locales/zh.json
```

Output:
```
-rw-r--r--  1 user  staff   2.8K  en.json (modified 5 minutes ago)
-rw-r--r--  1 user  staff   2.6K  zh.json (created 2 days ago)
```

### Step 3: Execute Incremental Translation

```bash
# Use incremental mode
jta locales/en.json --to zh --incremental
```

### Step 4: Analyze Changes

Console output:
```
📄 Loading source file...
✓ Source file loaded (105 keys)

📄 Loading existing translation...
✓ Existing translation loaded (98 keys)

🔍 Analyzing changes...

Changes detected:
  📝 New keys:       5
  ✏️  Modified keys:  2
  ✅ Unchanged keys: 93
  🗑️  Deleted keys:   1

Breakdown:
  New keys:
    - settings.notification.email
    - settings.notification.push
    - user.preferences.theme
    - user.preferences.language
    - app.features.newFeature

  Modified keys:
    - app.welcome (content changed)
    - user.credits (content changed)

  Deleted keys:
    - app.deprecated (will be removed from translation)

Continue with incremental translation? [Y/n]
```

### Step 5: User Confirms

Since the request included "update", Claude proceeds:
```bash
# Continue (non-interactive with -y would skip prompt)
# User presses Y or Claude proceeds automatically
```

### Step 6: Translate Only Changes

Console output:
```
📚 Loading terminology...
✓ Terminology loaded (8 terms from .jta/)

🤖 Translating 7 items (5 new + 2 modified)...
Processing ━━━━━━━━━━━━━━━━━━━━ 100% (7/7)
✓ Translation completed

🔄 Merging with existing translations...
✓ Merged 93 unchanged translations
✓ Added 5 new translations
✓ Updated 2 modified translations
✓ Removed 1 deleted key

💾 Saving translation...
✓ Saved to locales/zh.json
```

### Step 7: Show Statistics

```
📊 Incremental Translation Statistics

Target: Chinese (zh)
  Total source keys:    105
  New keys translated:  5
  Modified keys:        2
  Unchanged keys:       93 (preserved from previous translation)
  Deleted keys:         1 (removed)

Translation performance:
  Items translated:     7 (5 + 2)
  Success:              7
  Failed:               0
  Duration:             8s
  API calls:            2 (1 translate + 1 reflect)

Cost savings:
  Without incremental:  ~$0.30 (105 keys, 15 API calls)
  With incremental:     ~$0.03 (7 keys, 2 API calls)
  Saved:                ~90% ($0.27)
```

## Claude's Response to User

```
✅ Chinese translation updated successfully!

I've updated your zh.json with the changes from en.json using incremental mode.

**Changes applied:**
- ✨ Added 5 new translations
- ✏️  Updated 2 modified translations
- 🗑️  Removed 1 obsolete translation
- ✅ Preserved 93 unchanged translations

**New translations added:**
1. `settings.notification.email` → "电子邮件通知"
2. `settings.notification.push` → "推送通知"
3. `user.preferences.theme` → "主题偏好"
4. `user.preferences.language` → "语言偏好"
5. `app.features.newFeature` → "新功能"

**Updated translations:**
1. `app.welcome` → "欢迎来到 {appName}！开始您的旅程"
2. `user.credits` → "您的账户有 {count} 积分"

**Performance:**
- Duration: 8 seconds (vs ~45s for full re-translation)
- API calls: 2 (vs 15 for full re-translation)
- Cost: ~$0.03 (vs ~$0.30, saved 90%)

Your existing high-quality translations were preserved, and only the new/changed content was translated with Agentic reflection for quality assurance.

Would you like to update other languages as well?
```

## Comparison: Incremental vs Full Translation

### Full Translation (without --incremental)
```bash
jta en.json --to zh
```
- Translates all 105 keys
- Takes ~45 seconds
- 15 API calls
- Cost: ~$0.30
- May lose existing manual refinements

### Incremental Translation (with --incremental)
```bash
jta en.json --to zh --incremental
```
- Translates only 7 changed keys
- Takes ~8 seconds
- 2 API calls
- Cost: ~$0.03
- Preserves existing translations exactly

## How Incremental Detection Works

Jta uses content hashing to detect changes:

```
1. Load source (en.json)
2. Load existing translation (zh.json)
3. For each key in source:
   - Calculate content hash of source value
   - Compare with hash stored in previous translation
   - If different → mark as "modified"
   - If not in previous → mark as "new"
4. For each key in previous translation:
   - If not in source → mark as "deleted"
5. Translate only "new" + "modified"
6. Merge with "unchanged"
7. Remove "deleted"
```

## Sample Changes

### Before (en.json - 2 days ago)
```json
{
  "app": {
    "welcome": "Welcome to {appName}!",
    "deprecated": "Old feature (no longer used)"
  },
  "user": {
    "credits": "You have {count} credits"
  }
}
```

### After (en.json - now)
```json
{
  "app": {
    "welcome": "Welcome to {appName}! Start your journey",
    "features": {
      "newFeature": "New Feature"
    }
  },
  "user": {
    "credits": "Your account has {count} credits",
    "preferences": {
      "theme": "Theme",
      "language": "Language"
    }
  },
  "settings": {
    "notification": {
      "email": "Email notifications",
      "push": "Push notifications"
    }
  }
}
```

### Changes Detected
- ✨ **New**: `app.features.newFeature`, `user.preferences.theme`, `user.preferences.language`, `settings.notification.email`, `settings.notification.push`
- ✏️  **Modified**: `app.welcome`, `user.credits`
- 🗑️  **Deleted**: `app.deprecated`

### Updated Translation (zh.json)
```json
{
  "app": {
    "welcome": "欢迎来到 {appName}！开始您的旅程",
    "features": {
      "newFeature": "新功能"
    }
  },
  "user": {
    "credits": "您的账户有 {count} 积分",
    "preferences": {
      "theme": "主题偏好",
      "language": "语言偏好"
    }
  },
  "settings": {
    "notification": {
      "email": "电子邮件通知",
      "push": "推送通知"
    }
  }
}
```

## Best Practices

1. **Always use incremental mode for updates**
   ```bash
   jta en.json --to zh --incremental
   ```

2. **Full re-translation when needed**
   - After major terminology changes
   - When quality is unsatisfactory
   - When starting fresh
   ```bash
   jta en.json --to zh  # without --incremental
   ```

3. **CI/CD integration**
   ```bash
   # In your CI pipeline
   jta locales/en.json --to zh,ja,ko --incremental -y
   ```

4. **Multiple languages**
   ```bash
   # Update all languages incrementally
   jta en.json --to zh,ja,ko,es,fr --incremental -y
   ```
