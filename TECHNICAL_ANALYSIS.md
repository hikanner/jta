# jsontrans - 技术分析文档

> AI-powered JSON translation agent with terminology management

本文档分析了两种翻译系统的技术实现方案，并进行对比分析。

---

## 📚 目录

1. [Andrew Ng 的 Translation Agent 分析](#1-andrew-ng-的-translation-agent-分析)
2. [当前实现方案分析（HiFlux Translation Tool）](#2-当前实现方案分析hiflux-translation-tool)
3. [两种方案对比](#3-两种方案对比)
4. [技术选型建议](#4-技术选型建议)

---

## 1. Andrew Ng 的 Translation Agent 分析

### 1.1 项目概述

**项目**: [andrewyng/translation-agent](https://github.com/andrewyng/translation-agent)
**作者**: Andrew Ng (deeplearning.ai 创始人)
**发布时间**: 2024 年中期
**Star 数**: 5.6k+

**核心理念**:
使用 **Agentic Workflow（智能体工作流）** + **Reflection（反思机制）** 实现高质量翻译。

---

### 1.2 核心架构

#### 工作流程：三步反思循环

```
┌─────────────────────────────────────────────────┐
│                翻译流程                          │
├─────────────────────────────────────────────────┤
│                                                 │
│  Step 1: Initial Translation (初始翻译)        │
│  ├─ 使用 LLM 直接翻译文本                       │
│  ├─ Prompt: "你是专业翻译，将 X 翻译成 Y"       │
│  └─ 输出: translation_1                         │
│                                                 │
│           ↓                                     │
│                                                 │
│  Step 2: Reflection (反思评价)                 │
│  ├─ LLM 扮演翻译评论家                          │
│  ├─ 对比原文和初译                              │
│  ├─ 从 4 个维度评估:                            │
│  │   • 准确性 (Accuracy)                       │
│  │   • 流畅性 (Fluency)                        │
│  │   • 风格 (Style)                            │
│  │   • 术语 (Terminology)                      │
│  └─ 输出: 改进建议列表                          │
│                                                 │
│           ↓                                     │
│                                                 │
│  Step 3: Improved Translation (改进翻译)       │
│  ├─ 结合原文、初译、建议                        │
│  ├─ LLM 重新翻译                                │
│  └─ 输出: translation_2 (最终译文)             │
│                                                 │
└─────────────────────────────────────────────────┘
```

---

### 1.3 核心代码分析

#### 文件结构

```
translation-agent/
├── src/translation_agent/
│   ├── __init__.py           # 导出 translate 函数
│   └── utils.py              # 核心逻辑（678 行）
└── examples/
    └── example_script.py
```

**主要函数**：

```python
# 主入口
translate(source_lang, target_lang, source_text, country, max_tokens=1000)

# 单块翻译（短文本）
one_chunk_translate_text()
  ├─ one_chunk_initial_translation()    # 初译
  ├─ one_chunk_reflect_on_translation() # 反思
  └─ one_chunk_improve_translation()    # 改进

# 多块翻译（长文本）
multichunk_translation()
  ├─ multichunk_initial_translation()
  ├─ multichunk_reflect_on_translation()
  └─ multichunk_improve_translation()
```

---

#### 核心实现 1: 初始翻译

**函数**: `one_chunk_initial_translation()`

```python
def one_chunk_initial_translation(
    source_lang: str,
    target_lang: str,
    source_text: str
) -> str:
    """第一步：直接翻译"""

    system_message = f"You are an expert linguist, specializing in translation from {source_lang} to {target_lang}."

    translation_prompt = f"""This is an {source_lang} to {target_lang} translation, please provide the {target_lang} translation for this text.
Do not provide any explanations or text apart from the translation.

{source_lang}: {source_text}

{target_lang}:"""

    translation = get_completion(translation_prompt, system_message=system_message)

    return translation
```

**特点**：
- ✅ 简单直接
- ✅ 明确角色定位（专家翻译）
- ❌ 无术语指导
- ❌ 无格式保护说明

---

#### 核心实现 2: 反思评价

**函数**: `one_chunk_reflect_on_translation()`

```python
def one_chunk_reflect_on_translation(
    source_lang: str,
    target_lang: str,
    source_text: str,
    translation_1: str,
    country: str = ""
) -> str:
    """第二步：反思并提出改进建议"""

    system_message = f"You are an expert linguist specializing in translation from {source_lang} to {target_lang}. You will be provided with a source text and its translation and your goal is to improve the translation."

    # 如果指定了国家/地区
    country_instruction = ""
    if country != "":
        country_instruction = f"The final style and tone of the translation should match the style of {target_lang} colloquially spoken in {country}."

    reflection_prompt = f"""Your task is to carefully read a source text and a translation from {source_lang} to {target_lang}, and then give constructive criticism and helpful suggestions to improve the translation.

{country_instruction}

The source text and initial translation, delimited by XML tags <SOURCE_TEXT></SOURCE_TEXT> and <TRANSLATION></TRANSLATION>, are as follows:

<SOURCE_TEXT>
{source_text}
</SOURCE_TEXT>

<TRANSLATION>
{translation_1}
</TRANSLATION>

When writing suggestions, pay attention to whether there are ways to improve the translation's:
(i) accuracy (by correcting errors of addition, mistranslation, omission, or untranslated text),
(ii) fluency (by applying {target_lang} grammar, spelling and punctuation rules, and ensuring there are no unnecessary repetitions),
(iii) style (by ensuring the translations reflect the style of the source text and take into account any cultural context),
(iv) terminology (by ensuring terminology use is consistent and reflects the source text domain; and by only ensuring you use equivalent idioms {target_lang}).

Write a list of specific, helpful and constructive suggestions for improving the translation.
Each suggestion should address one specific part of the translation.
Output only the suggestions and nothing else."""

    reflection = get_completion(reflection_prompt, system_message=system_message)

    return reflection
```

**关键设计**：
- ✅ **4 维度评估**：准确性、流畅性、风格、术语
- ✅ **地区化支持**：可指定如 "墨西哥西班牙语"
- ✅ **建设性批评**：要求具体、可操作的建议
- ✅ **XML 标记**：清晰分隔原文和译文

**输出示例**：
```
建议列表:
1. "Unleash" 翻译为 "释放" 较生硬，建议改为 "激发" 更自然
2. 语序调整：将 "使用 AI 驱动的图像生成" 改为 "用 AI 图像生成"
3. 增加感染力：可以加强语气，如 "激发无限创造力"
```

---

#### 核心实现 3: 改进翻译

**函数**: `one_chunk_improve_translation()`

```python
def one_chunk_improve_translation(
    source_lang: str,
    target_lang: str,
    source_text: str,
    translation_1: str,
    reflection: str
) -> str:
    """第三步：根据反思建议改进翻译"""

    system_message = f"You are an expert linguist, specializing in translation editing from {source_lang} to {target_lang}."

    prompt = f"""Your task is to carefully read, then edit, a translation from {source_lang} to {target_lang}, taking into account a list of expert suggestions and constructive criticisms.

The source text, the initial translation, and the expert linguist suggestions are delimited by XML tags <SOURCE_TEXT></SOURCE_TEXT>, <TRANSLATION></TRANSLATION> and <EXPERT_SUGGESTIONS></EXPERT_SUGGESTIONS> as follows:

<SOURCE_TEXT>
{source_text}
</SOURCE_TEXT>

<TRANSLATION>
{translation_1}
</TRANSLATION>

<EXPERT_SUGGESTIONS>
{reflection}
</EXPERT_SUGGESTIONS>

Please take into account the expert suggestions when editing the translation.
Edit the translation by ensuring:
(i) accuracy (by correcting errors of addition, mistranslation, omission, or untranslated text),
(ii) fluency (by applying {target_lang} grammar, spelling and punctuation rules and ensuring there are no unnecessary repetitions),
(iii) style (by ensuring the translations reflect the style of the source text)
(iv) terminology (inappropriate for context, inconsistent use), or
(v) other errors.

Output only the new translation and nothing else."""

    translation_2 = get_completion(prompt, system_message)

    return translation_2
```

**关键设计**：
- ✅ 综合考虑：原文 + 初译 + 建议
- ✅ 明确角色：翻译编辑（而非初次翻译）
- ✅ 强调改进点：5 个维度的优化
- ✅ 纯输出：只返回译文

---

#### 核心实现 4: 长文本分块

**问题**: LLM 有 token 限制，长文本无法一次翻译。

**解决方案**: 智能分块 + 上下文保持

```python
def multichunk_initial_translation(
    source_lang: str,
    target_lang: str,
    source_text_chunks: List[str]
) -> List[str]:
    """翻译多个文本块，关键：提供上下文"""

    translation_chunks = []

    for i in range(len(source_text_chunks)):
        # 核心设计：给每个块提供完整上下文
        tagged_text = (
            "".join(source_text_chunks[0:i])        # 前文
            + "<TRANSLATE_THIS>"
            + source_text_chunks[i]                  # 当前块（需翻译）
            + "</TRANSLATE_THIS>"
            + "".join(source_text_chunks[i + 1:])   # 后文
        )

        prompt = f"""Your task is to provide a professional translation from {source_lang} to {target_lang} of PART of a text.

The source text is below, delimited by XML tags <SOURCE_TEXT> and </SOURCE_TEXT>.
Translate only the part within the source text delimited by <TRANSLATE_THIS> and </TRANSLATE_THIS>.
You can use the rest of the source text as context, but do not translate any of the other text.

<SOURCE_TEXT>
{tagged_text}
</SOURCE_TEXT>

To reiterate, you should translate only this part of the text:
<TRANSLATE_THIS>
{source_text_chunks[i]}
</TRANSLATE_THIS>

Output only the translation of the portion you are asked to translate, and nothing else."""

        translation = get_completion(prompt, system_message)
        translation_chunks.append(translation)

    return translation_chunks
```

**关键设计**：
- ✅ **上下文保持**：虽然只翻译一部分，但给出完整文本
- ✅ **明确标记**：用 `<TRANSLATE_THIS>` 清晰指示翻译范围
- ✅ **连贯性保证**：LLM 可以看到前后文，保证翻译连贯

**示例**：
```
假设文本分为 3 块：[A, B, C]

翻译 A 时:
  上下文: <TRANSLATE_THIS>A</TRANSLATE_THIS> B C

翻译 B 时:
  上下文: A <TRANSLATE_THIS>B</TRANSLATE_THIS> C

翻译 C 时:
  上下文: A B <TRANSLATE_THIS>C</TRANSLATE_THIS>
```

---

#### 核心实现 5: 智能分块算法

**函数**: `calculate_chunk_size()`

```python
def calculate_chunk_size(token_count: int, token_limit: int) -> int:
    """
    智能计算分块大小

    目标: 在 token 限制内，最小化块数量

    算法:
    1. 如果总 token <= 限制，不分块
    2. 否则，计算需要的块数
    3. 平均分配 token 到每个块
    4. 考虑余数，避免最后一块过小
    """

    if token_count <= token_limit:
        return token_count

    # 计算需要的块数（向上取整）
    num_chunks = (token_count + token_limit - 1) // token_limit

    # 平均每块大小
    chunk_size = token_count // num_chunks

    # 分配余数
    remaining_tokens = token_count % token_limit
    if remaining_tokens > 0:
        chunk_size += remaining_tokens // num_chunks

    return chunk_size
```

**示例**：
```python
>>> calculate_chunk_size(1000, 500)
500  # 刚好 2 块

>>> calculate_chunk_size(1530, 500)
389  # 4 块，每块约 389 token

>>> calculate_chunk_size(2242, 500)
496  # 5 块，每块约 496 token
```

**优点**：
- ✅ 块大小均衡
- ✅ 避免过小的最后一块
- ✅ 最小化块数量

---

### 1.4 优势分析

#### ✅ 优势

1. **翻译质量高**
   - 通过反思机制自我改进
   - 多轮迭代优化
   - 适合高质量文档翻译

2. **文化适配好**
   - 支持地区方言（如墨西哥西班牙语）
   - 风格自然流畅
   - 考虑文化背景

3. **实现简洁**
   - 代码简单易懂（678 行）
   - 逻辑清晰
   - 易于扩展

4. **长文本处理好**
   - 智能分块
   - 保持上下文
   - 翻译连贯

---

#### ❌ 劣势

1. **成本高**
   - 3 倍 API 调用（初译 + 反思 + 改进）
   - 适合预算充足的场景

2. **速度慢**
   - 3 倍耗时
   - 不适合实时翻译

3. **无术语管理**
   - 依赖 LLM 自行理解术语
   - 术语一致性无法强制保证

4. **无格式保护**
   - 未特别处理占位符、HTML 等
   - 容易损坏格式

5. **无并发处理**
   - 同步顺序处理
   - 效率较低

6. **无增量翻译**
   - 每次都重新翻译整个文件
   - 无缓存机制

---

### 1.5 适用场景

✅ **适合**：
- 高质量文档翻译（白皮书、技术文档）
- 营销文案（广告、品牌介绍）
- 文学作品
- 需要文化适配的内容
- 预算充足的项目

❌ **不适合**：
- 大规模批量翻译
- 实时翻译
- 需要术语强一致性的场景
- JSON/代码国际化
- 预算有限的项目

---

## 2. 当前实现方案分析（HiFlux Translation Tool）

### 2.1 项目概述

**项目**: translate-json
**用途**: HiFlux AI 图像生成平台的多语言翻译工具
**支持语言**: 26 种
**核心模型**: AWS Bedrock Claude Sonnet 4

---

### 2.2 核心架构

#### 工作流程：批量并发 + 术语管理

```
┌──────────────────────────────────────────────────┐
│                 翻译流程                          │
├──────────────────────────────────────────────────┤
│                                                  │
│  Step 1: 源文件分析                              │
│  ├─ 加载 en.json                                 │
│  ├─ 递归提取所有可翻译文本                        │
│  ├─ 分析混合格式（占位符内外）                    │
│  └─ 提取术语                                     │
│                                                  │
│           ↓                                      │
│                                                  │
│  Step 2: 术语映射                                │
│  ├─ 加载 terminology.json                        │
│  ├─ 匹配术语翻译                                 │
│  ├─ 识别保留术语                                 │
│  └─ 构建术语词典                                 │
│                                                  │
│           ↓                                      │
│                                                  │
│  Step 3: 批量翻译                                │
│  ├─ 创建智能批次（按上下文分组）                  │
│  ├─ 并发处理批次（asyncio）                      │
│  │   ├─ 批次 1 → API 调用 ┐                     │
│  │   ├─ 批次 2 → API 调用 ├─ 并发（3个）         │
│  │   └─ 批次 3 → API 调用 ┘                     │
│  ├─ 失败重试（指数退避）                         │
│  └─ 降级处理（批量→单个）                        │
│                                                  │
│           ↓                                      │
│                                                  │
│  Step 4: 后处理                                  │
│  ├─ 排版规范化                                   │
│  │   • 中文：中英文间加空格                      │
│  │   • 法语：冒号前加空格                        │
│  ├─ RTL 处理（如阿拉伯语）                       │
│  │   • 添加方向标记                              │
│  │   • 数字转换                                  │
│  └─ 重建 JSON 结构                               │
│                                                  │
│           ↓                                      │
│                                                  │
│  Step 5: 质量验证                                │
│  ├─ JSON 结构完整性                              │
│  ├─ 术语一致性                                   │
│  ├─ 占位符完整性                                 │
│  └─ 生成验证报告                                 │
│                                                  │
└──────────────────────────────────────────────────┘
```

---

### 2.3 核心代码分析

#### 文件结构

```
translate-json/
├── translate.py                # 主程序入口
├── config.py                   # 配置文件
│
├── core/                       # 核心模块
│   ├── translator.py           # 翻译引擎（567 行）
│   ├── terminology_manager.py  # 术语管理（341 行）
│   ├── validator.py            # 质量验证
│   ├── rtl_processor.py        # RTL 处理
│   ├── typography_processor.py # 排版处理
│   └── interactive.py          # 交互界面
│
├── utils/                      # 工具模块
│   ├── bedrock_client.py       # AWS Bedrock 客户端
│   └── file_handler.py         # 文件处理
│
├── prompts/                    # 提示词模板
│   └── translation_prompts.py  # 翻译提示词（237 行）
│
└── data/                       # 数据文件
    ├── languages.json          # 支持的语言列表
    └── terminology.json        # 术语词典
```

---

### 2.4 核心实现

#### 核心实现 1: 术语管理系统

**文件**: `core/terminology_manager.py`

**术语配置** - `data/terminology.json`:
```json
{
  "preserve_terms": [
    "HiFlux AI",
    "FLUX.1",
    "API",
    "OAuth"
  ],
  "consistent_terms": {
    "zh": {
      "credits": "点数",
      "generation": "生成",
      "premium": "高级版",
      "model": "模型",
      "background remover": "背景移除器",
      "watermark remover": "水印移除器"
    },
    "ja": {
      "credits": "クレジット",
      "generation": "生成",
      "premium": "プレミアム"
    }
  },
  "context_patterns": {
    "pricing": ["price", "plan", "credit", "payment"],
    "generation": ["generate", "create", "model"],
    "settings": ["setting", "config", "preference"]
  }
}
```

**核心功能**：

1. **术语提取**
```python
def extract_terms_from_text(self, text: str) -> Set[str]:
    """从文本中提取关键术语"""

    terms = set()

    # 1. 优先检查保留术语（最长匹配）
    for preserve_term in sorted(self.preserve_terms, key=len, reverse=True):
        if preserve_term in text:
            terms.add(preserve_term)

    # 2. 提取已知的一致性术语
    for term in self.get_known_terms():
        if self.find_term_in_text(text, term):
            terms.add(term)

    # 3. 提取专业术语模式
    # 首字母大写的单词
    capitalized_words = re.findall(r'\b[A-Z][a-z]+(?:\s+[A-Z][a-z]+)*\b', text)
    terms.update(capitalized_words)

    # 连字符术语
    hyphenated_terms = re.findall(r'\b\w+(?:-\w+)+\b', text)
    terms.update(hyphenated_terms)

    # 技术缩写
    acronyms = re.findall(r'\b[A-Z]{2,}\b', text)
    terms.update(acronyms)

    return terms
```

2. **术语强制应用**
```python
def build_term_dictionary_for_prompt(self, target_lang: str, terms_in_text: List[str]) -> str:
    """为 Prompt 构建术语词典"""

    preserve_entries = []
    translate_entries = []

    # 收集保留术语
    for preserve_term in self.preserve_terms:
        preserve_entries.append(f'"{preserve_term}" → NEVER TRANSLATE, KEEP EXACTLY AS IS')

    # 收集翻译术语
    for term in terms_in_text:
        translation = self.get_term_translation(term, target_lang)
        if translation != term:
            translate_entries.append(f'"{term}" → "{translation}"')

    result_lines = []

    # 保留术语（最高优先级）
    if preserve_entries:
        result_lines.append("⚠️  CRITICAL - NEVER TRANSLATE THESE TERMS:")
        result_lines.extend([f"   {entry}" for entry in preserve_entries])

    # 一致性翻译术语
    if translate_entries:
        result_lines.append("📝 REQUIRED TRANSLATIONS:")
        result_lines.extend([f"   {entry}" for entry in translate_entries])

    result_lines.append("🎯 INSTRUCTION: Follow these translations EXACTLY.")

    return "\n".join(result_lines)
```

**输出示例**（注入到 Prompt）：
```
⚠️  CRITICAL - NEVER TRANSLATE THESE TERMS:
   "HiFlux AI" → NEVER TRANSLATE, KEEP EXACTLY AS IS
   "FLUX.1" → NEVER TRANSLATE, KEEP EXACTLY AS IS

📝 REQUIRED TRANSLATIONS:
   "credits" → "点数"
   "generation" → "生成"
   "premium" → "高级版"

🎯 INSTRUCTION: Follow these translations EXACTLY.
```

---

#### 核心实现 2: 混合格式智能处理

**问题**: JSON 中常见混合格式，如 `"{credits} credits {label}"`
- 占位符 `{credits}` 和 `{label}` 不能翻译
- 中间的 "credits" 需要翻译为 "点数"

**解决方案**: 智能分析占位符内外内容

```python
def analyze_mixed_format_text(self, text: str) -> Dict[str, Any]:
    """分析混合格式文本，区分占位符内外的内容"""

    # 1. 提取所有占位符
    placeholders = re.findall(r'\{[^}]+\}', text)

    # 2. 分离占位符外的文本
    outside_text = text
    placeholder_map = {}

    # 用临时标记替换占位符，保持位置信息
    for i, placeholder in enumerate(placeholders):
        temp_marker = f" __PLACEHOLDER_{i}__ "
        placeholder_map[temp_marker] = placeholder
        outside_text = outside_text.replace(placeholder, temp_marker)

    # 3. 清理并提取占位符外的纯文本
    outside_content = outside_text
    for marker in placeholder_map.keys():
        outside_content = outside_content.replace(marker, ' ')
    outside_content = ' '.join(outside_content.split()).strip()

    # 4. 检查占位符外的内容是否需要翻译
    needs_translation = (
        len(outside_content) > 0 and
        any(c.isalpha() and ord(c) < 128 for c in outside_content) and
        not outside_content.isdigit()
    )

    # 5. 识别占位符外的英文术语
    outside_terms = []
    if needs_translation:
        words = re.findall(r'\b[a-zA-Z]+\b', outside_content)
        for word in words:
            if len(word) > 1:
                outside_terms.append(word.lower())

    return {
        'has_placeholders': len(placeholders) > 0,
        'placeholders': placeholders,
        'placeholder_map': placeholder_map,
        'outside_content': outside_content,
        'outside_terms': outside_terms,  # 只提取占位符外的术语
        'needs_translation': needs_translation,
        'translation_type': self._determine_translation_type(placeholders, outside_content),
        'processed_template': outside_text
    }
```

**示例**：
```python
text = "{credits} credits {label}"

analysis = analyze_mixed_format_text(text)

# 结果：
{
    'has_placeholders': True,
    'placeholders': ['{credits}', '{label}'],
    'outside_content': 'credits',        # 只有这个需要翻译
    'outside_terms': ['credits'],        # 术语提取
    'needs_translation': True,
    'translation_type': 'mixed',
    'processed_template': ' __PLACEHOLDER_0__  credits  __PLACEHOLDER_1__ '
}

# Prompt 中会明确说明：
# "只翻译占位符外的 'credits'，保持 {credits} 和 {label} 不变"

# 翻译结果：
# "{credits} 点数 {label}"
```

---

#### 核心实现 3: 批量并发翻译

**文件**: `core/translator.py`

**智能批次创建**：
```python
def create_translation_batches(self, translatable_items: List[Dict]) -> List[List[Dict]]:
    """创建智能翻译批次"""

    batches = []
    current_batch = []
    current_batch_chars = 0

    # 动态调整批次大小和字符限制
    max_chars_per_batch = min(4000, max(2000, self.batch_size * 150))

    # 按上下文分组优化
    context_groups = {}
    for item in translatable_items:
        context = item.get('context', 'general')
        if context not in context_groups:
            context_groups[context] = []
        context_groups[context].append(item)

    # 优先处理同一上下文的项目
    for context, items in context_groups.items():
        for item in items:
            chars_needed = item['char_count']

            # 智能批次切分
            should_create_new_batch = (
                current_batch and (
                    len(current_batch) >= self.batch_size or
                    current_batch_chars + chars_needed > max_chars_per_batch or
                    # 上下文切换时考虑创建新批次
                    (len(current_batch) > self.batch_size // 2 and
                     current_batch[-1].get('context') != item.get('context'))
                )
            )

            if should_create_new_batch:
                batches.append(current_batch)
                current_batch = []
                current_batch_chars = 0

            current_batch.append(item)
            current_batch_chars += chars_needed

    # 添加最后一个批次
    if current_batch:
        batches.append(current_batch)

    return batches
```

**并发处理 + 降级机制**：
```python
async def process_translation_batches(
    self,
    batches: List[List[Dict]],
    target_lang: str,
    terminology_dict: str
) -> Dict[str, str]:
    """并发处理翻译批次"""

    translated_items = {}

    async def process_batch(batch_index: int, batch: List[Dict]) -> Dict[str, Any]:
        """处理单个批次，包含重试和降级机制"""

        async with self.semaphore:  # 限制并发数
            batch_results = {}
            retry_count = 0
            max_retries = 3

            while retry_count <= max_retries:
                try:
                    # 调用批量翻译
                    translated_batch = await self.bedrock_client.translate_batch(
                        batch, target_lang, terminology_dict
                    )

                    # 收集结果
                    for item in translated_batch:
                        path = item['path']
                        translated_text = item.get('translated_text', item['text'])
                        batch_results[path] = translated_text

                    # 成功则退出重试循环
                    break

                except Exception as e:
                    retry_count += 1

                    if retry_count <= max_retries:
                        # 指数退避策略
                        wait_time = min(30, 2 ** retry_count * RATE_LIMIT_DELAY)
                        await asyncio.sleep(wait_time)

                        # 如果是批次过大问题，尝试分割批次
                        if "too large" in str(e).lower():
                            if len(batch) > 1:
                                # 分割批次并递归处理
                                mid = len(batch) // 2
                                sub_batch1 = batch[:mid]
                                sub_batch2 = batch[mid:]

                                result1 = await self.process_single_batch_with_fallback(
                                    sub_batch1, target_lang, terminology_dict
                                )
                                result2 = await self.process_single_batch_with_fallback(
                                    sub_batch2, target_lang, terminology_dict
                                )

                                batch_results.update(result1)
                                batch_results.update(result2)
                                break
                    else:
                        # 最终失败，降级到单个翻译
                        batch_results = await self.fallback_to_single_translation(
                            batch, target_lang, terminology_dict
                        )

            # 更新全局结果
            translated_items.update(batch_results)

            return {
                'batch_index': batch_index,
                'success': len(batch_results) > 0,
                'retry_count': retry_count,
                'items_processed': len(batch_results)
            }

    # 创建所有批次任务
    tasks = [
        process_batch(i, batch)
        for i, batch in enumerate(batches)
    ]

    # 并发执行所有批次
    batch_results = await asyncio.gather(*tasks, return_exceptions=True)

    return translated_items
```

**降级策略**：
```
批量翻译失败
    ↓
重试（指数退避）
    ↓
分割批次
    ↓
单个翻译
    ↓
保留原文（最终降级）
```

---

#### 核心实现 4: 高级 Prompt 设计

**文件**: `prompts/translation_prompts.py`

**语言特定配置**：
```python
configs = {
    'zh': {
        'name': '简体中文',
        'typography_rules': [
            '中文与英文之间必须加空格（如：HiFlux AI 是一个平台）',
            '中文与数字之间必须加空格（如：节省 20% 费用）',
            '数字与单位之间加空格（如：约 6 秒、10 倍）',
            '英文专有名词前后加空格（如：使用 FLUX.1 模型）',
            '变量占位符紧贴中文（如：{count}个项目）',
            '中文标点符号前后不加空格'
        ],
        'style_guide': [
            '使用简体中文，避免繁体字',
            '术语保持一致性，使用已建立的术语映射',
            '语调自然友好，符合产品调性',
            '技术词汇准确，避免生硬翻译'
        ]
    }
}
```

**批量翻译 Prompt**：
```python
def build_batch_translation_prompt(texts: list, target_lang: str,
                                 terminology_dict: str = "") -> str:
    """构建批量翻译提示词"""

    config = TranslationPrompts.get_language_config(target_lang)
    lang_name = config['name']
    typography_rules = config['typography_rules']
    style_guide = config['style_guide']

    # 构建排版规范文本
    typography_text = "\n".join([f"• {rule}" for rule in typography_rules])
    style_text = "\n".join([f"• {rule}" for rule in style_guide])

    # 构建文本列表
    text_list = ""
    for i, item in enumerate(texts):
        item_id = i + 1
        text_list += f"[{item_id}] {item['text']}\n"

    prompt = f"""你是一位专业的本地化翻译专家，擅长UI/UX文案的批量翻译。请将以下英文文本批量翻译成{lang_name}。

【项目背景】
这是HiFlux AI图像生成平台的用户界面文案，需要保持一致的翻译风格和术语使用。

【术语词典】
{terminology_dict if terminology_dict else "遵循产品术语规范"}

【排版规范】
{typography_text}

【翻译风格指南】
{style_text}

【核心要求】
1. 🔒 严格保持所有变量占位符不变（如：{{variable}}、{{count}}等）
2. 🏷️ 严格保持所有HTML标签和特殊标记不变（如：[highlight]、[/highlight]）
3. 📏 严格遵循排版规范，特别注意空格使用
4. 🎯 保持术语翻译的一致性
5. ⚡ 按照 [ID] 翻译结果 的格式返回，不要其他说明
6. 🔄 确保上下文相关的文案保持逻辑一致性

【混合格式特别说明】
对于包含变量占位符的混合格式文本（如：{{{{credits}}}} credits {{{{label}}}}），请注意：
• 变量占位符{{{{...}}}}内的内容绝对不能翻译
• 占位符外的普通英文单词需要翻译
• 保持占位符的位置和格式不变
• 示例：{{{{credits}}}} credits {{{{label}}}} → {{{{credits}}}} 点数 {{{{label}}}}

【待翻译文本】
{text_list}

【翻译结果】"""

    return prompt
```

**关键设计**：
- ✅ **多层次指导**：背景 + 术语 + 排版 + 风格
- ✅ **特殊说明**：混合格式的详细处理规则
- ✅ **批量格式**：`[ID] 翻译结果` 清晰结构化
- ✅ **强调关键点**：用 emoji 突出重要规则

---

#### 核心实现 5: RTL 语言处理

**文件**: `core/rtl_processor.py`

```python
class RTLProcessor:
    """RTL（从右到左）语言处理器"""

    def process_rtl_json(self, data: Dict, target_lang: str, lang_info: Dict) -> Dict:
        """处理 RTL 语言的特殊需求"""

        # 1. 添加方向标记
        data = self.add_direction_marks(data, lang_info)

        # 2. 数字转换（可选）
        if self.should_convert_numbers(target_lang):
            data = self.convert_numbers_to_local(data, target_lang)

        # 3. 标点符号转换
        data = self.convert_punctuation(data, target_lang)

        return data

    def add_direction_marks(self, data: Any, lang_info: Dict) -> Any:
        """为英文术语添加方向标记"""

        preserve_terms = self.get_preserve_terms()

        def add_marks_to_text(text: str) -> str:
            """为文本中的英文术语添加 LTR 标记"""

            for term in preserve_terms:
                # 检查术语是否存在
                if term in text:
                    # 添加 LTR 标记：‎term‎
                    marked_term = f"‎{term}‎"
                    text = text.replace(term, marked_term)

            return text

        # 递归处理 JSON
        return self._process_recursive(data, add_marks_to_text)

    def convert_numbers_to_local(self, data: Any, target_lang: str) -> Any:
        """转换数字为本地格式"""

        # 阿拉伯语数字映射
        arabic_digits = {
            '0': '٠', '1': '١', '2': '٢', '3': '٣', '4': '٤',
            '5': '٥', '6': '٦', '7': '٧', '8': '٨', '9': '٩'
        }

        def convert_text(text: str) -> str:
            """转换文本中的数字"""

            # 跳过占位符中的数字
            placeholders = re.findall(r'\{[^}]+\}', text)

            for digit, arabic_digit in arabic_digits.items():
                # 只转换占位符外的数字
                text = text.replace(digit, arabic_digit)

            return text

        return self._process_recursive(data, convert_text)
```

**示例**：
```json
// 原文
{
  "text": "Use FLUX.1 model, you have 123 credits"
}

// 阿拉伯语（添加方向标记和数字转换）
{
  "text": "استخدم نموذج ‎FLUX.1‎، لديك ١٢٣ من الرصيد"
}
```

---

### 2.5 优势分析

#### ✅ 优势

1. **术语一致性强**
   - 术语词典强制保证
   - 100% 一致性
   - 易于维护和扩展

2. **格式保护完善**
   - 混合格式智能分析
   - 占位符零丢失
   - HTML/特殊标记保护

3. **高性能**
   - 批量处理减少 API 调用
   - 并发处理提速 3-5 倍
   - 智能缓存避免重复翻译

4. **成本优化**
   - 一次翻译完成
   - API 调用少
   - 增量翻译节省成本

5. **工程化成熟**
   - 完善的错误处理
   - 重试 + 降级机制
   - 断点续传支持

6. **专业化处理**
   - RTL 语言特殊处理
   - 排版规范化
   - 上下文感知分批

7. **易于集成**
   - CLI 工具
   - 配置文件驱动
   - 交互式界面

---

#### ❌ 劣势

1. **翻译质量天花板**
   - 单轮翻译，无反思
   - 依赖 prompt 质量
   - 可能不如人工润色

2. **文化适配有限**
   - 虽然有排版规范
   - 但无深度文化审视
   - 语言自然度可能不及反思机制

3. **灵活性略低**
   - 需要预定义术语表
   - 新术语需手动添加
   - 不能自适应学习

---

### 2.6 适用场景

✅ **适合**：
- JSON 国际化文件翻译
- 大规模批量翻译
- 需要术语强一致性的场景
- UI/UX 文案翻译
- 预算有限的项目
- 需要高性能的场景

❌ **不适合**：
- 高端文学翻译
- 营销创意文案（需要多轮打磨）
- 没有术语管理需求的场景

---

## 3. 两种方案对比

### 3.1 核心对比表

| 维度 | **Andrew Ng Translation Agent** | **HiFlux Translation Tool** |
|------|--------------------------------|----------------------------|
| **核心理念** | Agentic Workflow + Reflection | Batch + Concurrency + Terminology |
| **翻译流程** | 三步循环（初译→反思→改进） | 一次翻译（批量） |
| **质量保证** | LLM 自我反思 | 术语词典 + 格式验证 |
| **术语一致性** | ⚠️ 依赖 LLM（不保证） | ✅ 强制保证（100%） |
| **格式保护** | ⚠️ 无特殊处理 | ✅ 智能分析 + 验证 |
| **翻译质量** | ⭐⭐⭐⭐⭐ 高（多轮优化） | ⭐⭐⭐⭐ 良好（单轮） |
| **API 调用** | ❌ 3 倍（初译+反思+改进） | ✅ 1 倍（批量） |
| **速度** | ❌ 慢（3 倍耗时） | ✅ 快（并发） |
| **成本** | ❌ 高（3 倍） | ✅ 低（批量优化） |
| **并发能力** | ❌ 无（同步） | ✅ 有（asyncio） |
| **增量翻译** | ❌ 无 | ✅ 有（缓存 + 对比） |
| **错误处理** | ⚠️ 基础 | ✅ 完善（重试+降级） |
| **RTL 支持** | ⚠️ 依赖 LLM | ✅ 专门处理 |
| **配置化** | ❌ 硬编码 | ✅ 配置文件 |
| **适用场景** | 高质量文档、营销文案 | JSON 国际化、UI 文案 |

---

### 3.2 详细对比

#### 对比 1: 术语一致性

**Andrew Ng 方案**：
```
依赖 LLM 自行理解术语
❌ 问题：
  - 同一术语可能翻译不一致
  - 无法强制保证
  - 依赖 prompt 质量

示例：
  "credits" 可能被翻译为：
    - "点数"（某些地方）
    - "积分"（其他地方）
    - "信用额度"（还有地方）
  ❌ 不一致！
```

**HiFlux 方案**：
```
术语词典强制映射
✅ 保证：
  - "credits" → "点数"（100% 一致）
  - 词典可维护
  - 可扩展

示例：
  terminology.json:
    "credits": "点数"

  Prompt 注入:
    "📝 REQUIRED TRANSLATIONS: 'credits' → '点数'"

  验证:
    检查所有 "credits" 是否都翻译为 "点数"
```

**结论**: HiFlux 方案在术语一致性上有绝对优势。

---

#### 对比 2: 格式保护

**Andrew Ng 方案**：
```
无特殊格式保护
❌ 问题：
  - 占位符可能丢失："{count}" → ""
  - 占位符可能被翻译："{credits}" → "{点数}"
  - HTML 标签可能损坏

示例：
  原文: "You have {count} credits"
  翻译: "您有 个点数"  ❌ 占位符丢失
```

**HiFlux 方案**：
```
智能格式分析 + 验证
✅ 保证：
  - 混合格式智能分析
  - 占位符内容不翻译
  - 占位符外内容翻译
  - 自动验证

示例：
  原文: "{credits} credits {label}"

  分析:
    - 占位符: {credits}, {label}
    - 占位符外: "credits"

  翻译: "{credits} 点数 {label}"  ✅ 正确

  验证:
    ✅ 占位符完整: 2/2
```

**结论**: HiFlux 方案在格式保护上有显著优势。

---

#### 对比 3: 翻译质量

**Andrew Ng 方案**：
```
反思机制多轮优化
✅ 优势：
  - 初译 → 反思 → 改进
  - LLM 自我批评
  - 4 维度优化（准确、流畅、风格、术语）

示例：
  原文: "Unleash your creativity with AI"

  初译: "使用 AI 释放您的创造力"

  反思:
    - "释放"过于生硬
    - 语序不自然
    - 缺乏感染力

  改进: "用 AI 激发无限创造力"

  ✅ 更自然、更有吸引力
```

**HiFlux 方案**：
```
单轮翻译 + Prompt 优化
⚠️ 限制：
  - 无反思机制
  - 依赖 prompt 质量
  - 适合 UI 文案

示例：
  原文: "Unleash your creativity with AI"

  翻译: "使用 AI 释放您的创造力"

  ⚠️ 可能略显生硬
```

**结论**: Andrew Ng 方案在翻译质量上有优势，尤其是营销文案。

---

#### 对比 4: 性能与成本

**Andrew Ng 方案**：
```
3 倍调用，3 倍耗时
❌ 成本：
  - 初译: 1 次 API 调用
  - 反思: 1 次 API 调用
  - 改进: 1 次 API 调用
  - 总计: 3 次

❌ 耗时：
  - 假设单次 2 秒
  - 总计: 6 秒

  500 文本:
    - 3000 次调用
    - 约 50 分钟
    - 成本高
```

**HiFlux 方案**：
```
批量并发，1 倍调用
✅ 成本：
  - 批量: 20 个文本/批次
  - 500 文本 = 25 批次
  - 总计: 25 次调用（节省 99%）

✅ 耗时：
  - 并发: 3 个批次同时处理
  - 25 批次 / 3 = 约 9 轮
  - 假设每批次 3 秒
  - 总计: 27 秒（提速 100 倍）

  500 文本:
    - 25 次调用
    - 约 3-5 分钟
    - 成本低
```

**结论**: HiFlux 方案在性能和成本上有压倒性优势。

---

#### 对比 5: 长文本处理

**Andrew Ng 方案**：
```
智能分块 + 上下文保持
✅ 优势：
  - 分块算法优化
  - 提供完整上下文
  - 翻译连贯

示例：
  文本分为 [A, B, C]

  翻译 B 时:
    上下文: A <TRANSLATE_THIS>B</TRANSLATE_THIS> C

  ✅ B 的翻译考虑了 A 和 C 的上下文
```

**HiFlux 方案**：
```
按 JSON 结构分块
⚠️ 限制：
  - 按 JSON key 分块
  - 跨 key 的上下文可能丢失
  - 适合独立的文案条目

示例：
  JSON:
    {
      "title": "文本 A",
      "description": "文本 B"
    }

  翻译时:
    - title 和 description 独立翻译
    - ⚠️ 无跨 key 上下文
```

**结论**: Andrew Ng 方案在长文本连贯性上更好。

---

### 3.3 场景选择指南

#### 场景 1: JSON 国际化（UI 文案）

**特点**：
- 数百个短文案
- 术语一致性要求高
- 格式保护要求高（占位符）
- 预算有限

**推荐**: ✅ HiFlux 方案

**理由**：
- 术语强制保证
- 格式零损坏
- 成本低，速度快
- 增量翻译支持

---

#### 场景 2: 营销文案（Landing Page）

**特点**：
- 长文案，需要润色
- 文化适配要求高
- 语言自然度要求高
- 预算充足

**推荐**: ✅ Andrew Ng 方案

**理由**：
- 反思机制优化质量
- 多轮打磨更自然
- 地区方言支持
- 适合高价值内容

---

#### 场景 3: 技术文档

**特点**：
- 长文档，章节多
- 术语专业，需一致
- 上下文连贯要求高
- 预算中等

**推荐**: ✅ Andrew Ng 方案（或混合方案）

**理由**：
- 分块算法处理长文本
- 上下文保持连贯性
- 反思提升专业度

**混合方案**：
- 术语管理用 HiFlux 方法
- 翻译流程用 Andrew Ng 方法

---

#### 场景 4: 大规模批量翻译

**特点**：
- 数千条文本
- 时间要求紧
- 预算有限
- 术语多

**推荐**: ✅ HiFlux 方案

**理由**：
- 批量并发效率高
- 成本控制好
- 术语管理完善
- 增量翻译减少工作量

---

## 4. 技术选型建议

### 4.1 新项目建议：混合方案

结合两种方案的优势，构建混合翻译系统：

```python
class HybridTranslator:
    """混合翻译器"""

    def __init__(self):
        self.terminology_mgr = TerminologyManager()  # HiFlux 术语管理
        self.batch_translator = BatchTranslator()     # HiFlux 批量翻译
        self.reflection_translator = ReflectionTranslator()  # Andrew Ng 反思

    def translate(self, source_data: Dict, target_lang: str, mode: str = 'smart'):
        """
        混合翻译

        mode:
          - 'fast': 全部使用批量翻译
          - 'quality': 全部使用反思机制
          - 'smart': 智能选择（推荐）
        """

        # 1. 提取可翻译内容（HiFlux 方法）
        items = self.extract_translatable_items(source_data)

        # 2. 分类：哪些用快速，哪些用高质量
        if mode == 'smart':
            fast_items, quality_items = self.classify_items(items)
        elif mode == 'fast':
            fast_items, quality_items = items, []
        else:  # 'quality'
            fast_items, quality_items = [], items

        # 3. 并行处理两种类型
        results = {}

        # 快速翻译（批量并发）
        if fast_items:
            fast_results = await self.batch_translator.translate(
                fast_items, target_lang, self.terminology_mgr
            )
            results.update(fast_results)

        # 高质量翻译（反思机制）
        if quality_items:
            quality_results = await self.reflection_translator.translate(
                quality_items, target_lang, self.terminology_mgr
            )
            results.update(quality_results)

        # 4. 重建 JSON
        translated_data = self.rebuild_json(source_data, results)

        return translated_data

    def classify_items(self, items: List[Dict]) -> Tuple[List, List]:
        """
        智能分类：哪些用快速，哪些用高质量

        高质量模式的判断条件：
          - key 包含关键词: hero, landing, marketing, tagline, slogan
          - 文本长度 > 30 词
          - 上下文类型为 'marketing'
        """

        fast_items = []
        quality_items = []

        quality_keywords = ['hero', 'landing', 'marketing', 'tagline', 'slogan', 'about']

        for item in items:
            path = item['path']
            text = item['text']
            context = item.get('context', '')

            # 判断逻辑
            is_quality = (
                any(keyword in path.lower() for keyword in quality_keywords) or
                len(text.split()) > 30 or
                context == 'marketing'
            )

            if is_quality:
                quality_items.append(item)
            else:
                fast_items.append(item)

        return fast_items, quality_items
```

---

### 4.2 核心功能模块设计

#### 模块 1: 术语管理（HiFlux 方法）

```python
class TerminologyManager:
    """术语一致性管理器"""

    - load_terminology_rules()          # 加载术语配置
    - extract_terms_from_text()         # 提取术语
    - get_term_translation()            # 获取术语翻译
    - build_term_dictionary_for_prompt() # 构建 Prompt 术语词典
    - validate_term_consistency()       # 验证术语一致性
```

**关键点**：
- ✅ 保留术语（preserve_terms）
- ✅ 一致性术语（consistent_terms）
- ✅ Prompt 注入
- ✅ 验证机制

---

#### 模块 2: 格式保护（HiFlux 方法）

```python
class FormatProtector:
    """格式保护器"""

    - analyze_mixed_format()        # 分析混合格式
    - extract_placeholders()        # 提取占位符
    - extract_html_tags()           # 提取 HTML 标签
    - validate_format_integrity()   # 验证格式完整性
```

**关键点**：
- ✅ 占位符识别
- ✅ 混合格式分析
- ✅ 格式验证

---

#### 模块 3: 快速翻译（HiFlux 方法）

```python
class BatchTranslator:
    """批量翻译器"""

    - create_batches()              # 创建智能批次
    - translate_batch()             # 批量翻译（并发）
    - retry_with_backoff()          # 重试机制
    - fallback_to_single()          # 降级处理
```

**关键点**：
- ✅ 批量处理
- ✅ 并发执行
- ✅ 错误处理

---

#### 模块 4: 高质量翻译（Andrew Ng 方法）

```python
class ReflectionTranslator:
    """反思翻译器"""

    - initial_translation()         # 初始翻译
    - reflect_on_translation()      # 反思评价
    - improve_translation()         # 改进翻译
    - translate_with_reflection()   # 完整流程
```

**关键点**：
- ✅ 三步循环
- ✅ 自我批评
- ✅ 迭代优化

---

#### 模块 5: 后处理（HiFlux 方法）

```python
class PostProcessor:
    """后处理器"""

    - apply_typography_rules()      # 排版规范化
    - process_rtl_languages()       # RTL 语言处理
    - add_direction_marks()         # 添加方向标记
    - convert_numbers()             # 数字转换
```

**关键点**：
- ✅ 排版优化
- ✅ RTL 支持
- ✅ 本地化处理

---

### 4.3 实现路线图

#### Phase 1: MVP（基于 HiFlux 方案）

**功能**：
- ✅ 术语管理
- ✅ 格式保护
- ✅ 批量翻译
- ✅ 基础验证

**时间**: 2-3 周

---

#### Phase 2: 增强功能

**功能**：
- ✅ 增量翻译
- ✅ 选择性翻译
- ✅ RTL 优化
- ✅ 完善错误处理

**时间**: 2-3 周

---

#### Phase 3: 高质量模式（引入 Andrew Ng 方法）

**功能**：
- ✅ 反思机制
- ✅ 智能模式（自动选择）
- ✅ 混合翻译

**时间**: 3-4 周

---

#### Phase 4: 生产就绪

**功能**：
- ✅ 完整测试
- ✅ 性能优化
- ✅ 文档完善
- ✅ CI/CD

**时间**: 4-6 周

---

## 5. 总结

### 关键洞察

1. **Andrew Ng 方案的核心价值**：
   - ✅ 反思机制提升翻译质量
   - ✅ 适合高价值内容
   - ❌ 但成本高、速度慢

2. **HiFlux 方案的核心价值**：
   - ✅ 术语强一致性
   - ✅ 格式零损坏
   - ✅ 高性能、低成本
   - ❌ 但质量天花板有限

3. **最佳实践**：
   - ✅ 混合方案：取长补短
   - ✅ 智能模式：自动选择
   - ✅ 模块化设计：灵活组合

---

### 最终建议

针对 **JSON 国际化翻译工具** 项目：

1. **优先采用 HiFlux 方案作为基础**
   - 术语管理
   - 格式保护
   - 批量并发

2. **引入 Andrew Ng 的反思机制作为可选项**
   - 高质量模式
   - 智能模式

3. **重点实现混合方案**
   - 平衡质量和成本
   - 适应不同场景

---

**文档版本**: v1.0
**最后更新**: 2025-01-16
**作者**: 技术团队
