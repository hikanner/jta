# 🧪 本地测试指南

本文档提供 Jta CLI 的完整本地测试流程。

## 📋 前置准备

### 1. 检查环境

```bash
# 确认 Go 版本 (需要 1.23+)
go version

# 确认项目目录
pwd  # 应该在 /Users/ronan/Documents/projects/jta

# 查看项目结构
ls -la
```

### 2. 设置 API Keys

Jta 需要 AI Provider 的 API key。选择以下任一方式：

#### 方式 A: 环境变量 (推荐用于测试)

```bash
# OpenAI (推荐 - 最稳定)
export OPENAI_API_KEY="sk-..."

# 或 Anthropic
export ANTHROPIC_API_KEY="sk-ant-..."

# 或 Gemini
export GEMINI_API_KEY="..."
# 或
export GOOGLE_API_KEY="..."
```

#### 方式 B: 命令行参数

```bash
# 使用 --api-key 参数
jta translate source.json --api-key "sk-..."
```

**验证 API Key 是否设置**:

```bash
# 检查环境变量
echo $OPENAI_API_KEY

# 或尝试翻译一个简单文件（见下文）
```

---

## 🚀 快速开始

### 步骤 1: 构建/安装 CLI

选择以下任一方式：

#### 选项 A: 使用已有的编译文件

```bash
# 项目根目录已有 jta 二进制文件
./jta --help

# 添加到 PATH (可选)
export PATH=$PATH:$(pwd)
jta --help
```

#### 选项 B: 重新构建

```bash
# 使用 Makefile
make build

# 或直接使用 go build
go build -o jta ./cmd/jta

# 验证构建
./jta --version
```

#### 选项 C: 安装到 GOPATH

```bash
go install ./cmd/jta

# 验证安装
jta --version
```

### 步骤 2: 准备测试文件

使用项目提供的示例文件：

```bash
# 查看示例文件
ls -la examples/

# 示例文件内容:
# - source.json (英文源文件)
# - target.json (目标翻译文件)
# - terminology.json (术语表)
```

**或创建一个简单的测试文件**:

```bash
cat > test-source.json << 'EOF'
{
  "welcome": "Welcome to Jta",
  "description": "A powerful JSON translation agent",
  "start": "Get started",
  "docs": "Documentation"
}
EOF
```

---

## 🧪 测试用例

### 测试 1: 基础翻译 (最简单)

```bash
# 翻译到西班牙语
./jta translate test-source.json \
  --source en \
  --target es \
  --output test-es.json

# 查看结果
cat test-es.json
```

**预期输出**:
```json
{
  "welcome": "Bienvenido a Jta",
  "description": "Un potente agente de traducción JSON",
  "start": "Comenzar",
  "docs": "Documentación"
}
```

**如果成功**: ✅ 基础功能正常！

**如果失败**: 
- 检查 API key 是否正确设置
- 查看错误信息
- 尝试添加 `--verbose` 查看详细日志

---

### 测试 2: 指定 Provider 和 Model

```bash
# 使用 OpenAI GPT-4o
./jta translate test-source.json \
  --source en \
  --target zh \
  --output test-zh.json \
  --provider openai \
  --model gpt-5 \
  --verbose

# 使用 Anthropic Claude
./jta translate test-source.json \
  --source en \
  --target fr \
  --output test-fr.json \
  --provider anthropic \
  --model claude-sonnet-4-5
```

---

### 测试 3: 术语管理 (核心功能)

```bash
# 创建带技术术语的源文件
cat > test-tech.json << 'EOF'
{
  "api_available": "GitHub API is available",
  "clone_repo": "Clone the Git repository",
  "oauth_token": "OAuth token required"
}
EOF

# 创建术语表
cat > test-terminology.json << 'EOF'
{
  "source_language": "en",
  "preserve_terms": ["GitHub", "API", "Git", "OAuth"],
  "consistent_terms": {
    "en": ["repository", "token"],
    "zh": ["仓库", "令牌"]
  }
}
EOF

# 使用术语表翻译
./jta translate test-tech.json \
  --source en \
  --target zh \
  --output test-tech-zh.json \
  --terminology test-terminology.json \
  --verbose

# 验证术语是否被保留
cat test-tech-zh.json | grep "GitHub"  # 应该保留 GitHub
cat test-tech-zh.json | grep "仓库"    # repository -> 仓库
```

**预期**: `GitHub`, `API`, `Git`, `OAuth` 保持不变，其他术语正确翻译。

---

### 测试 4: 增量翻译 (高级功能)

```bash
# 第一次完整翻译
./jta translate test-source.json \
  --source en \
  --target es \
  --output test-es.json

# 修改源文件 (添加新内容)
cat > test-source.json << 'EOF'
{
  "welcome": "Welcome to Jta",
  "description": "A powerful JSON translation agent",
  "start": "Get started",
  "docs": "Documentation",
  "new_feature": "This is a new feature",
  "another_new": "Another new item"
}
EOF

# 增量翻译 (只翻译新增内容)
./jta translate test-source.json \
  --source en \
  --target es \
  --output test-es.json \
  --incremental \
  --verbose

# 验证: 只有新增的两项被翻译
cat test-es.json | grep "new"
```

**预期**: 只调用 1 次 API，只翻译新增的 2 项。

---

### 测试 5: Key 过滤 (选择性翻译)

```bash
# 创建大文件
cat > test-large.json << 'EOF'
{
  "auth": {
    "login": "Login",
    "logout": "Logout",
    "register": "Register"
  },
  "settings": {
    "profile": "Profile",
    "preferences": "Preferences"
  },
  "home": {
    "welcome": "Welcome",
    "banner": "Banner"
  }
}
EOF

# 只翻译 auth 和 settings 部分
./jta translate test-large.json \
  --source en \
  --target es \
  --output test-filtered.json \
  --keys "auth.*" \
  --keys "settings.*" \
  --verbose

# 验证: home.* 应该不在输出中
cat test-filtered.json
```

**预期**: 只有 `auth` 和 `settings` 部分被翻译，`home` 被过滤。

---

### 测试 6: 批量翻译多种语言

```bash
# 创建简单脚本
cat > translate-all.sh << 'EOF'
#!/bin/bash

SOURCE="test-source.json"
LANGUAGES=("es" "fr" "de" "zh" "ja")

for lang in "${LANGUAGES[@]}"; do
  echo "🌍 Translating to $lang..."
  ./jta translate "$SOURCE" \
    --source en \
    --target "$lang" \
    --output "test-$lang.json" \
    --verbose
  echo "✅ $lang done!"
  echo ""
done

echo "🎉 All translations completed!"
EOF

chmod +x translate-all.sh
./translate-all.sh

# 查看所有输出文件
ls -lh test-*.json
```

---

### 测试 7: RTL 语言 (阿拉伯语/希伯来语)

```bash
# 翻译到阿拉伯语
./jta translate test-source.json \
  --source en \
  --target ar \
  --output test-ar.json \
  --verbose

# 检查 RTL 标记
cat test-ar.json
# 应该看到 Unicode 方向标记 (LRM/RLM)
```

---

## 🔍 调试和排错

### 启用详细日志

```bash
# 添加 --verbose 查看详细信息
./jta translate test-source.json \
  --source en \
  --target es \
  --output test-es.json \
  --verbose

# 查看完整的 API 调用、批处理、反思步骤
```

### 常见问题排查

#### 问题 1: "API key not found"

```bash
# 检查环境变量
env | grep -i "api_key"

# 重新设置
export OPENAI_API_KEY="sk-..."

# 或使用命令行参数
./jta translate test-source.json --api-key "sk-..." ...
```

#### 问题 2: "Rate limit exceeded"

```bash
# 减少并发数
./jta translate test-source.json \
  --concurrency 1 \
  ...

# 增加批处理大小 (减少 API 调用次数)
./jta translate test-source.json \
  --batch-size 50 \
  ...
```

#### 问题 3: 翻译质量不佳

```bash
# 1. 使用更好的模型
./jta translate test-source.json \
  --model gpt-5 \  # 或 claude-sonnet-4-5
  ...

# 2. 添加术语表
./jta translate test-source.json \
  --terminology my-terms.json \
  ...

# 3. 关闭反思机制 (如果需要更快速度)
# 注意: 目前 Agentic 反思默认启用
```

#### 问题 4: 格式元素丢失

```bash
# 测试格式保护
cat > test-format.json << 'EOF'
{
  "greeting": "Hello {name}!",
  "html": "Click <a href=\"url\">here</a>",
  "markdown": "**Bold** text"
}
EOF

./jta translate test-format.json \
  --source en \
  --target es \
  --output test-format-es.json \
  --verbose

# 验证: {name}, <a>标签, **应该被保留
cat test-format-es.json
```

---

## 📊 性能基准测试

### 小文件 (< 20 项)

```bash
time ./jta translate test-source.json \
  --source en \
  --target es \
  --output test-es.json

# 预期: 2-5 秒 (1 批次, 3x API 调用 - translate + reflect + improve)
```

### 中等文件 (50-100 项)

```bash
# 创建 50 项的测试文件
python3 << 'EOF'
import json
data = {f"key_{i}": f"Text item {i}" for i in range(50)}
with open('test-50.json', 'w') as f:
    json.dump(data, f, indent=2)
EOF

time ./jta translate test-50.json \
  --source en \
  --target es \
  --output test-50-es.json \
  --batch-size 20 \
  --concurrency 3

# 预期: 10-20 秒 (3 批次并发)
```

### 大文件 (200+ 项)

```bash
# 测试增量模式的性能优势
python3 << 'EOF'
import json
data = {f"key_{i}": f"Text item {i}" for i in range(200)}
with open('test-200.json', 'w') as f:
    json.dump(data, f, indent=2)
EOF

# 第一次完整翻译
time ./jta translate test-200.json \
  --source en \
  --target es \
  --output test-200-es.json \
  --batch-size 20 \
  --concurrency 5

# 添加 10 个新项
python3 << 'EOF'
import json
with open('test-200.json', 'r') as f:
    data = json.load(f)
for i in range(200, 210):
    data[f"key_{i}"] = f"New text item {i}"
with open('test-200.json', 'w') as f:
    json.dump(data, f, indent=2)
EOF

# 增量翻译 (应该快很多)
time ./jta translate test-200.json \
  --source en \
  --target es \
  --output test-200-es.json \
  --incremental

# 预期: 增量翻译只需 2-3 秒 (vs 完整翻译 40-60 秒)
```

---

## 🧹 清理测试文件

```bash
# 删除所有测试生成的文件
rm -f test-*.json
rm -f translate-all.sh

# 或保留用于后续测试
mkdir -p test-results
mv test-*.json test-results/
```

---

## ✅ 验证清单

测试完成后，确认以下功能正常：

- [ ] ✅ 基础翻译 (单语言)
- [ ] ✅ 多 Provider 支持 (OpenAI/Anthropic/Gemini)
- [ ] ✅ 术语管理 (preserve + consistent terms)
- [ ] ✅ 增量翻译 (只翻译新增内容)
- [ ] ✅ Key 过滤 (选择性翻译)
- [ ] ✅ 批量多语言翻译
- [ ] ✅ RTL 语言支持
- [ ] ✅ 格式保护 (placeholders, HTML, markdown)
- [ ] ✅ 并发批处理性能
- [ ] ✅ 错误处理和重试

---

## 📚 下一步

测试通过后，可以：

1. **阅读完整文档**: `README.md`
2. **查看高级用例**: `README.md` 的 Examples 部分
3. **集成到项目**: 将 `jta` 添加到 CI/CD 流程
4. **贡献代码**: 参考 `CONTRIBUTING.md`

---

## 💡 实用技巧

### Tip 1: 创建别名

```bash
# 添加到 ~/.bashrc 或 ~/.zshrc
alias jta-zh='jta translate --source en --target zh --provider openai'
alias jta-es='jta translate --source en --target es --provider openai'

# 使用
jta-zh input.json --output output-zh.json
```

### Tip 2: 使用配置文件 (未来功能)

```bash
# 创建 .jta.yaml
cat > .jta.yaml << 'EOF'
provider: openai
model: gpt-5
batch_size: 20
concurrency: 3
source: en
terminology: ./my-terms.json
EOF

# 简化命令
jta translate input.json --target es
```

### Tip 3: 监控 API 成本

```bash
# 使用 --verbose 查看 token 使用量
./jta translate input.json ... --verbose | grep -i "tokens"

# 输出示例:
# Total tokens: 1,234 (prompt: 800, completion: 434)
# Estimated cost: $0.02 (based on GPT-4o pricing)
```

---

🎉 **Happy Testing!**
