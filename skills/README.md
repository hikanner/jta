# Jta Agent Skills

This directory contains Agent Skills that enable AI agents (like Claude) to use Jta for automatic JSON translation.

## What are Agent Skills?

[Agent Skills](https://docs.anthropic.com/docs/agents-and-tools/agent-skills) are modular capabilities that extend what AI agents can do. They consist of instructions, scripts, and resources that help agents perform specialized tasks autonomously.

## Available Skills

### jta

Enables AI agents to automatically translate JSON i18n files with Agentic quality optimization.

**Location:** [jta/](jta/)

**Core Files:**
- [SKILL.md](jta/SKILL.md) - Complete skill definition and instructions for AI agents
- [examples/](jta/examples/) - Step-by-step use cases
- [scripts/](jta/scripts/) - Helper scripts (installation, etc.)

## Installation

### For Individual Users

Copy the skill to your Claude skills directory:

```bash
# From the jta repository root
cp -r skills/jta ~/.claude/skills/

# Or create a symbolic link (recommended for development)
ln -s $(pwd)/skills/jta ~/.claude/skills/jta
```

### For Project Teams

Add the skill to your project's `.claude/skills/` directory:

```bash
# In your project root
mkdir -p .claude/skills
cp -r /path/to/jta/skills/jta .claude/skills/

# Commit to version control
git add .claude/skills/jta
git commit -m "feat: add jta translation skill"
```

Team members will automatically get the skill when they clone the repository.

## Usage

Once installed, simply ask your AI agent:

### Basic Translation
> "Translate my en.json to Chinese, Japanese, and Korean"

### Incremental Updates
> "I added new keys to en.json, please update the translations"

### CI/CD Integration
> "Set up automatic translation in GitHub Actions"

The agent will:
1. Check if Jta is installed (install if needed)
2. Verify API key configuration
3. Execute translation with optimal settings
4. Validate results
5. Report statistics and output locations

## Requirements

### Jta Binary

The skill requires the `jta` command-line tool. The agent will automatically install it if not present, or you can install manually:

```bash
# macOS/Linux via Homebrew
brew tap hikanner/jta
brew install jta

# Or download binary
curl -L https://github.com/hikanner/jta/releases/latest/download/jta-darwin-arm64 -o jta
chmod +x jta
sudo mv jta /usr/local/bin/
```

### API Key

You need an API key from one of these providers:
- **OpenAI**: `OPENAI_API_KEY` (recommended: GPT-5)
- **Anthropic**: `ANTHROPIC_API_KEY` (recommended: Claude Sonnet 4.5)
- **Google Gemini**: `GEMINI_API_KEY` (recommended: Gemini 2.5 Flash)

Set as environment variable:
```bash
export OPENAI_API_KEY=sk-...
```

## Features

The jta skill provides:

- ✅ **Agentic Translation**: AI translates, evaluates, and improves its own work (3x API calls)
- ✅ **Smart Terminology**: Automatic detection and consistent translation of domain terms
- ✅ **Format Protection**: Preserves `{variables}`, HTML tags, URLs, Markdown
- ✅ **Incremental Mode**: Only translates new/changed content (saves 80-90% cost)
- ✅ **27 Languages**: Including RTL languages (Arabic, Hebrew, Persian, Urdu)
- ✅ **CI/CD Ready**: Automatic workflow generation for GitHub Actions
- ✅ **Error Handling**: Comprehensive error detection and user guidance

## Examples

### Example 1: First-time Translation

**User request:**
> "Translate my locales/en.json to Chinese and Japanese"

**Agent actions:**
1. Locates `locales/en.json`
2. Checks Jta installation
3. Verifies API key
4. Executes: `jta locales/en.json --to zh,ja -y`
5. Shows results and statistics

**Output:**
- `locales/zh.json` (Chinese)
- `locales/ja.json` (Japanese)
- `.jta/terminology.json` (detected terms)

### Example 2: Incremental Update

**User request:**
> "I updated en.json with 5 new keys, update the translations"

**Agent actions:**
1. Detects existing translations
2. Uses incremental mode: `jta locales/en.json --to zh,ja --incremental -y`
3. Translates only new/changed keys
4. Preserves existing translations

**Result:** 90% cost savings, faster execution

### Example 3: CI/CD Setup

**User request:**
> "Set up automatic translation in our CI"

**Agent actions:**
1. Creates `.github/workflows/translate-i18n.yml`
2. Configures incremental mode
3. Sets up commit automation
4. Documents API key setup

**Result:** Fully automated translation pipeline

## How It Works

### Skill Activation

The agent automatically activates the skill when you:
- Mention "translate", "translation", "i18n", "internationalization", "locale"
- Reference JSON files in translation contexts
- Specify target languages

### Workflow

```
1. User Request
   ↓
2. Skill Activation (automatic)
   ↓
3. Environment Check (install if needed)
   ↓
4. Execute Translation
   ↓
5. Verify Results
   ↓
6. Report to User
```

### Progressive Disclosure

The skill uses a three-level information architecture:

1. **SKILL.md** (~600 lines): Core instructions, always loaded when relevant
2. **reference.md** (~1000 lines): Detailed docs, loaded on-demand
3. **examples/** (~500 lines each): Specific use cases, loaded as needed

This ensures the agent has enough context without overwhelming the context window.

## Troubleshooting

### Skill Not Activating

Ensure your request includes keywords like:
- "translate JSON"
- "i18n", "internationalization", "locale"
- Specific languages ("Chinese", "Japanese", etc.)

### Installation Issues

If Jta installation fails:
```bash
# Manual installation
curl -L https://github.com/hikanner/jta/releases/latest/download/jta-darwin-arm64 -o jta
chmod +x jta
sudo mv jta /usr/local/bin/
```

### API Key Not Found

Set your API key:
```bash
export OPENAI_API_KEY=sk-...
# Or
export ANTHROPIC_API_KEY=sk-ant-...
# Or
export GEMINI_API_KEY=...
```

## Best Practices

### For Development

1. **Use symbolic links** when developing skills:
   ```bash
   ln -s $(pwd)/skills/jta ~/.claude/skills/jta
   ```

2. **Test changes immediately** - modifications to SKILL.md take effect on next agent interaction

3. **Check allowed-tools** - ensure your skill only uses necessary tools

### For Production

1. **Always use incremental mode** for updates to save costs

2. **Set up CI/CD** for automated translations

3. **Version control terminology** - commit `.jta/` directory for consistency

4. **Monitor API usage** - track costs and optimize batch sizes

## Contributing

To improve the jta skill:

1. Edit files in `skills/jta/`
2. Test with Claude or another agent
3. Submit a pull request to the main Jta repository

## Support

- 📖 **Skill Documentation**: [jta/SKILL.md](jta/SKILL.md)
- 💡 **Examples**: [jta/examples/](jta/examples/)
- 🐛 **Issues**: [GitHub Issues](https://github.com/hikanner/jta/issues)
- 💬 **Discussions**: [GitHub Discussions](https://github.com/hikanner/jta/discussions)

## Related Documentation

- [Agent Skills Overview](https://docs.anthropic.com/docs/agents-and-tools/agent-skills)
- [Agent Skills Best Practices](https://docs.anthropic.com/docs/agents-and-tools/agent-skills/best-practices)
- [Jta Main Documentation](../README.md)

---

**Made for the Agent Skills ecosystem**

*Enabling AI agents to handle internationalization automatically.*
