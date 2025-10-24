# Jta - 产品需求文档（PRD）

> AI-powered Agentic JSON Translation

## 📋 项目概述

**项目名称**: Jta (JSON Translation Agent)  
**命令行名称**: `jta`  
**版本**: v1.0  
**目标用户**: 出海项目、独立开发者、前端团队、开源项目维护者  
**核心价值**: AI 驱动的 Agentic JSON Translation 工具，具备智能术语管理、格式保护、质量优化等能力

### 背景与问题

在 Web 应用开发中，多语言国际化是常见需求。开发者通常使用 JSON 文件（如 `en.json`）管理可翻译文案。但将整个文件翻译成其他语言时面临以下挑战：

1. **术语一致性难保证**: 同一个术语（如 "credits"）在不同上下文中可能被翻译成不同的词
2. **格式容易损坏**: 变量占位符 `{variable}` 可能在翻译中丢失或损坏
3. **手动翻译效率低**: 大型项目可能包含数百个文案条目
4. **质量参差不齐**: 简单翻译vs高质量翻译的权衡

### 解决方案

Jta 采用 **Agentic JSON Translation** 方法，让 AI 智能体自主完成 JSON 文件的高质量翻译：

**什么是 Agentic JSON Translation？**

Agentic JSON Translation 是专为 JSON 国际化文件设计的智能翻译方法。AI 不仅执行翻译任务，还能自主决策、优化和管理整个 JSON 翻译过程：

- ✅ **智能术语管理**: 自动检测术语、翻译缺失项、确保一致性
- ✅ **格式智能保护**: 自动识别和保护占位符、HTML、Markdown 等格式
- ✅ **质量自我优化**: 内置轻量级反思机制，持续优化翻译质量
- ✅ **上下文理解**: 批量翻译时保持上下文连贯，避免孤立翻译
- ✅ **自主决策**: 根据内容类型自动调整翻译策略
- ✅ **高效执行**: 智能批量和并发处理，平衡质量和效率

---

## 🎯 设计理念

### 简单至上

**一个命令，AI 智能体自主完成 JSON 翻译的所有工作**:
```bash
jta en.json --to zh
```

Agentic JSON Translation 自动执行：
1. 解析 JSON 文件结构
2. 分析并检测术语
3. 翻译缺失的术语
4. 应用术语进行翻译
5. 自动保护特殊格式
6. 优化翻译质量
7. 输出完整 JSON 文件

### 智能默认（Agentic 特性）

AI 智能体自主决策，无需用户干预：
- 自动检测源语言和目标语言特性
- 自动识别需要保留的术语（品牌名、API 等）
- 自动识别需要统一翻译的术语（高频词汇）
- 自动保护格式（占位符、HTML 等）
- 自动检测 RTL 语言并应用正确处理
- 自动选择合适的批量大小和并发策略

### 用户控制

虽然 AI 智能体自主执行，但用户仍有完全控制权：
- 术语文件可手动编辑（JSON 格式）
- 可跳过术语检测（`--skip-terms`）
- 可完全禁用术语管理（`--no-terminology`）
- 所有配置通过命令行选项指定
- 交互式确认或非交互模式（`-y`）

---

## 🚀 核心功能

### 1. 基础翻译能力

**基本用法**:
```bash
# 翻译到单个语言
jta en.json --to zh

# 翻译到多个语言
jta en.json --to zh,ja,ko

# 指定输出目录
jta en.json --to zh --output ./locales/

# 指定 AI 提供商
jta en.json --to zh --provider anthropic --model claude-3-5-sonnet-20250116

# 强制完整重新翻译（忽略已有翻译）
jta en.json --to zh --force

# 非交互模式（CI/CD 友好）
jta en.json --to zh -y
```

**输入要求**:
- ✅ 支持标准 JSON 格式
- ✅ 支持嵌套结构（无限深度）
- ✅ 支持数组
- ✅ 自动保持原始格式和键的顺序

**输出保证**:
- ✅ 生成对应的目标语言文件（如 `zh.json`）
- ✅ 保持与源文件相同的结构
- ✅ 保留代码格式（缩进、换行）
- ✅ JSON 结构 100% 完整
- ✅ 格式元素（占位符、HTML 等）完整无损

**语言支持**:
- 支持 25+ 种主流语言
- 包括 RTL 语言（阿拉伯语、希伯来语等）
- 任何语言都可以作为源语言

---

### 2. 智能术语管理（Agentic 核心能力）

**Agentic 术语管理的智能特性**：

Jta 的术语管理不是简单的词典映射，而是 AI 智能体主动管理的过程：
- 🤖 **自动检测**：使用 LLM 分析上下文，智能识别术语（非简单频率统计）
- 🤖 **自动翻译**：发现缺失的术语翻译时，自动翻译并更新
- 🤖 **自动应用**：翻译过程中自动应用术语，确保一致性
- 👤 **人工编辑**：用户可随时手动编辑术语文件
- 🔄 **持续优化**：每次翻译都会检查和更新术语

**术语类型**:

1. **保留术语（Preserve Terms）**: 永不翻译
   - 品牌名: `MyApp`, `OpenAI`
   - 技术术语: `API`, `OAuth`, `JSON`
   - 产品名: `FLUX.1`, `GPT-4`
   - 自动检测: 全大写词、专有名词模式

2. **一致性术语（Consistent Terms）**: 统一翻译
   - 高频业务词汇: `credits`, `premium`, `generation`
   - 自动检测: 高频出现（通过 LLM 分析上下文判断）
   - 需要翻译为各目标语言

**自动化流程**:

**场景 1: 首次翻译（自动检测术语）**
```bash
$ jta en.json --to zh

🔍 检查术语文件...
❌ 未找到 .jta-terminology.json

🤖 正在分析源文件术语...
✨ 检测到 3 个保留术语: MyApp, API, FLUX.1
✨ 检测到 5 个一致性术语: credits, premium, generation, model, background
   
   保存术语文件到 .jta-terminology.json? [Y/n]
   > Y

✅ 术语文件已保存

🔍 检查术语翻译...
⚠️  发现 5 个术语缺少中文翻译

🤖 正在翻译术语...
   credits → 点数
   premium → 高级版
   generation → 生成
   model → 模型
   background → 背景

✅ 术语翻译完成

🚀 开始翻译主文件...
   [████████████████████] 100%

✅ 翻译完成: zh.json
```

**场景 2: 术语文件已存在（用户手动编辑过）**
```bash
$ jta en.json --to ja

🔍 检查术语文件...
✅ 找到 .jta-terminology.json

💡 是否检测新术语并合并到现有术语文件？[Y/n]
> n  # 用户选择跳过检测

🔍 检查术语翻译...
⚠️  发现 3 个术语缺少日语翻译

🤖 正在翻译术语...
   credits → クレジット
   premium → プレミアム
   generation → 生成

✅ 术语翻译完成

🚀 开始翻译主文件...
   [████████████████████] 100%

✅ 翻译完成: ja.json
```

**场景 3: 跳过术语检测（手动管理术语）**
```bash
$ jta en.json --to ko --skip-terms

🔍 检查术语文件...
✅ 找到 .jta-terminology.json
⏩ 跳过术语检测（--skip-terms）

🔍 检查术语翻译...
⚠️  发现 2 个术语缺少韩语翻译

🤖 正在翻译术语...
   credits → 크레딧
   premium → 프리미엄

✅ 术语翻译完成

🚀 开始翻译主文件...
   [████████████████████] 100%

✅ 翻译完成: ko.json
```

**场景 4: 完全不使用术语管理**
```bash
$ jta en.json --to zh --no-terminology

⏩ 跳过术语管理（--no-terminology）

🚀 开始翻译主文件...
   [████████████████████] 100%

✅ 翻译完成: zh.json
```

**术语文件格式** (`.jta-terminology.json`):
```json
{
  "sourceLanguage": "en",
  "preserveTerms": [
    "MyApp",
    "API",
    "FLUX.1",
    "OAuth"
  ],
  "consistentTerms": {
    "en": [
      "credits",
      "premium",
      "generation",
      "model"
    ],
    "zh": [
      "点数",
      "高级版",
      "生成",
      "模型"
    ],
    "ja": [
      "クレジット",
      "プレミアム",
      "生成",
      "モデル"
    ]
  }
}
```

**手动管理**:
- 用户可以直接编辑 `.jta-terminology.json` 文件
- 添加/删除术语
- 修改术语翻译
- 添加新语言的术语翻译

**术语管理选项**:
```bash
# 跳过术语检测（不分析新术语，但仍会翻译缺失的术语）
# 适合用户手动编辑术语文件的场景
jta en.json --to zh --skip-terms

# 完全不使用术语管理（不检测、不加载、不保护）
# 适合不需要术语一致性的场景
jta en.json --to zh --no-terminology
```

**LLM 智能检测**:
- 不使用简单的频率阈值
- 通过上下文理解术语的重要性
- 识别哪些词需要保留原文
- 识别哪些词需要统一翻译
- 提供检测理由（可通过 `--verbose` 查看）

---

### 3. Agentic 翻译机制

**核心能力**:

Jta 的 Agentic 翻译不是简单的文本替换，而是智能体主动管理的翻译过程：

🤖 **上下文理解**: 批量翻译时保持上下文连贯，避免孤立翻译  
🤖 **质量自我优化**: 内置轻量级反思机制，自动检查并改进翻译质量  
🤖 **术语自动应用**: 翻译过程中自动应用术语词典，确保一致性  
🤖 **格式智能处理**: 自动识别和保护特殊格式，无需用户配置  
🤖 **策略自主选择**: 根据内容类型自动调整翻译策略

**反思机制优化**:

Jta 借鉴了 Andrew Ng Translation Agent 的反思方法，但进行了轻量化优化：

1. **初始翻译**: 使用术语词典进行首次翻译
2. **轻量反思**: 快速检查翻译质量和术语一致性
3. **选择性改进**: 只对需要改进的部分进行重译

**与传统方法的对比**:

| 维度 | 传统批量翻译 | Andrew Ng 完整反思 | Jta Agentic 方案 |
|------|------------|------------------|------------------|
| API 调用 | 1x | 3x | 1.2-1.5x |
| 翻译质量 | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ |
| 术语一致性 | ❌ 不保证 | ⚠️ 依赖 LLM | ✅ 强制保证 |
| 成本 | 低 | 高 | 中 |
| 速度 | 快 | 慢 | 中 |
| 适用场景 | 简单文本 | 高价值内容 | 通用 |

**实现方式**:
- 批量翻译时在 prompt 中加入反思指导
- 对检测到的潜在问题进行二次翻译
- 智能平衡质量和效率
- 无需用户选择模式，AI 智能体自动优化

---

### 4. 智能格式保护（Agentic 能力）

**AI 智能体自动保护**:

格式保护是 Jta 的重要 Agentic 能力之一。AI 智能体自动识别、保护和验证各种格式元素：

🤖 **自动识别**: 智能检测文本中的特殊格式，无需用户配置  
🤖 **自动保护**: 翻译过程中确保格式元素完整无损  
🤖 **自动验证**: 翻译完成后自动验证格式完整性  
🤖 **自动修复**: 发现格式问题时自动尝试修复

**支持的格式类型**:

| 格式类型 | 示例 | 智能处理方式 |
|---------|------|------------|
| 变量占位符 | `{variable}`, `{{count}}`, `%s` | 自动识别并保持原样 |
| HTML 标签 | `<b>text</b>`, `<span>` | 自动保持标签完整 |
| 自定义标记 | `[highlight]...[/highlight]` | 自动识别并保护 |
| URL | `https://example.com` | 自动完整保留 |
| Markdown | `**bold**`, `*italic*` | 自动保持语法 |
| 技术参数 | `16:9`, `1920x1080` | 自动原样保留 |

**智能验证机制**:
```bash
✅ 格式验证:
  - 占位符: 45/45 ✓
  - HTML 标签: 12/12 ✓
  - URL: 3/3 ✓
  
⚠️  发现 1 个格式问题:
  - settings.message: 缺少占位符 {name}
  🤖 正在自动修复...
```

**示例**:
```json
// 原文
{
  "message": "You have {count} credits",
  "description": "Use **FLUX.1** model for <b>best</b> results"
}

// 译文（中文）
{
  "message": "您有 {count} 个点数",
  "description": "使用 **FLUX.1** 模型获得<b>最佳</b>效果"
}
```

---

### 5. 智能 RTL 语言支持（Agentic 能力）

**AI 智能体自动处理**:

RTL（从右到左）语言支持是 Jta 的另一个 Agentic 能力。AI 智能体自动识别语言特性并应用正确处理：

🤖 **自动检测**: 智能识别目标语言是否为 RTL（阿拉伯语、希伯来语、波斯语、乌尔都语）  
🤖 **自动处理**: 根据语言特性自动应用正确的文本方向和格式  
🤖 **智能混排**: 正确处理 RTL 文本中的 LTR 元素（品牌名、数字等）  
🤖 **标点转换**: 自动使用目标语言的标点符号

**智能处理项**:

1. **方向标记**: 自动为 LTR 文本（英文单词、品牌名）添加方向标记
   ```
   原文: "Use FLUX.1 model"
   阿拉伯语: "استخدم نموذج ‎FLUX.1‎"  // ‎...‎ 确保 FLUX.1 从左到右显示
   ```

2. **标点符号转换**: 智能转换为目标语言的标点符号
   ```
   ? → ؟ (阿拉伯语问号)
   , → ، (阿拉伯语逗号)
   ```

3. **数字格式**: 智能保持西文数字（0-9），因为更通用
   ```
   "123 credits" → "١٢٣ من الرصيد"  // 123 保持不变
   ```

**自动化示例**:
```bash
# AI 智能体自动检测并处理阿拉伯语
jta en.json --to ar

# 翻译多个语言（包括 RTL）
jta en.json --to zh,ar,ja,he
# zh 和 ja 使用 LTR 处理，ar 和 he 自动使用 RTL 处理
```

---

### 6. 智能增量翻译（Agentic 能力）

**AI 智能体自动处理**:

增量翻译是 Jta 的核心 Agentic 能力之一。AI 智能体自动检测文件变更，只翻译需要翻译的部分：

🤖 **自动检测**: 智能对比源文件和目标文件，识别差异
🤖 **智能决策**: 自动判断哪些内容需要翻译，哪些可以保留
🤖 **效率优化**: 只翻译新增和修改的内容，节省时间和成本
🤖 **完整性保证**: 自动删除源文件中已不存在的 key

**工作原理**:

当目标文件已存在时，Jta 会：
1. **对比结构**: 分析源文件和目标文件的差异
2. **检测变更**: 识别新增、修改、删除的 key
3. **增量翻译**: 只翻译新增和修改的 key
4. **智能合并**: 保留未变更的翻译，删除多余的 key

**自动化示例**:

**首次翻译**:
```bash
$ jta en.json --to zh

🔍 检查目标文件...
❌ 未找到 zh.json

🚀 开始完整翻译...
   [████████████████████] 100% (100 keys)

✅ 翻译完成: zh.json (100 keys)
   💰 API 费用: ~$0.50
   ⏱️  耗时: 30 秒
```

**第二次（增量更新）**:
```bash
$ jta en.json --to zh

🔍 检查目标文件...
✅ 找到 zh.json

📊 分析差异:
  ✨ 新增: 5 keys
  🔄 修改: 2 keys (源文本变更)
  🗑️  删除: 3 keys (源文件中已删除)
  ✅ 保持: 90 keys (无变化)

💡 将增量翻译 7 keys，保留 90 个已有翻译。
   继续？[Y/n]
   > Y

🤖 正在翻译 7 keys...
   [████████████████████] 100%

✅ 增量翻译完成: zh.json
   - 翻译: 7 keys
   - 保留: 90 keys
   - 删除: 3 keys
   💰 API 费用: ~$0.05 (节省 90%)
   ⏱️  耗时: 3 秒
```

**强制完整重新翻译**:
```bash
$ jta en.json --to zh --force

⚠️  将忽略已有翻译，完整重新翻译 100 keys。
   继续？[Y/n]
   > Y

🚀 开始完整翻译...
   [████████████████████] 100%

✅ 完整翻译完成: zh.json (100 keys)
```

**多语言增量更新**:
```bash
$ jta en.json --to zh,ja,ko

🔍 检查目标文件...

[中文]
  ✅ 找到 zh.json
  📊 增量: 7 keys | 保持: 90 keys

[日语]
  ✅ 找到 ja.json
  📊 增量: 7 keys | 保持: 90 keys

[韩语]
  ❌ 未找到 ko.json
  📊 完整翻译: 97 keys

💡 将翻译 21 keys (zh: 7, ja: 7, ko: 97)。
   继续？[Y/n]
   > Y

✅ 所有翻译完成
   💰 总费用: ~$0.60 (增量节省 50%)
```

**差异检测规则**:

| 情况 | 行为 | 示例 |
|------|------|------|
| **源文件新增 key** | 翻译该 key | `"new": "New text"` |
| **源文本修改** | 重新翻译该 key | `"hello": "Hi"` → `"hello": "Hello"` |
| **目标文件多余 key** | 删除该 key | 源文件删除后，目标文件也删除 |
| **未变更** | 保留现有翻译 | 不调用 API |

**智能保留用户修改**:

如果用户手动修改了目标文件的翻译，增量翻译会：
- ✅ 保留手动修改（如果源文本未变更）
- 🔄 重新翻译（如果源文本已变更）

**选项控制**:
```bash
# 默认：智能增量翻译
jta en.json --to zh

# 强制完整重新翻译
jta en.json --to zh --force

# 非交互模式（自动确认）
jta en.json --to zh -y
```

**优势对比**:

| 维度 | 完整翻译 | 增量翻译 |
|------|---------|---------|
| **速度** | 慢（需翻译所有） | 快（只翻译差异） |
| **成本** | 高（每次全量） | 低（按需翻译） |
| **保留手动修改** | ❌ 会覆盖 | ✅ 会保留 |
| **适用场景** | 首次翻译 | 持续更新 |

---

### 7. 指定 Key 翻译

**精确控制翻译范围**:

在实际项目中，有时只需要翻译特定区域或排除某些区域，Jta 提供灵活的 key 过滤功能：

**使用场景**:

| 场景 | 需求 | 方案 |
|------|------|------|
| **翻译特定区域** | 只翻译 settings 和 user 相关 | `--keys "settings.*,user.*"` |
| **排除敏感区域** | 不翻译 admin 和 internal | `--exclude-keys "admin.*,internal.*"` |
| **精确指定** | 只翻译特定几个 key | `--keys "app.title,home.welcome"` |
| **组合使用** | 增量翻译特定区域 | `--keys "settings.*"` + 自动增量 |

**通配符支持**:

```bash
settings.*         # settings 下的所有一级 key
settings.**        # settings 下的所有 key（递归）
*.title           # 所有名为 title 的 key
settings.*.desc   # settings 下所有对象的 desc key
```

**使用示例**:

**示例 1：只翻译指定区域**
```bash
$ jta en.json --to zh --keys "settings.*,user.profile.*"

🔍 检查翻译范围...
✅ 匹配到 12 keys:
  - settings.title
  - settings.description
  - settings.theme.*
  - user.profile.name
  - user.profile.bio
  ...

💡 将翻译 12 keys (其他 88 keys 保持不变)。
   继续？[Y/n]
   > Y

🤖 正在翻译...
✅ 完成: 12 keys 已翻译
```

**示例 2：排除特定区域**
```bash
$ jta en.json --to zh --exclude-keys "admin.*,internal.*,debug.*"

🔍 检查翻译范围...
⏭️  排除 15 keys:
  - admin.* (8 keys)
  - internal.* (5 keys)
  - debug.* (2 keys)

💡 将翻译 85 keys (排除 15 keys)。
   继续？[Y/n]
   > Y

🤖 正在翻译...
✅ 完成: 85 keys 已翻译
```

**示例 3：组合增量翻译**
```bash
$ jta en.json --to zh --keys "settings.*"

🔍 检查目标文件...
✅ 找到 zh.json

🔍 检查翻译范围...
✅ 匹配到 10 keys (settings.*)

📊 分析差异（仅限 settings.* keys）:
  ✨ 新增: 2 keys
  🔄 修改: 1 key
  ✅ 保持: 7 keys
  ⏭️  跳过: 90 keys (其他区域)

💡 将增量翻译 3 keys。
   继续？[Y/n]
   > Y

✅ 增量翻译完成
   - 翻译: 3 keys (settings.*)
   - 保留: 7 keys (settings.*)
   - 跳过: 90 keys (其他区域)
```

**示例 4：强制重新翻译指定区域**
```bash
$ jta en.json --to zh --keys "settings.*" --force

🔍 检查翻译范围...
✅ 匹配到 10 keys (settings.*)

⚠️  将完整重新翻译 settings.* 区域 (10 keys)。
   继续？[Y/n]
   > Y

✅ 完成: 10 keys 已重新翻译
```

**JSON 结构示例**:

```json
// en.json
{
  "settings": {
    "title": "Settings",
    "theme": {
      "light": "Light Mode",
      "dark": "Dark Mode"
    }
  },
  "user": {
    "profile": {
      "name": "Name",
      "bio": "Bio"
    }
  },
  "admin": {
    "panel": "Admin Panel"
  },
  "internal": {
    "debug": "Debug Mode"
  }
}
```

**不同选项的翻译结果**:

| 选项 | 翻译的 keys | 保持原样的 keys |
|------|-----------|--------------|
| `--keys "settings.*"` | settings.title, settings.theme.* | user.*, admin.*, internal.* |
| `--keys "settings.**"` | settings 下所有 key（递归） | user.*, admin.*, internal.* |
| `--keys "*.title"` | settings.title, (其他 .title) | 其他所有 |
| `--exclude-keys "admin.*,internal.*"` | settings.*, user.* | admin.*, internal.* |

**优先级规则**:

```bash
# 1. --keys 和 --exclude-keys 同时使用
jta en.json --to zh --keys "settings.*" --exclude-keys "settings.advanced.*"
# 结果：翻译 settings.* 但排除 settings.advanced.*

# 2. 优先级：--exclude-keys > --keys
# 先应用 --keys 过滤，再应用 --exclude-keys 排除
```

---

## 🛠 CLI 设计

### 主命令

```bash
jta <source> --to <languages> [options]
```

> 注：核心功能就是翻译，所以直接使用 `jta` 命令，无需子命令

### 核心选项

**AI 提供商配置**:
```bash
--provider string       AI 提供商 (openai, anthropic, google)
--model string         模型名称
--api-key string       API Key (或使用环境变量)
--base-url string      自定义 API 端点
```

**术语管理**:
```bash
--terminology string   术语文件路径 (默认: .jta-terminology.json)
--skip-terms          跳过术语检测（不分析新术语，但仍会翻译缺失的术语）
--no-terminology      完全不使用术语管理（不检测、不加载、不保护）
```

**翻译范围控制**:
```bash
--keys string         只翻译指定的 key，逗号分隔，支持通配符
                      示例: --keys "settings.*,user.profile.*"
--exclude-keys string 排除指定的 key，逗号分隔，支持通配符
                      示例: --exclude-keys "admin.*,internal.*"
```

**输出控制**:
```bash
--output, -o string   输出文件或目录
--force              强制完整重新翻译（忽略已有翻译，跳过增量检测）
--overwrite          覆盖现有文件（不询问）
```

**性能调优**:
```bash
--batch-size int     批次大小 (默认: 20)
--concurrency int    并发数 (默认: 3)
```

**交互模式**:
```bash
--yes, -y           非交互模式（自动确认所有操作）
```

**调试选项**:
```bash
--verbose, -v       详细输出
--debug            调试模式
```

### 使用示例

**基础使用**:
```bash
# 最简单的用法
jta en.json --to zh

# 多语言
jta en.json --to zh,ja,ko,es

# 指定输出目录
jta en.json --to zh -o ./locales/
```

**配置 AI 提供商**:
```bash
# 使用环境变量（推荐）
export OPENAI_API_KEY=sk-...
jta en.json --to zh

# 命令行指定
jta en.json --to zh \
  --provider anthropic \
  --model claude-3-5-sonnet-20250116 \
  --api-key sk-ant-...

# 使用自定义端点
jta en.json --to zh \
  --provider openai \
  --base-url https://api.openai-proxy.com/v1
```

**术语控制**:
```bash
# 使用自定义术语文件
jta en.json --to zh --terminology ./my-terms.json

# 跳过术语检测（适合手动管理术语文件）
# 注：仍会翻译缺失的术语
jta en.json --to zh --skip-terms

# 完全不使用术语管理
jta en.json --to zh --no-terminology
```

**翻译范围控制**:
```bash
# 只翻译指定区域
jta en.json --to zh --keys "settings.*,user.*"

# 排除特定区域
jta en.json --to zh --exclude-keys "admin.*,internal.*"

# 组合增量翻译
jta en.json --to zh --keys "settings.*"  # 只增量翻译 settings 区域

# 强制重新翻译指定区域
jta en.json --to zh --keys "settings.*" --force
```

**CI/CD 集成**:
```bash
# 非交互模式
jta en.json --to zh,ja,ko -y

# 完整 CI 命令
ANTHROPIC_API_KEY=${{ secrets.ANTHROPIC_API_KEY }} \
  jta en.json --to zh,ja,ko \
  --output ./locales \
  --provider anthropic \
  -y
```

---

## 📦 安装方式

### 方式 1: Homebrew（macOS/Linux 推荐）

```bash
brew tap hikanner/jta
brew install jta
```

### 方式 2: Install Script（快速安装）

```bash
# macOS/Linux
curl -fsSL https://raw.githubusercontent.com/hikanner/jta/main/install.sh | bash

# 或使用 wget
wget -qO- https://raw.githubusercontent.com/hikanner/jta/main/install.sh | bash

# Windows (PowerShell)
iwr https://raw.githubusercontent.com/hikanner/jta/main/install.ps1 -useb | iex
```

### 方式 3: Go Install

```bash
go install github.com/hikanner/jta/cmd/jta@latest
```

### 方式 4: 下载二进制文件

从 [GitHub Releases](https://github.com/hikanner/jta/releases) 下载对应平台的二进制文件：

- `jta-darwin-amd64` (macOS Intel)
- `jta-darwin-arm64` (macOS Apple Silicon)
- `jta-linux-amd64` (Linux)
- `jta-windows-amd64.exe` (Windows)

解压后移动到 PATH 目录：
```bash
# macOS/Linux
sudo mv jta /usr/local/bin/

# Windows
# 移动到 C:\Program Files\jta\jta.exe
# 并添加到 PATH 环境变量
```

### 验证安装

```bash
jta --version
# 输出: jta version 1.0.0
```

---

## 🎯 支持的 AI 提供商

| 提供商 | 环境变量 | 推荐模型 |
|--------|---------|---------|
| **OpenAI** | `OPENAI_API_KEY` | `gpt-4o`, `gpt-4-turbo` |
| **Anthropic** | `ANTHROPIC_API_KEY` | `claude-3-5-sonnet-20250116` |
| **Google** | `GEMINI_API_KEY` | `gemini-2.0-flash-exp` |

**使用环境变量（推荐）**:
```bash
# 添加到 ~/.bashrc 或 ~/.zshrc
export OPENAI_API_KEY=sk-...
export ANTHROPIC_API_KEY=sk-ant-...
export GEMINI_API_KEY=...
```

**或创建 .env 文件**:
```bash
# .env
OPENAI_API_KEY=sk-...
ANTHROPIC_API_KEY=sk-ant-...
GEMINI_API_KEY=...
```

---

## 📊 输出与反馈

### 进度显示

```bash
🚀 开始翻译到中文(简体)...

📊 准备阶段:
  ✅ 加载源文件: en.json (423 keys)
  ✅ 加载术语: 8 个
  ✅ 创建翻译批次: 22 个

⚡ 翻译阶段:
  [████████████████████] 100% (22/22 批次)

✅ 翻译完成！
  📝 翻译: 420/423 成功
  ⏱️  耗时: 2 分 15 秒
  💰 API 调用: 24 次

🔍 验证:
  ✅ JSON 结构: 100%
  ✅ 术语一致性: 100%
  ✅ 格式完整性: 100%

📁 输出: ./zh.json
```

### 简化报告

```bash
✅ 翻译完成

统计:
  - 翻译成功: 420/423 (99.3%)
  - 耗时: 2 分 15 秒
  - API 调用: 24 次

术语一致性:
  - credits: 点数 (15 次) ✓
  - premium: 高级版 (8 次) ✓
  - generation: 生成 (12 次) ✓

保留术语:
  - MyApp (15 次) ✓
  - API (23 次) ✓

⚠️  3 个翻译失败 (已保留原文):
  - settings.advanced.api_key
  - profile.bio_template
  - admin.permissions

输出文件: zh.json
```

---

## 🌟 与同类工具对比

| 特性 | Jta | Google Translate | i18n-ally | 其他 AI 翻译工具 |
|------|-----|-----------------|-----------|----------------|
| **翻译方法** | 🤖 **Agentic JSON Translation** | 传统机器翻译 | 编辑器插件 | LLM 翻译 |
| **JSON 专用** | ✅ 专为 JSON 设计 | ❌ 通用文本翻译 | ⚠️ 多格式支持 | ❌ 通用文本翻译 |
| **结构保持** | ✅ 完美保持 JSON 结构 | ❌ 不理解结构 | ✅ 保持 | ⚠️ 有限 |
| **术语管理** | ✅ AI 智能体自动管理 | ❌ 无 | ⚠️ 手动管理 | ❌ 依赖 LLM 记忆 |
| **格式保护** | ✅ AI 智能识别保护 | ❌ 经常损坏 | ⚠️ 有限支持 | ⚠️ 有限支持 |
| **翻译质量** | ⭐⭐⭐⭐ (自我优化) | ⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐⭐ |
| **批量处理** | ✅ 智能并发 | ❌ 单个文本 | ✅ 支持 | ⚠️ 有限 |
| **RTL 支持** | ✅ AI 自动处理 | ⚠️ 基础支持 | ❌ 无 | ❌ 无 |
| **质量优化** | ✅ 内置反思机制 | ❌ 无 | ❌ 无 | ⚠️ 可选（多次调用） |
| **自主决策** | ✅ AI 智能体自主执行 | ❌ 无 | ❌ 需手动配置 | ⚠️ 有限 |
| **单一命令** | ✅ 一个命令完成所有 | N/A | ❌ 复杂操作 | ⚠️ 多步骤 |
| **成本效率** | 🤖 智能优化（中） | 免费/付费 | 免费 | 高（多次调用） |

---

## 🎯 使用场景

### 场景 1: 独立开发者

**需求**: 快速将 SaaS 产品国际化，预算有限

**方案**:
```bash
# 一键翻译到所有主要语言
jta en.json --to zh,ja,ko,es,fr,de -y

# 低成本: 批量处理 + 并发优化
# 高质量: 术语一致性 + 智能反思
```

### 场景 2: 开源项目

**需求**: 社区贡献多语言翻译，需要保持术语一致

**方案**:
```bash
# 1. 项目维护者创建术语文件
jta en.json --to zh  # 自动生成 .jta-terminology.json

# 2. 提交术语文件到 Git
git add .jta-terminology.json
git commit -m "Add terminology for consistent translation"

# 3. 贡献者翻译新语言时使用相同术语
jta en.json --to fr  # 自动使用现有术语
```

### 场景 3: 企业团队

**需求**: CI/CD 自动化翻译，支持多语言发布

**方案**:
```yaml
# .github/workflows/translate.yml
name: Auto Translation

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
        run: |
          curl -fsSL https://raw.githubusercontent.com/hikanner/jta/main/install.sh | bash
      
      - name: Translate
        env:
          ANTHROPIC_API_KEY: ${{ secrets.ANTHROPIC_API_KEY }}
        run: |
          jta locales/en.json \
            --to zh,ja,ko,es,fr,de \
            --output locales \
            -y
      
      - name: Commit translations
        run: |
          git config --local user.email "bot@example.com"
          git config --local user.name "Translation Bot"
          git add locales/
          git commit -m "chore: update translations" || exit 0
          git push
```

---

## 📚 最佳实践

### 1. 术语管理

**推荐工作流**:
1. 首次翻译时让系统自动生成术语文件
2. 审查并手动编辑术语文件（添加/删除术语）
3. 将术语文件提交到版本控制
4. 后续翻译使用 `--skip-terms` 跳过术语检测
   - 注：系统仍会自动翻译缺失的术语到目标语言

**术语文件组织**:
```
project/
├── locales/
│   ├── en.json
│   ├── zh.json
│   └── ja.json
└── .jta-terminology.json  # 提交到 Git
```

### 2. 文件命名约定

**推荐命名**:
- `en.json` - 英语（源语言）
- `zh.json` - 简体中文
- `zh-TW.json` - 繁体中文
- `ja.json` - 日语
- `ko.json` - 韩语

### 3. 批量翻译优化

**调整批次大小和并发数**:
```bash
# 小文件（< 100 keys）：增大批次减少调用
jta en.json --to zh --batch-size 50

# 大文件（> 500 keys）：增加并发提速
jta en.json --to zh --concurrency 5

# API 限流严格：减少并发
jta en.json --to zh --concurrency 1
```

### 4. CI/CD 集成

**使用 secrets 保护 API Key**:
```bash
# 不要在代码中硬编码 API Key
❌ jta en.json --to zh --api-key sk-xxx

# 使用环境变量
✅ OPENAI_API_KEY=${{ secrets.OPENAI_API_KEY }} \
   jta en.json --to zh -y
```

---

## 🚀 发布计划

### v1.0.0 (首次发布)

**核心功能**:
- ✅ 基础翻译（单/多语言）
- ✅ 智能增量翻译（自动检测差异，只翻译变更）
- ✅ 术语管理（自动检测 + 翻译）
- ✅ Agentic 翻译（内置反思机制）
- ✅ 格式保护
- ✅ RTL 语言支持
- ✅ 多 AI 提供商
- ✅ 批量并发处理

**安装方式**:
- ✅ Homebrew
- ✅ Install script
- ✅ Go install
- ✅ Binary releases

**平台支持**:
- ✅ macOS (Intel + Apple Silicon)
- ✅ Linux (amd64 + arm64)
- ✅ Windows (amd64)

### v1.1.0 (增强功能)

**计划功能**:
- 配置文件支持（可选 .jtarc）
- 更多 AI 提供商（Groq、DeepSeek 等）
- 翻译质量评分
- 自定义 prompt 模板

### v2.0.0 (高级功能)

**计划功能**:
- 交互式审查模式
- 翻译质量报告
- 自定义翻译规则
- 插件系统

---

## 💡 常见问题

### Q: Jta 和原来的 jsontrans 有什么区别？

**A**: 主要区别：
1. **语言**: Python → Golang (单一二进制，更快)
2. **定位**: 重新定位为 Agentic JSON Translation 工具
3. **术语**: AI 智能体自动管理，无需单独的命令
4. **配置**: 无需配置文件，全部通过 CLI 选项

### Q: 为什么使用 LLM 检测术语而不是频率阈值？

**A**: 频率阈值无法理解上下文：
- ❌ 高频但不重要的词会被误判
- ❌ 低频但关键的术语会被遗漏
- ✅ LLM 可以理解语义和上下文
- ✅ 更准确地识别品牌名、技术术语

### Q: 如何在没有术语文件的情况下使用？

**A**: 三种方式：
```bash
# 1. 让系统自动生成（推荐）
jta en.json --to zh  # 会提示创建术语文件

# 2. 跳过术语检测，但仍翻译缺失的术语
jta en.json --to zh --skip-terms

# 3. 完全不使用术语管理
jta en.json --to zh --no-terminology
```

### Q: 支持哪些 JSON 格式？

**A**: 支持所有标准 JSON，包括：
- 嵌套对象（无限深度）
- 数组
- 混合类型
- Unicode 字符

### Q: 翻译质量如何保证？

**A**: 多层保障：
1. 术语一致性强制保证
2. 格式完整性自动验证
3. 内置轻量反思机制
4. 批量翻译时的上下文保持

### Q: API 成本大概多少？

**A**: 示例估算（500 个文本，每个平均 20 词）：
- OpenAI GPT-4o: ~$0.50-1.00
- Anthropic Claude: ~$0.40-0.80
- Google Gemini: ~$0.20-0.40

批量处理和并发优化可以显著降低成本。

---

## 📄 许可证

MIT License - 开源自由使用

---

## 🤝 贡献

欢迎贡献！请查看 [CONTRIBUTING.md](CONTRIBUTING.md)

---

**文档版本**: v1.0  
**最后更新**: 2025-10-21  
**项目地址**: https://github.com/hikanner/jta
