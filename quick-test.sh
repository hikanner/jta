#!/bin/bash

# 🧪 Jta CLI 快速测试脚本
# 用法: ./quick-test.sh

set -e  # 遇到错误立即退出

echo "🚀 Jta CLI 快速测试"
echo "===================="
echo ""

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 检查函数
check_step() {
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ $1${NC}"
    else
        echo -e "${RED}❌ $1 失败${NC}"
        exit 1
    fi
}

# 1. 检查 API Key
echo -e "${BLUE}📋 步骤 1: 检查 API Key${NC}"
if [ -z "$OPENAI_API_KEY" ] && [ -z "$ANTHROPIC_API_KEY" ] && [ -z "$GEMINI_API_KEY" ]; then
    echo -e "${RED}❌ 未找到 API Key!${NC}"
    echo ""
    echo "请设置以下任一环境变量:"
    echo "  export OPENAI_API_KEY='sk-...'"
    echo "  export ANTHROPIC_API_KEY='sk-ant-...'"
    echo "  export GEMINI_API_KEY='...'"
    exit 1
fi

if [ ! -z "$OPENAI_API_KEY" ]; then
    echo -e "${GREEN}✅ OpenAI API Key 已设置${NC}"
    PROVIDER="openai"
elif [ ! -z "$ANTHROPIC_API_KEY" ]; then
    echo -e "${GREEN}✅ Anthropic API Key 已设置${NC}"
    PROVIDER="anthropic"
elif [ ! -z "$GEMINI_API_KEY" ]; then
    echo -e "${GREEN}✅ Google Gemini API Key 已设置${NC}"
    PROVIDER="google"
fi
echo ""

# 2. 构建 CLI
echo -e "${BLUE}📋 步骤 2: 构建 CLI${NC}"
if [ ! -f "./jta" ]; then
    echo "正在构建..."
    go build -o jta ./cmd/jta
    check_step "构建 CLI"
else
    echo -e "${GREEN}✅ CLI 已存在 (跳过构建)${NC}"
fi
echo ""

# 3. 创建测试文件
echo -e "${BLUE}📋 步骤 3: 创建测试文件${NC}"
cat > test-quick.json << 'EOF'
{
  "welcome": "Welcome to Jta",
  "description": "A powerful JSON translation agent with Agentic reflection",
  "start": "Get started",
  "docs": "Read documentation",
  "support": "Get help and support"
}
EOF
check_step "创建测试文件"
echo ""

# 4. 基础翻译测试
echo -e "${BLUE}📋 步骤 4: 基础翻译测试 (英语 → 西班牙语)${NC}"
./jta translate test-quick.json \
  --source en \
  --target es \
  --output test-quick-es.json \
  --provider "$PROVIDER"
check_step "基础翻译"

echo ""
echo -e "${GREEN}翻译结果:${NC}"
cat test-quick-es.json | head -10
echo ""

# 5. 带术语的翻译测试
echo -e "${BLUE}📋 步骤 5: 术语管理测试${NC}"
cat > test-quick-tech.json << 'EOF'
{
  "api_desc": "GitHub API is now available",
  "repo_info": "Clone the Git repository",
  "auth": "OAuth token is required for authentication"
}
EOF

cat > test-quick-terms.json << 'EOF'
{
  "source_language": "en",
  "preserve_terms": ["GitHub", "API", "Git", "OAuth"],
  "consistent_terms": {
    "en": ["repository", "token", "authentication"],
    "zh": ["仓库", "令牌", "身份验证"]
  }
}
EOF

./jta translate test-quick-tech.json \
  --source en \
  --target zh \
  --output test-quick-tech-zh.json \
  --terminology test-quick-terms.json \
  --provider "$PROVIDER"
check_step "术语翻译"

echo ""
echo -e "${GREEN}术语翻译结果:${NC}"
cat test-quick-tech-zh.json
echo ""

# 6. 增量翻译测试
echo -e "${BLUE}📋 步骤 6: 增量翻译测试${NC}"
# 修改源文件
cat > test-quick.json << 'EOF'
{
  "welcome": "Welcome to Jta",
  "description": "A powerful JSON translation agent with Agentic reflection",
  "start": "Get started",
  "docs": "Read documentation",
  "support": "Get help and support",
  "new_feature": "This is a brand new feature",
  "another_new": "Another newly added item"
}
EOF

./jta translate test-quick.json \
  --source en \
  --target es \
  --output test-quick-es.json \
  --incremental \
  --provider "$PROVIDER"
check_step "增量翻译"

echo ""
echo -e "${GREEN}增量翻译结果 (只翻译新增项):${NC}"
cat test-quick-es.json | grep -A2 "new"
echo ""

# 7. 显示统计信息
echo -e "${BLUE}📋 步骤 7: 查看统计信息${NC}"
echo "生成的文件:"
ls -lh test-quick*.json
echo ""

# 8. 完成
echo -e "${GREEN}════════════════════════════════════════${NC}"
echo -e "${GREEN}🎉 所有测试通过！${NC}"
echo -e "${GREEN}════════════════════════════════════════${NC}"
echo ""
echo "生成的文件:"
echo "  - test-quick-es.json (西班牙语翻译)"
echo "  - test-quick-tech-zh.json (中文术语翻译)"
echo ""
echo "下一步:"
echo "  1. 查看完整文档: cat README.md"
echo "  2. 查看详细测试: cat LOCAL_TESTING.md"
echo "  3. 清理测试文件: rm test-quick*.json"
echo ""
