# Jta - 技术实施方案

> Golang 1.25+ implementation with OOP design patterns and interface-driven architecture

本文档详细描述 Jta 项目的技术架构、实施方案和开发计划。

---

## 📋 目录

1. [技术栈选择](#1-技术栈选择)
2. [架构设计](#2-架构设计)
3. [核心接口设计](#3-核心接口设计)
4. [项目结构](#4-项目结构)
5. [开发阶段规划](#5-开发阶段规划)
6. [详细任务分解](#6-详细任务分解)
7. [测试策略](#7-测试策略)
8. [构建与发布](#8-构建与发布)
9. [性能优化](#9-性能优化)

---

## 1. 技术栈选择

### 1.1 核心技术

| 类别 | 技术/库 | 版本 | 用途 |
|------|---------|------|------|
| **语言** | Go | 1.25+ | 主开发语言 |
| **CLI 框架** | Cobra | v1.8+ | 命令行界面 |
| **配置管理** | Viper | v1.18+ | 配置加载 |
| **UI/进度** | bubbletea + lipgloss | latest | 终端 UI |
| **HTTP 客户端** | Go 标准库 `net/http` | - | HTTP 请求 |
| **JSON 处理** | `github.com/bytedance/sonic` | v1.12+ | 高性能 JSON 解析 |
| **并发控制** | errgroup | - | 并发错误处理 |
| **日志** | zerolog | v1.32+ | 结构化日志 |

### 1.2 AI Provider SDKs（官方 SDK）

| 提供商 | 官方 SDK | 版本 | Context Window |
|--------|---------|------|---------------|
| **OpenAI** | `github.com/openai/openai-go/v3` | v3.6+ | 128K (GPT-4o) |
| **Anthropic** | `github.com/anthropics/anthropic-sdk-go` | latest | 200K (Claude 3.5 Sonnet) |
| **Google** | `google.golang.org/genai` | latest | 1M (Gemini 2.0 Flash) |

### 1.3 开发工具

| 工具 | 用途 |
|------|------|
| **golangci-lint** | 代码检查 |
| **gofmt/goimports** | 代码格式化 |
| **go test** | 测试框架 |
| **testify** | 测试断言库 |
| **mockery** | Mock 生成 |
| **goreleaser** | 跨平台构建和发布 |

---

## 2. 架构设计

### 2.1 整体架构

```
┌─────────────────────────────────────────────────────────────┐
│                      CLI Layer                               │
│                  (Cobra Root Command)                        │
│                 核心功能即翻译，无子命令                        │
└────────────────────────┬────────────────────────────────────┘
                         │
         ┌───────────────┴───────────────┐
         │                               │
┌────────▼────────┐            ┌────────▼────────┐
│   Application   │            │      Config     │
│     Service     │◄───────────┤     Manager     │
└────────┬────────┘            └─────────────────┘
         │
         ├─────────────────────────────────┐
         │                                 │
┌────────▼────────┐              ┌────────▼────────┐
│   Translation   │              │   Terminology   │
│     Engine      │◄─────────────┤     Manager     │
└────────┬────────┘              └─────────────────┘
         │
         ├──────────┬──────────┬──────────┐
         │          │          │          │
┌────────▼──┐  ┌───▼────┐  ┌──▼─────┐  ┌▼────────┐
│  Batch    │  │Agentic │  │ Format │  │   RTL   │
│Translator │  │ Trans  │  │Protect │  │Processor│
└───────────┘  └────────┘  └────────┘  └─────────┘
         │          │          │          │
         └──────────┴──────────┴──────────┘
                     │
         ┌───────────▼───────────┐
         │                       │
┌────────▼────────┐    ┌────────▼────────┐
│   AI Provider   │    │   Validator     │
│     Factory     │    │                 │
└────────┬────────┘    └─────────────────┘
         │
    ┌────┴────┬────────┬────────┐
    │         │        │        │
┌───▼──┐ ┌───▼──┐ ┌───▼──┐ ┌───▼──┐
│OpenAI│ │Claude│ │Gemini│ │Custom│
└──────┘ └──────┘ └──────┘ └──────┘
```

### 2.2 设计模式

#### 2.2.1 Strategy Pattern (AI 提供商)

```go
// 策略接口
type AIProvider interface {
    Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error)
    Name() string
}

// 具体策略
type OpenAIProvider struct { /* ... */ }
type AnthropicProvider struct { /* ... */ }
type GoogleProvider struct { /* ... */ }

// 上下文
type TranslationEngine struct {
    provider AIProvider  // 可替换的策略
}
```

#### 2.2.2 Factory Pattern (提供商创建)

```go
type ProviderFactory interface {
    CreateProvider(providerType string, config ProviderConfig) (AIProvider, error)
}

type DefaultProviderFactory struct {}

func (f *DefaultProviderFactory) CreateProvider(
    providerType string, 
    config ProviderConfig,
) (AIProvider, error) {
    switch providerType {
    case "openai":
        return NewOpenAIProvider(config), nil
    case "anthropic":
        return NewAnthropicProvider(config), nil
    case "google":
        return NewGoogleProvider(config), nil
    default:
        return nil, fmt.Errorf("unknown provider: %s", providerType)
    }
}
```

#### 2.2.3 Repository Pattern (术语存储)

```go
type TerminologyRepository interface {
    Load(path string) (*Terminology, error)
    Save(path string, terminology *Terminology) error
    Exists(path string) bool
}

type JSONTerminologyRepository struct {}

func (r *JSONTerminologyRepository) Load(path string) (*Terminology, error) {
    // 实现 JSON 文件加载
}

func (r *JSONTerminologyRepository) Save(path string, terminology *Terminology) error {
    // 实现 JSON 文件保存
}
```

#### 2.2.4 Decorator Pattern (格式保护)

```go
type Translator interface {
    Translate(ctx context.Context, text string) (string, error)
}

// 基础翻译器
type BaseTranslator struct {
    provider AIProvider
}

// 格式保护装饰器
type FormatProtectionDecorator struct {
    translator Translator
    protector  FormatProtector
}

func (d *FormatProtectionDecorator) Translate(ctx context.Context, text string) (string, error) {
    // 1. 提取格式元素
    elements := d.protector.Extract(text)
    
    // 2. 调用被装饰的翻译器
    translated, err := d.translator.Translate(ctx, text)
    if err != nil {
        return "", err
    }
    
    // 3. 验证格式完整性
    if err := d.protector.Validate(elements, translated); err != nil {
        return "", err
    }
    
    return translated, nil
}
```

#### 2.2.5 Chain of Responsibility (翻译管道)

```go
type TranslationHandler interface {
    Handle(ctx context.Context, req *TranslationRequest) (*TranslationResponse, error)
    SetNext(handler TranslationHandler) TranslationHandler
}

type BaseHandler struct {
    next TranslationHandler
}

func (h *BaseHandler) SetNext(handler TranslationHandler) TranslationHandler {
    h.next = handler
    return handler
}

// 术语处理器
type TerminologyHandler struct {
    BaseHandler
    termManager TerminologyManager
}

// 格式保护处理器
type FormatProtectionHandler struct {
    BaseHandler
    protector FormatProtector
}

// 翻译处理器
type TranslationHandler struct {
    BaseHandler
    translator Translator
}

// 构建管道
func BuildPipeline() TranslationHandler {
    terminology := &TerminologyHandler{}
    format := &FormatProtectionHandler{}
    translation := &TranslationHandler{}
    
    terminology.SetNext(format).SetNext(translation)
    
    return terminology
}
```

### 2.3 依赖注入

使用 Wire 或手动依赖注入：

```go
// 应用程序容器
type App struct {
    TranslationService *TranslationService
    TerminologyManager TerminologyManager
    ConfigManager      *ConfigManager
}

// 依赖注入构造函数
func NewApp(cfg *Config) (*App, error) {
    // 创建 AI 提供商
    provider, err := NewProviderFromConfig(cfg)
    if err != nil {
        return nil, err
    }

    // 创建术语管理器
    termRepo := &JSONTerminologyRepository{}
    termManager := NewTerminologyManager(termRepo, provider)

    // 创建翻译引擎
    engine := NewTranslationEngine(provider, termManager)

    // 创建翻译服务
    service := NewTranslationService(engine, termManager)

    return &App{
        TranslationService: service,
        TerminologyManager: termManager,
        ConfigManager:      NewConfigManager(cfg),
    }, nil
}
```

### 2.4 增量翻译和 Key 过滤的集成流程

增量翻译和 Key 过滤是 v1.0.0 的核心功能，与整体翻译流程深度集成：

```
用户请求
    │
    ├─ 命令行参数解析
    │   ├─ --keys "settings.*,user.*"      (Key 过滤)
    │   ├─ --exclude-keys "admin.*"         (Key 过滤)
    │   └─ --force                          (强制完整翻译)
    │
    ▼
┌──────────────────────────────────────────────────────┐
│            1. 加载源文件和目标文件                      │
│  - 读取源文件 en.json                                 │
│  - 尝试读取目标文件 zh.json（如果存在）                 │
└──────────────────────────────────┬───────────────────┘
                                   │
                                   ▼
┌──────────────────────────────────────────────────────┐
│            2. Key 过滤（如果指定）                      │
│  KeyFilter.FilterKeys()                              │
│  - 解析 --keys 和 --exclude-keys 模式                 │
│  - 递归遍历 JSON 树，匹配模式                          │
│  - 返回过滤后的 keys                                  │
└──────────────────────────────────┬───────────────────┘
                                   │
                                   ▼
┌──────────────────────────────────────────────────────┐
│       3. 差异分析（如果目标文件存在且未 --force）        │
│  IncrementalTranslator.AnalyzeDiff()                 │
│  - 对比源文件和目标文件                                │
│  - 识别新增、修改、删除、未变更的 keys                  │
│  - 考虑 Key 过滤的结果                                 │
│  - 返回差异报告                                       │
└──────────────────────────────────┬───────────────────┘
                                   │
                    ┌──────────────┴──────────────┐
                    │                             │
       无差异或只有删除                有新增或修改
                    │                             │
                    ▼                             ▼
           ┌─────────────────┐         ┌─────────────────┐
           │ 无需翻译         │         │ 需要翻译         │
           │ - 更新目标文件   │         │ - 翻译新增/修改   │
           │ - 删除多余 keys  │         │   的 keys        │
           └─────────────────┘         └────────┬────────┘
                    │                            │
                    │                            ▼
                    │              ┌──────────────────────────┐
                    │              │   4. 术语管理             │
                    │              │  - 检测术语（可选）        │
                    │              │  - 翻译缺失的术语          │
                    │              │  - 构建术语字典           │
                    │              └────────┬─────────────────┘
                    │                       │
                    │                       ▼
                    │              ┌──────────────────────────┐
                    │              │   5. 批量翻译             │
                    │              │  - 创建批次               │
                    │              │  - 并发调用 AI Provider   │
                    │              │  - 格式保护验证           │
                    │              └────────┬─────────────────┘
                    │                       │
                    │                       ▼
                    │              ┌──────────────────────────┐
                    │              │   6. 合并结果             │
                    │              │  - 新翻译的内容           │
                    │              │  - 未变更的翻译           │
                    │              │  - 删除多余的 keys        │
                    │              └────────┬─────────────────┘
                    │                       │
                    └───────────────────────┘
                                   │
                                   ▼
                    ┌──────────────────────────┐
                    │   7. 写入目标文件         │
                    │  - 保持 JSON 结构         │
                    │  - 保留代码格式           │
                    └──────────┬───────────────┘
                               │
                               ▼
                    ┌──────────────────────────┐
                    │   8. 输出报告             │
                    │  - 翻译统计               │
                    │  - 增量统计               │
                    │  - Key 过滤统计           │
                    │  - 成本和耗时             │
                    └──────────────────────────┘
```

**关键集成点**:

1. **Key 过滤优先**: 在差异分析之前应用 Key 过滤，确保只分析需要翻译的 keys

2. **差异分析在过滤后**: 差异分析只对过滤后的 keys 进行，避免不必要的计算

3. **智能决策逻辑**:
   ```go
   if force {
       // 完整翻译所有过滤后的 keys
       translateAll(filteredKeys)
   } else if targetExists {
       diff := analyzeDiff(source, target, filteredKeys)
       if diff.HasChanges() {
           // 只翻译变更的 keys
           translateIncremental(diff.New, diff.Modified)
           merge(translated, diff.Unchanged, diff.Deleted)
       } else {
           // 无需翻译，只更新目标文件（删除多余 keys）
           updateTarget(diff.Unchanged, diff.Deleted)
       }
   } else {
       // 首次翻译
       translateAll(filteredKeys)
   }
   ```

4. **统计信息累计**: 各个阶段的统计信息会累计，最终输出完整报告：
   ```
   ✅ 翻译完成

   Key 过滤:
     - 总 keys: 100
     - 匹配: 30 keys (settings.*, user.*)
     - 排除: 10 keys (admin.*)
     - 实际处理: 20 keys

   增量分析:
     - 新增: 5 keys
     - 修改: 2 keys
     - 删除: 3 keys
     - 保持: 10 keys

   翻译:
     - 翻译: 7 keys
     - 保留: 10 keys
     - 删除: 3 keys
     - API 费用: ~$0.05 (节省 90%)
     - 耗时: 3 秒
   ```

**优势**:

- **高效**: Key 过滤 + 增量翻译可以大幅减少 API 调用
- **灵活**: 支持各种组合使用场景
- **透明**: 详细的统计信息让用户了解每一步发生了什么
- **安全**: 保留用户手动修改的翻译（如果源文本未变更）

---

## 3. 核心接口设计

### 3.1 AI Provider Interface

```go
package provider

import "context"

// CompletionRequest 表示完成请求
type CompletionRequest struct {
    Prompt      string
    Model       string
    Temperature float32
    MaxTokens   int
    SystemMsg   string
}

// CompletionResponse 表示完成响应
type CompletionResponse struct {
    Content      string
    FinishReason string
    Usage        Usage
}

type Usage struct {
    PromptTokens     int
    CompletionTokens int
    TotalTokens      int
}

// AIProvider 定义 AI 提供商接口
type AIProvider interface {
    // Complete 执行文本完成
    Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error)
    
    // Name 返回提供商名称
    Name() string
    
    // GetModelName 返回当前使用的模型名称
    GetModelName() string
    
    // ValidateConfig 验证配置
    ValidateConfig() error
}
```

#### 3.1.1 OpenAI Provider 实现（官方 SDK）

```go
package provider

import (
    "context"
    "fmt"
    
    "github.com/openai/openai-go/v3"
    "github.com/openai/openai-go/v3/option"
)

type OpenAIProvider struct {
    client    *openai.Client
    apiKey    string
    modelName string
}

func NewOpenAIProvider(apiKey string, modelName string) (*OpenAIProvider, error) {
    if apiKey == "" {
        return nil, fmt.Errorf("OpenAI API key is required")
    }
    
    if modelName == "" {
        modelName = "gpt-4o" // 默认模型
    }
    
    client := openai.NewClient(
        option.WithAPIKey(apiKey),
    )
    
    return &OpenAIProvider{
        client:    client,
        apiKey:    apiKey,
        modelName: modelName,
    }, nil
}

func (p *OpenAIProvider) Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
    // 构建 messages
    messages := []openai.ChatCompletionMessageParamUnion{}
    
    // 添加系统消息（如果有）
    if req.SystemMsg != "" {
        messages = append(messages, openai.SystemMessage(req.SystemMsg))
    }
    
    // 添加用户消息
    messages = append(messages, openai.UserMessage(req.Prompt))
    
    // 调用 API
    chatCompletion, err := p.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
        Messages:    messages,
        Model:       openai.ChatModel(req.Model),
        Temperature: openai.Float(float64(req.Temperature)),
        MaxTokens:   openai.Int(int64(req.MaxTokens)),
    })
    
    if err != nil {
        return nil, fmt.Errorf("OpenAI API call failed: %w", err)
    }
    
    // 解析响应
    if len(chatCompletion.Choices) == 0 {
        return nil, fmt.Errorf("no response from OpenAI")
    }
    
    return &CompletionResponse{
        Content:      chatCompletion.Choices[0].Message.Content,
        FinishReason: string(chatCompletion.Choices[0].FinishReason),
        Usage: Usage{
            PromptTokens:     int(chatCompletion.Usage.PromptTokens),
            CompletionTokens: int(chatCompletion.Usage.CompletionTokens),
            TotalTokens:      int(chatCompletion.Usage.TotalTokens),
        },
    }, nil
}

func (p *OpenAIProvider) Name() string {
    return "openai"
}

func (p *OpenAIProvider) GetModelName() string {
    return p.modelName
}

func (p *OpenAIProvider) ValidateConfig() error {
    if p.apiKey == "" {
        return fmt.Errorf("OpenAI API key is required")
    }
    return nil
}
```

#### 3.1.2 Anthropic Provider 实现（官方 SDK）

```go
package provider

import (
    "context"
    "fmt"
    
    "github.com/anthropics/anthropic-sdk-go"
    "github.com/anthropics/anthropic-sdk-go/option"
)

type AnthropicProvider struct {
    client    *anthropic.Client
    apiKey    string
    modelName string
}

func NewAnthropicProvider(apiKey string, modelName string) (*AnthropicProvider, error) {
    if apiKey == "" {
        return nil, fmt.Errorf("Anthropic API key is required")
    }
    
    if modelName == "" {
        modelName = "claude-3-5-sonnet-20250116" // 默认模型
    }
    
    client := anthropic.NewClient(
        option.WithAPIKey(apiKey),
    )
    
    return &AnthropicProvider{
        client:    client,
        apiKey:    apiKey,
        modelName: modelName,
    }, nil
}

func (p *AnthropicProvider) Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
    // 构建参数
    params := anthropic.MessageNewParams{
        Model:       anthropic.Model(req.Model),
        MaxTokens:   int64(req.MaxTokens),
        Temperature: anthropic.Float(float64(req.Temperature)),
        Messages: []anthropic.MessageParam{
            anthropic.NewUserMessage(anthropic.NewTextBlock(req.Prompt)),
        },
    }
    
    // 添加系统提示（如果有）
    if req.SystemMsg != "" {
        params.System = []anthropic.TextBlockParam{
            {Text: anthropic.String(req.SystemMsg)},
        }
    }
    
    // 调用 API
    message, err := p.client.Messages.New(ctx, params)
    if err != nil {
        return nil, fmt.Errorf("Anthropic API call failed: %w", err)
    }
    
    // 提取文本内容
    var content string
    for _, block := range message.Content {
        if block.Type == anthropic.ContentBlockTypeText {
            content += block.Text
        }
    }
    
    return &CompletionResponse{
        Content:      content,
        FinishReason: string(message.StopReason),
        Usage: Usage{
            PromptTokens:     int(message.Usage.InputTokens),
            CompletionTokens: int(message.Usage.OutputTokens),
            TotalTokens:      int(message.Usage.InputTokens + message.Usage.OutputTokens),
        },
    }, nil
}

func (p *AnthropicProvider) Name() string {
    return "anthropic"
}

func (p *AnthropicProvider) GetModelName() string {
    return p.modelName
}

func (p *AnthropicProvider) ValidateConfig() error {
    if p.apiKey == "" {
        return fmt.Errorf("Anthropic API key is required")
    }
    return nil
}
```

#### 3.1.3 Google Gemini Provider 实现（官方 SDK）

```go
package provider

import (
    "context"
    "fmt"
    
    "google.golang.org/genai"
)

type GeminiProvider struct {
    client    *genai.Client
    apiKey    string
    modelName string
}

func NewGeminiProvider(ctx context.Context, apiKey string, modelName string) (*GeminiProvider, error) {
    if apiKey == "" {
        return nil, fmt.Errorf("Gemini API key is required")
    }
    
    if modelName == "" {
        modelName = "gemini-2.0-flash-exp" // 默认模型
    }
    
    // 初始化客户端
    client, err := genai.NewClient(ctx, &genai.ClientConfig{
        APIKey:  apiKey,
        Backend: genai.BackendGeminiAPI,
    })
    
    if err != nil {
        return nil, fmt.Errorf("failed to create Gemini client: %w", err)
    }
    
    return &GeminiProvider{
        client:    client,
        apiKey:    apiKey,
        modelName: modelName,
    }, nil
}

func (p *GeminiProvider) Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
    // 构建内容
    var contents []*genai.Content
    
    // 用户消息
    contents = append(contents, &genai.Content{
        Role: "user",
        Parts: []*genai.Part{
            genai.NewTextPart(req.Prompt),
        },
    })
    
    // 构建配置
    temperature := float32(req.Temperature)
    maxTokens := int32(req.MaxTokens)
    
    config := &genai.GenerateContentConfig{
        GenerationConfig: &genai.GenerationConfig{
            Temperature:     &temperature,
            MaxOutputTokens: &maxTokens,
        },
    }
    
    // 添加系统提示（如果有）
    if req.SystemMsg != "" {
        config.SystemInstruction = &genai.Content{
            Role: "system",
            Parts: []*genai.Part{
                genai.NewTextPart(req.SystemMsg),
            },
        }
    }
    
    // 调用 API
    result, err := p.client.Models.GenerateContent(
        ctx,
        req.Model,
        contents,
        config,
    )
    
    if err != nil {
        return nil, fmt.Errorf("Gemini API call failed: %w", err)
    }
    
    // 提取文本内容
    text, err := result.Text()
    if err != nil {
        return nil, fmt.Errorf("failed to extract text from Gemini response: %w", err)
    }
    
    // 提取 token 使用信息
    var promptTokens, completionTokens int
    if result.UsageMetadata != nil {
        promptTokens = int(result.UsageMetadata.PromptTokenCount)
        completionTokens = int(result.UsageMetadata.CandidatesTokenCount)
    }
    
    return &CompletionResponse{
        Content:      text,
        FinishReason: "stop", // Gemini 的 finish reason 映射
        Usage: Usage{
            PromptTokens:     promptTokens,
            CompletionTokens: completionTokens,
            TotalTokens:      promptTokens + completionTokens,
        },
    }, nil
}

func (p *GeminiProvider) Name() string {
    return "google"
}

func (p *GeminiProvider) GetModelName() string {
    return p.modelName
}

func (p *GeminiProvider) ValidateConfig() error {
    if p.apiKey == "" {
        return fmt.Errorf("Gemini API key is required")
    }
    return nil
}

func (p *GeminiProvider) Close() error {
    return p.client.Close()
}
```

#### 3.1.4 Provider Factory

```go
package provider

import (
    "context"
    "fmt"
)

type ProviderType string

const (
    ProviderTypeOpenAI    ProviderType = "openai"
    ProviderTypeAnthropic ProviderType = "anthropic"
    ProviderTypeGoogle    ProviderType = "google"
)

type ProviderConfig struct {
    Type   ProviderType
    APIKey string
    Model  string
}

func NewProvider(ctx context.Context, config *ProviderConfig) (AIProvider, error) {
    // 如果没有指定模型，使用默认模型
    modelName := config.Model
    if modelName == "" {
        modelName = GetDefaultModel(config.Type)
    }
    
    switch config.Type {
    case ProviderTypeOpenAI:
        return NewOpenAIProvider(config.APIKey, modelName)
        
    case ProviderTypeAnthropic:
        return NewAnthropicProvider(config.APIKey, modelName)
        
    case ProviderTypeGoogle:
        return NewGeminiProvider(ctx, config.APIKey, modelName)
        
    default:
        return nil, fmt.Errorf("unsupported provider type: %s", config.Type)
    }
}

// GetDefaultModel 返回 provider 的默认模型
func GetDefaultModel(providerType ProviderType) string {
    switch providerType {
    case ProviderTypeOpenAI:
        return "gpt-4o"
    case ProviderTypeAnthropic:
        return "claude-3-5-sonnet-20250116"
    case ProviderTypeGoogle:
        return "gemini-2.0-flash-exp"
    default:
        return ""
    }
}

// GetContextWindowSize 返回模型的 context window 大小
func GetContextWindowSize(providerType ProviderType) int {
    switch providerType {
    case ProviderTypeOpenAI:
        return 128000  // GPT-4o: 128K tokens
    case ProviderTypeAnthropic:
        return 200000  // Claude 3.5 Sonnet: 200K tokens
    case ProviderTypeGoogle:
        return 1000000 // Gemini 2.0 Flash: 1M tokens
    default:
        return 100000  // 保守估计
    }
}
```

**环境变量支持**:

```go
package provider

import (
    "context"
    "fmt"
    "os"
)

// NewProviderFromEnv 从环境变量创建 provider
// 可选参数 modelName，如果为空则使用默认模型
func NewProviderFromEnv(ctx context.Context, providerType ProviderType, modelName string) (AIProvider, error) {
    var apiKey string
    
    switch providerType {
    case ProviderTypeOpenAI:
        apiKey = os.Getenv("OPENAI_API_KEY")
        if apiKey == "" {
            return nil, fmt.Errorf("OPENAI_API_KEY environment variable not set")
        }
        
    case ProviderTypeAnthropic:
        apiKey = os.Getenv("ANTHROPIC_API_KEY")
        if apiKey == "" {
            return nil, fmt.Errorf("ANTHROPIC_API_KEY environment variable not set")
        }
        
    case ProviderTypeGoogle:
        apiKey = os.Getenv("GEMINI_API_KEY")
        if apiKey == "" {
            // Fallback to GOOGLE_API_KEY
            apiKey = os.Getenv("GOOGLE_API_KEY")
        }
        if apiKey == "" {
            return nil, fmt.Errorf("GEMINI_API_KEY or GOOGLE_API_KEY environment variable not set")
        }
        
    default:
        return nil, fmt.Errorf("unsupported provider type: %s", providerType)
    }
    
    return NewProvider(ctx, &ProviderConfig{
        Type:   providerType,
        APIKey: apiKey,
        Model:  modelName, // 会在 NewProvider 中处理空值
    })
}
```

### 3.2 Terminology Manager Interface

```go
package terminology

import "context"

// Term 表示一个术语
type Term struct {
    Term     string
    Type     TermType  // Preserve 或 Consistent
    Context  string    // LLM 提供的上下文说明
    Reason   string    // 为什么检测为术语
}

type TermType string

const (
    TermTypePreserve   TermType = "preserve"
    TermTypeConsistent TermType = "consistent"
)

// Terminology 表示术语集合
type Terminology struct {
    SourceLanguage string            `json:"sourceLanguage"`
    PreserveTerms  []string          `json:"preserveTerms"`
    ConsistentTerms map[string][]string `json:"consistentTerms"`
}

// TerminologyManager 定义术语管理接口
type TerminologyManager interface {
    // DetectTerms 使用 LLM 检测术语（可跳过）
    DetectTerms(ctx context.Context, texts []string, sourceLang string) ([]Term, error)
    
    // LoadTerminology 加载术语文件
    LoadTerminology(path string) (*Terminology, error)
    
    // SaveTerminology 保存术语文件
    SaveTerminology(path string, terminology *Terminology) error
    
    // TranslateTerms 翻译术语（总是执行，确保所有术语都有目标语言翻译）
    TranslateTerms(ctx context.Context, terms []string, targetLang string) (map[string]string, error)
    
    // GetTermTranslation 获取术语翻译
    GetTermTranslation(term string, targetLang string) (string, bool)
    
    // GetMissingTranslations 获取缺失的术语翻译
    GetMissingTranslations(targetLang string) []string
    
    // BuildPromptDictionary 构建用于 prompt 的术语字典
    BuildPromptDictionary(targetLang string) string
}
```

#### 3.2.1 术语检测实现详解

术语检测是 Jta 的核心 Agentic 能力之一。我们采用**分层策略**：对于小文件使用纯 LLM 分析，对于大文件使用**统计预处理 + LLM 验证**的混合方案。

**核心类型定义**:

```go
package terminology

// Detector 术语检测器
type Detector struct {
    provider  provider.AIProvider
    maxTokens int // LLM 的 context window 限制
}

// CandidateWord 候选术语（用于大文件场景）
type CandidateWord struct {
    Word      string   // 候选词
    Frequency int      // 出现频率
    Contexts  []string // 该词出现的上下文（最多 5 个）
}
```

**策略选择逻辑**:

```go
const (
    // MAX_CONTEXT_TOKENS 术语检测的最大上下文 token 数
    // 设置为 10K：保守估计，适用于所有主流模型（GPT-3.5+ 都支持 16K+）
    // 大约可以处理 2500 条平均长度的 i18n 文本（假设每条 25 tokens）
    MAX_CONTEXT_TOKENS = 10000
    
    // CONTEXT_USAGE_RATIO 实际使用的上下文比例（预留空间给 prompt 和输出）
    CONTEXT_USAGE_RATIO = 0.7
)

func (d *Detector) DetectTerms(ctx context.Context, texts []string, sourceLang string) ([]Term, error) {
    // 1. 估算 token 数
    estimatedTokens := d.estimateTokens(texts)
    
    // 2. 计算可用的 token 数（70% 用于文本，30% 用于 prompt 和输出）
    maxUsableTokens := int(float64(MAX_CONTEXT_TOKENS) * CONTEXT_USAGE_RATIO)
    
    // 3. 根据文件大小选择策略
    if estimatedTokens <= maxUsableTokens {
        // 策略 A: 小文件（< 7K tokens）- 纯 LLM 一次性分析
        // 适用于 95% 的场景
        log.Println("📊 Using full-context LLM analysis...")
        return d.analyzeWithLLM(ctx, texts, sourceLang)
    }
    
    // 策略 B: 大文件（> 7K tokens）- 混合方案（统计 + LLM）
    // 只在极少数场景使用
    log.Printf("📊 File too large (%d tokens), using hybrid approach...", estimatedTokens)
    return d.hybridDetection(ctx, texts, sourceLang)
}

func (d *Detector) estimateTokens(texts []string) int {
    totalChars := 0
    for _, text := range texts {
        totalChars += len(text)
    }
    // 粗略估算：英文平均 4 chars per token
    return totalChars / 4
}
```

**策略 A: 纯 LLM 分析（小文件，< 70% context window）**:

```go
func (d *Detector) analyzeWithLLM(ctx context.Context, texts []string, lang string) ([]Term, error) {
    // 1. 构建完整文档
    doc := d.buildFullDocument(texts)
    
    // 2. 构建 prompt
    prompt := d.buildDetectionPrompt(doc, lang, len(texts))
    
    // 3. 调用 LLM（只需一次）
    resp, err := d.provider.Complete(ctx, &provider.CompletionRequest{
        Prompt:      prompt,
        Temperature: 0.3,
        MaxTokens:   2000, // 输出不需要太长
    })
    
    if err != nil {
        return nil, fmt.Errorf("LLM analysis failed: %w", err)
    }
    
    // 4. 解析结果
    return d.parseTermsFromJSON(resp.Content)
}

func (d *Detector) buildFullDocument(texts []string) string {
    var builder strings.Builder
    builder.WriteString(fmt.Sprintf("Total texts: %d\n\n", len(texts)))
    
    for i, text := range texts {
        builder.WriteString(fmt.Sprintf("[%d] %s\n", i+1, text))
    }
    
    return builder.String()
}

func (d *Detector) buildDetectionPrompt(doc string, lang string, totalCount int) string {
    return fmt.Sprintf(`You are an expert terminology analyst for JSON internationalization files.

Your task: Analyze this COMPLETE %s JSON i18n file (containing %d texts) and identify terms that need special handling for translation consistency.

<DOCUMENT>
%s
</DOCUMENT>

Analysis Instructions:
1. Read through the ENTIRE document carefully
2. Notice which terms appear MULTIPLE TIMES in different contexts
3. Consider term importance based on:
   - Frequency of occurrence
   - Context (technical, business, branding)
   - Impact on translation consistency

Identify TWO types of terms:

A. PRESERVE (never translate):
   - Brand names (e.g., "MyApp", "OpenAI")
   - Technical terms (e.g., "API", "OAuth", "JSON")
   - Product names with versions (e.g., "FLUX.1", "GPT-4")
   - Proper nouns

B. CONSISTENT (must translate uniformly):
   - Business domain terms appearing multiple times
   - Core concepts specific to this application
   - Terms where inconsistent translation would confuse users

Response Format (JSON only, no explanation):
{
  "preserveTerms": [
    {
      "term": "API",
      "reason": "Technical acronym",
      "frequency": 15,
      "examples": ["API key", "API access", "API documentation"]
    }
  ],
  "consistentTerms": [
    {
      "term": "credits",
      "reason": "Core business concept",
      "frequency": 23,
      "examples": ["You have 10 credits", "Buy credits", "Unlimited credits"]
    }
  ]
}

Important:
- Only include terms that appear in the document
- Provide accurate frequency counts
- Include 2-3 example usages for each term
- Focus on quality over quantity (typically 5-15 terms total)`, lang, totalCount, doc)
}
```

**策略 B: 混合方案（大文件，> 70% context window）**:

这个方案分为三个步骤：

**步骤 1: 本地统计分析（无 LLM 调用）**:

```go
func (d *Detector) hybridDetection(ctx context.Context, texts []string, lang string) ([]Term, error) {
    // 第 1 步：本地统计分析（简化版，不依赖外部 NLP 库）
    log.Println("Step 1/3: Extracting candidate terms (local analysis)...")
    candidates := d.extractCandidatesSimplified(texts)
    log.Printf("Found %d candidates\n", len(candidates))
    
    // 第 2 步：LLM 批量验证
    log.Println("Step 2/3: Validating candidates with LLM...")
    return d.validateWithLLM(ctx, candidates, lang)
}

func (d *Detector) extractCandidatesSimplified(texts []string) map[string]*CandidateWord {
    candidates := make(map[string]*CandidateWord)
    
    for _, text := range texts {
        // 简单分词（按空格和标点）
        words := d.simpleTokenize(text)
        
        // 提取 1-3 个词的短语
        for i := 0; i < len(words); i++ {
            // 单词
            d.addCandidate(candidates, words[i], text)
            
            // 二词组（bigram）
            if i+1 < len(words) {
                phrase := words[i] + " " + words[i+1]
                d.addCandidate(candidates, phrase, text)
            }
            
            // 三词组（trigram）
            if i+2 < len(words) {
                phrase := words[i] + " " + words[i+1] + " " + words[i+2]
                d.addCandidate(candidates, phrase, text)
            }
        }
    }
    
    // 过滤：只保留满足条件的候选词
    return d.filterCandidates(candidates)
}

func (d *Detector) addCandidate(candidates map[string]*CandidateWord, word string, context string) {
    // 跳过太短或太长的
    if len(word) < 2 || len(word) > 50 {
        return
    }
    
    // 跳过停用词
    if d.isStopWord(word) {
        return
    }
    
    word = strings.TrimSpace(word)
    
    if cand, exists := candidates[word]; exists {
        cand.Frequency++
        // 只保留前 5 个上下文（避免内存爆炸）
        if len(cand.Contexts) < 5 {
            cand.Contexts = append(cand.Contexts, context)
        }
    } else {
        candidates[word] = &CandidateWord{
            Word:      word,
            Frequency: 1,
            Contexts:  []string{context},
        }
    }
}

func (d *Detector) filterCandidates(candidates map[string]*CandidateWord) map[string]*CandidateWord {
    filtered := make(map[string]*CandidateWord)
    
    for word, info := range candidates {
        // 保留条件：
        // 1. 频率 >= 3（高频词）
        // 2. 或者是特殊格式（全大写、包含版本号等）
        if info.Frequency >= 3 || d.isSpecialFormat(word) {
            filtered[word] = info
        }
    }
    
    return filtered
}

func (d *Detector) simpleTokenize(text string) []string {
    // 简单的分词（不依赖外部库）
    // 替换标点为空格（但保留连字符和点号）
    text = strings.Map(func(r rune) rune {
        if r == '-' || r == '.' || unicode.IsLetter(r) || unicode.IsDigit(r) {
            return r
        }
        if unicode.IsSpace(r) || unicode.IsPunct(r) {
            return ' '
        }
        return r
    }, text)
    
    words := strings.Fields(text)
    
    // 转小写（除了全大写词，如 API）
    result := []string{}
    for _, word := range words {
        if word == strings.ToUpper(word) && len(word) >= 2 {
            result = append(result, word) // 保持全大写
        } else {
            result = append(result, strings.ToLower(word))
        }
    }
    
    return result
}

func (d *Detector) isStopWord(word string) bool {
    // 简单的英文停用词列表
    stopWords := map[string]bool{
        "the": true, "a": true, "an": true, "and": true, "or": true,
        "but": true, "in": true, "on": true, "at": true, "to": true,
        "for": true, "of": true, "with": true, "by": true, "from": true,
        "is": true, "are": true, "was": true, "were": true, "be": true,
        "this": true, "that": true, "these": true, "those": true,
        "your": true, "you": true, "it": true, "its": true,
    }
    
    return stopWords[strings.ToLower(word)]
}

func (d *Detector) isSpecialFormat(word string) bool {
    // 全大写（如 API, JSON）
    if len(word) >= 2 && word == strings.ToUpper(word) && !strings.ContainsAny(word, " ") {
        return true
    }
    
    // 包含版本号（如 FLUX.1, GPT-4）
    if strings.Contains(word, ".") || strings.ContainsAny(word, "0123456789") {
        return true
    }
    
    // 驼峰命名（如 MyApp, OpenAI）
    if len(word) > 1 && unicode.IsUpper(rune(word[0])) {
        for i := 1; i < len(word); i++ {
            if unicode.IsUpper(rune(word[i])) {
                return true
            }
        }
    }
    
    return false
}
```

**步骤 2: LLM 批量验证**:

```go
func (d *Detector) validateWithLLM(ctx context.Context, candidates map[string]*CandidateWord, lang string) ([]Term, error) {
    // 将候选词分批处理（每批 30 个）
    batches := d.batchCandidates(candidates, 30)
    
    allTerms := []Term{}
    
    for i, batch := range batches {
        log.Printf("Validating batch %d/%d (%d candidates)...", i+1, len(batches), len(batch))
        
        terms, err := d.validateBatchWithLLM(ctx, batch, lang)
        if err != nil {
            return nil, fmt.Errorf("batch %d validation failed: %w", i+1, err)
        }
        
        allTerms = append(allTerms, terms...)
    }
    
    log.Println("Step 3/3: Validation complete")
    return allTerms, nil
}

func (d *Detector) validateBatchWithLLM(ctx context.Context, batch []*CandidateWord, lang string) ([]Term, error) {
    prompt := d.buildValidationPrompt(batch, lang)
    
    resp, err := d.provider.Complete(ctx, &provider.CompletionRequest{
        Prompt:      prompt,
        Temperature: 0.3,
        MaxTokens:   3000,
    })
    
    if err != nil {
        return nil, err
    }
    
    return d.parseValidationResult(resp.Content)
}

func (d *Detector) buildValidationPrompt(candidates []*CandidateWord, lang string) string {
    var builder strings.Builder
    
    builder.WriteString(fmt.Sprintf(`You are a terminology validation expert for JSON i18n files.

I have extracted candidate terms from a large %s JSON file using statistical analysis.
Your task: Verify which candidates are TRUE TERMS that need special handling for translation.

TRUE TERMS are:
1. PRESERVE (never translate): brand names, technical terms, product names, proper nouns
2. CONSISTENT (must translate uniformly): business domain terms, core concepts

NOT TERMS (ignore these):
- Common words that don't need special handling
- Generic phrases
- Complete sentences

Below are the candidates with their frequency and example contexts:

`, lang))
    
    for i, cand := range candidates {
        builder.WriteString(fmt.Sprintf("\n%d. Candidate: \"%s\"\n", i+1, cand.Word))
        builder.WriteString(fmt.Sprintf("   Frequency: %d times in file\n", cand.Frequency))
        builder.WriteString("   Example contexts:\n")
        for j, ctx := range cand.Contexts {
            if j >= 3 {
                break // 最多 3 个上下文
            }
            builder.WriteString(fmt.Sprintf("   - \"%s\"\n", ctx))
        }
    }
    
    builder.WriteString(`

Return JSON array with your decisions (ONLY include terms where is_term is true):
[
  {
    "term": "API",
    "is_term": true,
    "type": "preserve",
    "reason": "Technical acronym, appears in multiple technical contexts"
  },
  {
    "term": "user profile",
    "is_term": true,
    "type": "consistent",
    "reason": "Core UI feature name, appears frequently across different contexts"
  }
]`)
    
    return builder.String()
}

func (d *Detector) parseValidationResult(content string) ([]Term, error) {
    jsonStr := d.extractJSON(content)
    
    var results []struct {
        Term   string `json:"term"`
        IsTerm bool   `json:"is_term"`
        Type   string `json:"type"`
        Reason string `json:"reason"`
    }
    
    err := json.Unmarshal([]byte(jsonStr), &results)
    if err != nil {
        return nil, fmt.Errorf("failed to parse validation result: %w", err)
    }
    
    terms := []Term{}
    for _, r := range results {
        if !r.IsTerm {
            continue
        }
        
        var termType TermType
        if r.Type == "preserve" {
            termType = TermTypePreserve
        } else {
            termType = TermTypeConsistent
        }
        
        terms = append(terms, Term{
            Term:   r.Term,
            Type:   termType,
            Reason: r.Reason,
        })
    }
    
    return terms, nil
}

func (d *Detector) batchCandidates(candidates map[string]*CandidateWord, batchSize int) [][]*CandidateWord {
    batches := [][]*CandidateWord{}
    currentBatch := []*CandidateWord{}
    
    for _, cand := range candidates {
        currentBatch = append(currentBatch, cand)
        
        if len(currentBatch) >= batchSize {
            batches = append(batches, currentBatch)
            currentBatch = []*CandidateWord{}
        }
    }
    
    if len(currentBatch) > 0 {
        batches = append(batches, currentBatch)
    }
    
    return batches
}
```

**工具函数**:

```go
func (d *Detector) extractJSON(content string) string {
    // LLM 可能返回 markdown 格式的 JSON
    // 尝试提取 ```json ... ``` 之间的内容
    
    start := strings.Index(content, "```json")
    if start != -1 {
        start += 7 // len("```json")
        end := strings.Index(content[start:], "```")
        if end != -1 {
            return strings.TrimSpace(content[start : start+end])
        }
    }
    
    // 尝试提取 ``` ... ``` 之间的内容
    start = strings.Index(content, "```")
    if start != -1 {
        start += 3
        end := strings.Index(content[start:], "```")
        if end != -1 {
            return strings.TrimSpace(content[start : start+end])
        }
    }
    
    // 尝试找到 JSON 数组或对象的开始和结束
    content = strings.TrimSpace(content)
    if strings.HasPrefix(content, "[") || strings.HasPrefix(content, "{") {
        return content
    }
    
    // 查找第一个 { 或 [
    for i, c := range content {
        if c == '{' || c == '[' {
            return strings.TrimSpace(content[i:])
        }
    }
    
    return content
}
```

**策略对比**:

| 策略 | 适用场景 | Token 消耗 | API 调用次数 | 准确性 | 实现复杂度 |
|------|---------|-----------|------------|--------|----------|
| **纯 LLM** | < 70% context window | 高 | 1 次 | ⭐⭐⭐⭐⭐ | 低 |
| **混合方案** | > 70% context window | 中 | N 次（批次数） | ⭐⭐⭐⭐ | 中 |

**实际使用场景估算**:

```
文件大小示例（假设平均每条文本 25 tokens）：

小型应用（500 条）:   12,500 tokens  → 纯 LLM
中型应用（2,000 条）: 50,000 tokens  → 纯 LLM
大型应用（5,000 条）: 125,000 tokens → 纯 LLM (GPT-4o/Claude 3.5)
超大型（10,000 条）:  250,000 tokens → 混合方案 (GPT-4o) 或 纯 LLM (Gemini 2.0)

结论：99% 的场景使用纯 LLM，混合方案作为可靠的 fallback
```



### 3.3 Translator Interface

```go
package translator

import "context"

// TranslationInput 表示翻译输入
type TranslationInput struct {
    Source       map[string]interface{}  // 源 JSON
    SourceLang   string
    TargetLang   string
    Terminology  *terminology.Terminology
    Options      TranslationOptions
}

// TranslationOptions 表示翻译选项
type TranslationOptions struct {
    BatchSize       int
    Concurrency     int
    SkipTerms       bool  // 跳过术语检测（但仍翻译缺失的术语）
    NoTerminology   bool  // 完全不使用术语管理
}

// TranslationResult 表示翻译结果
type TranslationResult struct {
    Target      map[string]interface{}  // 翻译后的 JSON
    Stats       TranslationStats
    Errors      []TranslationError
}

// TranslationStats 表示翻译统计
type TranslationStats struct {
    TotalItems    int
    SuccessItems  int
    FailedItems   int
    Duration      time.Duration
    APICallsCount int
}

// Translator 定义翻译器接口
type Translator interface {
    // Translate 执行翻译
    Translate(ctx context.Context, input TranslationInput) (*TranslationResult, error)
}
```

### 3.4 Format Protector Interface

```go
package format

// FormatElement 表示格式元素
type FormatElement struct {
    Type     ElementType
    Value    string
    Position int
}

type ElementType string

const (
    ElementTypePlaceholder ElementType = "placeholder"
    ElementTypeHTML        ElementType = "html"
    ElementTypeURL         ElementType = "url"
    ElementTypeMarkdown    ElementType = "markdown"
)

// FormatProtector 定义格式保护接口
type FormatProtector interface {
    // Extract 提取格式元素
    Extract(text string) []FormatElement
    
    // Validate 验证格式完整性
    Validate(original string, translated string) error
    
    // GetValidationReport 获取验证报告
    GetValidationReport(original string, translated string) ValidationReport
}

// ValidationReport 表示验证报告
type ValidationReport struct {
    IsValid         bool
    MissingElements []FormatElement
    ExtraElements   []FormatElement
    Errors          []string
}
```

### 3.5 Batch Processor Interface

```go
package batch

import "context"

// Batch 表示一个批次
type Batch struct {
    Items []BatchItem
    Index int
}

// BatchItem 表示批次中的一个项目
type BatchItem struct {
    Key   string
    Text  string
    Context string
}

// BatchProcessor 定义批处理接口
type BatchProcessor interface {
    // CreateBatches 创建批次
    CreateBatches(items []BatchItem, batchSize int) []Batch

    // ProcessBatches 处理批次（并发）
    ProcessBatches(ctx context.Context, batches []Batch, concurrency int) (map[string]string, error)

    // ProcessSingleBatch 处理单个批次
    ProcessSingleBatch(ctx context.Context, batch Batch) (map[string]string, error)
}
```

### 3.6 Incremental Translator Interface

```go
package incremental

import "time"

// DiffResult 表示差异分析结果
type DiffResult struct {
    New      map[string]interface{}  // 新增的 keys
    Modified map[string]interface{}  // 修改的 keys（源文本变更）
    Deleted  []string                // 删除的 keys（源文件中已删除）
    Unchanged map[string]interface{} // 未变更的 keys
    Stats    DiffStats
}

// DiffStats 表示差异统计
type DiffStats struct {
    NewCount       int
    ModifiedCount  int
    DeletedCount   int
    UnchangedCount int
    TotalKeys      int
}

// IncrementalTranslator 定义增量翻译接口
type IncrementalTranslator interface {
    // AnalyzeDiff 分析源文件和目标文件的差异
    // source: 源文件内容（JSON）
    // target: 目标文件内容（JSON）- 如果不存在则为 nil
    // returns: 差异分析结果
    AnalyzeDiff(source, target map[string]interface{}) (*DiffResult, error)

    // ShouldTranslate 判断是否需要翻译（基于差异分析）
    // result: 差异分析结果
    // force: 是否强制完整翻译
    // returns: true 表示需要翻译
    ShouldTranslate(result *DiffResult, force bool) bool

    // MergeDiff 合并翻译结果和未变更的内容
    // translated: 新翻译的内容
    // unchanged: 未变更的内容
    // deleted: 需要删除的 keys
    // returns: 合并后的完整内容
    MergeDiff(translated, unchanged map[string]interface{}, deleted []string) map[string]interface{}
}

// ChangeDetector 变更检测器
type ChangeDetector interface {
    // DetectChanges 检测源文件是否有变更
    // sourcePath: 源文件路径
    // targetPath: 目标文件路径
    // returns: 是否有变更，最后修改时间
    DetectChanges(sourcePath, targetPath string) (hasChanges bool, sourceModTime time.Time, err error)

    // CompareTexts 比较两个文本是否相同
    // text1, text2: 要比较的文本
    // returns: true 表示相同
    CompareTexts(text1, text2 string) bool
}
```

**实现要点**:

1. **差异检测算法**：递归遍历 JSON 树，对比源文件和目标文件的每个 key-value 对
2. **智能判断变更**：
   - 如果 key 在源文件中存在但目标文件中不存在 → 新增
   - 如果 key 在两个文件中都存在，但源文本不同 → 修改
   - 如果 key 在目标文件中存在但源文件中不存在 → 删除
   - 如果 key 和源文本都相同 → 未变更（保留现有翻译）
3. **合并策略**：保留未变更的翻译，删除多余的 key，添加新翻译的内容
4. **文件时间戳检测**：快速判断文件是否有变更，避免不必要的差异分析

### 3.7 Key Filter Interface

```go
package keyfilter

// KeyPattern 表示 key 匹配模式
type KeyPattern struct {
    Pattern  string      // 原始模式字符串（如 "settings.*"）
    Type     PatternType // 模式类型
    Parts    []string    // 模式的各个部分
    IsGlob   bool        // 是否包含通配符
}

type PatternType string

const (
    PatternTypeExact     PatternType = "exact"     // 精确匹配 "settings.title"
    PatternTypeSingleLevel PatternType = "single"  // 单层通配符 "settings.*"
    PatternTypeRecursive PatternType = "recursive" // 递归通配符 "settings.**"
    PatternTypeWildcard  PatternType = "wildcard"  // 任意位置通配符 "*.title"
)

// FilterResult 表示过滤结果
type FilterResult struct {
    Included map[string]interface{} // 包含的 keys
    Excluded map[string]interface{} // 排除的 keys
    Stats    FilterStats
}

// FilterStats 表示过滤统计
type FilterStats struct {
    TotalKeys    int
    IncludedKeys int
    ExcludedKeys int
}

// KeyFilter 定义 key 过滤接口
type KeyFilter interface {
    // ParsePatterns 解析模式字符串（逗号分隔）
    // patterns: 模式字符串，如 "settings.*,user.profile.*"
    // returns: 解析后的模式列表
    ParsePatterns(patterns string) ([]*KeyPattern, error)

    // FilterKeys 过滤 JSON keys
    // data: 原始 JSON 数据
    // includePatterns: 包含模式（--keys）
    // excludePatterns: 排除模式（--exclude-keys）
    // returns: 过滤结果
    FilterKeys(
        data map[string]interface{},
        includePatterns []*KeyPattern,
        excludePatterns []*KeyPattern,
    ) (*FilterResult, error)

    // MatchKey 判断 key 是否匹配模式
    // keyPath: key 的完整路径，如 "settings.theme.dark"
    // pattern: 匹配模式
    // returns: true 表示匹配
    MatchKey(keyPath string, pattern *KeyPattern) bool

    // BuildKeyPath 构建 key 的完整路径
    // parts: key 的各个部分
    // returns: 完整路径字符串
    BuildKeyPath(parts []string) string
}

// Matcher 模式匹配器（具体实现）
type Matcher interface {
    // Match 执行匹配
    Match(keyPath string, pattern *KeyPattern) bool
}
```

**实现要点**:

1. **模式解析**：
   - `settings.*` → 匹配 `settings.title`, `settings.description`，但不匹配 `settings.theme.dark`
   - `settings.**` → 递归匹配 `settings` 下所有 key（包括嵌套）
   - `*.title` → 匹配所有名为 `title` 的 key（任意层级）
   - `settings.*.desc` → 匹配 `settings.user.desc`, `settings.admin.desc`

2. **匹配算法**：
   - 精确匹配：直接字符串比较
   - 单层通配符：Split by "."，按层级匹配
   - 递归通配符：前缀匹配
   - 任意位置通配符：正则表达式匹配

3. **优先级规则**：
   - 先应用 `--keys` 过滤（包含）
   - 再应用 `--exclude-keys` 过滤（排除）
   - `--exclude-keys` 优先级高于 `--keys`

4. **递归遍历**：遍历 JSON 树时构建完整的 key path，与模式进行匹配

5. **性能优化**：
   - 缓存编译后的模式（避免重复解析）
   - 提前终止不匹配的分支
   - 使用 map 存储结果（O(1) 查找）

**使用示例**:

```go
// 示例 1: 只翻译 settings 区域
filter := NewKeyFilter()
includePatterns, _ := filter.ParsePatterns("settings.*")
result, _ := filter.FilterKeys(data, includePatterns, nil)
// result.Included 只包含 settings.* 的 keys

// 示例 2: 排除 admin 和 internal 区域
excludePatterns, _ := filter.ParsePatterns("admin.*,internal.*")
result, _ := filter.FilterKeys(data, nil, excludePatterns)
// result.Included 包含除 admin.* 和 internal.* 外的所有 keys

// 示例 3: 组合使用
includePatterns, _ := filter.ParsePatterns("settings.*,user.*")
excludePatterns, _ := filter.ParsePatterns("settings.advanced.*")
result, _ := filter.FilterKeys(data, includePatterns, excludePatterns)
// 翻译 settings.* 和 user.*，但排除 settings.advanced.*
```

---

## 4. 项目结构

```
jta/
├── cmd/
│   └── jta/
│       └── main.go                    # 主入口
│
├── internal/                          # 内部包（不对外暴露）
│   │
│   ├── app/                           # 应用层
│   │   ├── app.go                     # 应用容器
│   │   └── service.go                 # 翻译服务
│   │
│   ├── cli/                           # CLI 层
│   │   ├── root.go                    # 根命令（即翻译命令，无子命令）
│   │   └── ui/                        # 终端 UI
│   │       ├── progress.go            # 进度条
│   │       ├── prompt.go              # 交互提示
│   │       └── output.go              # 输出格式化
│   │
│   ├── config/                        # 配置管理
│   │   ├── config.go                  # 配置结构
│   │   ├── loader.go                  # 配置加载器
│   │   └── validator.go               # 配置验证
│   │
│   ├── domain/                        # 领域模型
│   │   ├── translation.go             # 翻译领域模型
│   │   ├── terminology.go             # 术语领域模型
│   │   └── language.go                # 语言定义
│   │
│   ├── translator/                    # 翻译引擎
│   │   ├── engine.go                  # 翻译引擎
│   │   ├── batch.go                   # 批量翻译器
│   │   ├── agentic.go                 # Agentic 翻译（反思）
│   │   ├── pipeline.go                # 翻译管道
│   │   ├── context.go                 # 翻译上下文
│   │   └── incremental.go             # 增量翻译器
│   │
│   ├── terminology/                   # 术语管理
│   │   ├── manager.go                 # 术语管理器
│   │   ├── detector.go                # 术语检测器（LLM）
│   │   ├── translator.go              # 术语翻译器
│   │   ├── repository.go              # 术语仓储接口
│   │   └── json_repository.go         # JSON 仓储实现
│   │
│   ├── provider/                      # AI 提供商
│   │   ├── provider.go                # 提供商接口
│   │   ├── factory.go                 # 提供商工厂
│   │   ├── openai/
│   │   │   └── openai.go              # OpenAI 实现
│   │   ├── anthropic/
│   │   │   └── anthropic.go           # Anthropic 实现
│   │   ├── google/
│   │   │   └── google.go              # Google 实现
│   │   └── custom/
│   │       └── custom.go              # 自定义提供商
│   │
│   ├── format/                        # 格式保护
│   │   ├── protector.go               # 格式保护器
│   │   ├── extractor.go               # 格式提取器
│   │   ├── validator.go               # 格式验证器
│   │   └── patterns.go                # 格式匹配模式
│   │
│   ├── rtl/                           # RTL 处理
│   │   ├── processor.go               # RTL 处理器
│   │   ├── detector.go                # RTL 语言检测
│   │   └── formatter.go               # RTL 格式化
│   │
│   ├── keyfilter/                     # Key 过滤
│   │   ├── filter.go                  # Key 过滤器
│   │   ├── parser.go                  # 模式解析器
│   │   ├── matcher.go                 # 模式匹配器
│   │   └── pattern.go                 # 模式定义
│   │
│   ├── diff/                          # 差异分析
│   │   ├── analyzer.go                # 差异分析器
│   │   ├── detector.go                # 变更检测器
│   │   └── merger.go                  # 合并器
│   │
│   ├── prompt/                        # Prompt 模板
│   │   ├── template.go                # 模板引擎
│   │   ├── translation.go             # 翻译 prompts
│   │   ├── terminology.go             # 术语 prompts
│   │   └── reflection.go              # 反思 prompts
│   │
│   ├── validator/                     # 验证器
│   │   ├── validator.go               # 翻译验证器
│   │   ├── structure.go               # 结构验证
│   │   ├── terminology.go             # 术语验证
│   │   └── format.go                  # 格式验证
│   │
│   └── util/                          # 工具函数
│       ├── json.go                    # JSON 处理
│       ├── file.go                    # 文件处理
│       ├── logger.go                  # 日志工具
│       ├── retry.go                   # 重试逻辑
│       └── concurrent.go              # 并发控制
│
├── pkg/                               # 公共库（可对外暴露）
│   └── jtaclient/                     # Go 客户端库（可选）
│       └── client.go
│
├── scripts/                           # 脚本
│   ├── install.sh                     # Linux/macOS 安装脚本
│   ├── install.ps1                    # Windows 安装脚本
│   └── Makefile                       # 构建任务
│
├── test/                              # 测试
│   ├── integration/                   # 集成测试
│   │   └── translate_test.go
│   ├── e2e/                          # 端到端测试
│   │   └── cli_test.go
│   └── fixtures/                      # 测试数据
│       ├── en.json
│       ├── zh.json
│       └── terminology.json
│
├── docs/                              # 文档
│   ├── architecture.md                # 架构文档
│   ├── development.md                 # 开发指南
│   └── api.md                         # API 文档
│
├── .github/                           # GitHub 配置
│   └── workflows/
│       ├── test.yml                   # 测试 CI
│       ├── release.yml                # 发布 CI
│       └── lint.yml                   # 代码检查
│
├── .goreleaser.yml                    # GoReleaser 配置
├── go.mod                             # Go 模块
├── go.sum                             # Go 依赖锁定
├── Makefile                           # 构建脚本
├── README.md                          # 项目说明
├── LICENSE                            # MIT 许可证
└── CONTRIBUTING.md                    # 贡献指南
```

---

## 5. 开发阶段规划

### Phase 1: Core Foundation (4-5 周)

**目标**: 实现核心翻译功能、术语管理、增量翻译和 Key 过滤

**里程碑**:
- ✅ 项目初始化和基础架构
- ✅ CLI 框架（Cobra）
- ✅ 配置管理
- ✅ OpenAI Provider
- ✅ 术语检测和管理
- ✅ 基础批量翻译
- ✅ 格式保护
- ✅ 智能增量翻译（差异分析、自动合并）
- ✅ Key 过滤器（支持通配符模式）
- ✅ 单元测试 (>60%)

**可交付成果**:
```bash
# 基本翻译功能
jta en.json --to zh

# 术语管理
# 自动检测、保存、翻译术语

# 增量翻译（自动检测变更）
jta en.json --to zh  # 只翻译变更的 keys

# 强制完整翻译
jta en.json --to zh --force

# Key 过滤
jta en.json --to zh --keys "settings.*,user.*"
jta en.json --to zh --exclude-keys "admin.*,internal.*"
```

---

### Phase 2: Advanced Features (2-3 周)

**目标**: 多提供商支持和高级功能

**里程碑**:
- ✅ Anthropic Provider
- ✅ Google Provider
- ✅ Agentic 翻译（轻量反思）
- ✅ RTL 语言处理
- ✅ 并发优化
- ✅ 错误处理增强
- ✅ 集成测试

**可交付成果**:
```bash
# 多提供商
jta en.json --to zh --provider anthropic

# RTL 语言
jta en.json --to ar  # 自动处理
```

---

### Phase 3: Polish & Documentation (2 周)

**目标**: 完善功能、文档和测试

**里程碑**:
- ✅ 完整文档
- ✅ E2E 测试
- ✅ 性能优化
- ✅ 错误信息优化
- ✅ 日志系统
- ✅ 代码覆盖率 >80%

**可交付成果**:
- 完整的用户文档
- 开发者文档
- 示例项目

---

### Phase 4: Release Preparation (1-2 周)

**目标**: 生产就绪，发布 v1.0.0

**里程碑**:
- ✅ GoReleaser 配置
- ✅ 跨平台构建
- ✅ 安装脚本
- ✅ Homebrew 发布
- ✅ GitHub Release
- ✅ 性能基准测试

**可交付成果**:
- 二进制文件（Mac/Linux/Windows）
- Homebrew formula
- 安装脚本
- Release notes

---

## 6. 详细任务分解

### 6.1 Phase 1: Core Foundation (Week 1-4)

#### Week 1: Project Setup & CLI Framework

**Day 1-2: 项目初始化**
- [ ] 创建 Git 仓库
- [ ] 初始化 Go 模块 (`go mod init`)
- [ ] 设置项目结构
- [ ] 配置 golangci-lint
- [ ] 配置 pre-commit hooks
- [ ] 设置 GitHub Actions (test.yml)
- [ ] 编写 README.md

**Day 3-5: CLI 框架**
- [ ] 安装 Cobra: `go get github.com/spf13/cobra`
- [ ] 实现 `cmd/jta/main.go`
- [ ] 实现 `internal/cli/root.go` (根命令即翻译命令)
- [ ] 添加所有翻译相关的 flags
- [ ] 添加 `--version` flag (Cobra 自动处理)
- [ ] CLI 基础测试

**Day 6-7: 配置管理**
- [ ] 实现 `internal/config/config.go`
- [ ] 实现 `internal/config/loader.go`
- [ ] 环境变量支持
- [ ] 配置验证
- [ ] 测试配置加载

---

#### Week 2: Domain Models & OpenAI Provider

**Day 1-3: 领域模型**
- [ ] 实现 `internal/domain/translation.go`
- [ ] 实现 `internal/domain/terminology.go`
- [ ] 实现 `internal/domain/language.go`
- [ ] 定义常量和枚举
- [ ] 单元测试

**Day 4-7: OpenAI Provider**
- [ ] 实现 `internal/provider/provider.go` (接口)
- [ ] 实现 `internal/provider/openai/openai.go`
- [ ] 实现 `internal/provider/factory.go`
- [ ] API 调用封装
- [ ] 错误处理
- [ ] 重试逻辑 (`internal/util/retry.go`)
- [ ] Provider 测试（包括 mock 测试）

---

#### Week 3: Terminology Management

**Day 1-3: 术语检测器**
- [ ] 实现 `internal/terminology/detector.go`
- [ ] 使用 LLM 分析文本检测术语
- [ ] 实现 `internal/prompt/terminology.go`
- [ ] 术语分类逻辑（preserve vs consistent）
- [ ] 支持跳过检测但仍翻译缺失术语的逻辑
- [ ] 测试术语检测

**Day 4-5: 术语仓储**
- [ ] 实现 `internal/terminology/repository.go` (接口)
- [ ] 实现 `internal/terminology/json_repository.go`
- [ ] Load/Save 术语文件
- [ ] 测试仓储操作

**Day 6-7: 术语管理器**
- [ ] 实现 `internal/terminology/manager.go`
- [ ] 实现 `internal/terminology/translator.go`
- [ ] 集成检测器和仓储
- [ ] 术语翻译功能
- [ ] 构建 prompt 字典
- [ ] 完整测试

---

#### Week 4: Translation Engine & Format Protection

**Day 1-3: 批量翻译器**
- [ ] 实现 `internal/translator/batch.go`
- [ ] 实现批次创建逻辑
- [ ] 实现并发处理 (goroutines + errgroup)
- [ ] 实现 `internal/util/concurrent.go`
- [ ] 测试批量翻译

**Day 4-5: 格式保护**
- [ ] 实现 `internal/format/protector.go`
- [ ] 实现 `internal/format/extractor.go`
- [ ] 实现 `internal/format/validator.go`
- [ ] 实现 `internal/format/patterns.go`
- [ ] 各种格式的正则表达式
- [ ] 测试格式保护

**Day 6-7: 翻译引擎集成**
- [ ] 实现 `internal/translator/engine.go`
- [ ] 实现 `internal/translator/pipeline.go`
- [ ] 集成术语管理器
- [ ] 集成格式保护器
- [ ] 集成批量翻译器
- [ ] 实现 `internal/app/service.go`
- [ ] E2E 测试（基础翻译流程）

---

#### Week 4.5: Incremental Translation & Key Filtering

**Day 1-2: 差异分析器**
- [ ] 实现 `internal/diff/analyzer.go`
- [ ] 递归遍历 JSON 树对比差异
- [ ] 实现 `AnalyzeDiff` 方法
- [ ] 分类差异（新增、修改、删除、未变更）
- [ ] 实现 `internal/diff/detector.go`
- [ ] 文件时间戳检测
- [ ] 文本比较逻辑
- [ ] 测试差异分析

**Day 3-4: 增量翻译器**
- [ ] 实现 `internal/translator/incremental.go`
- [ ] 实现 `ShouldTranslate` 逻辑
- [ ] 实现 `internal/diff/merger.go`
- [ ] 合并翻译结果和未变更内容
- [ ] 删除多余的 keys
- [ ] 保持 JSON 结构完整性
- [ ] 测试增量翻译流程

**Day 5-6: Key 过滤器**
- [ ] 实现 `internal/keyfilter/pattern.go`
- [ ] 定义模式类型和结构
- [ ] 实现 `internal/keyfilter/parser.go`
- [ ] 解析模式字符串（逗号分隔）
- [ ] 识别模式类型（精确、单层、递归、通配符）
- [ ] 实现 `internal/keyfilter/matcher.go`
- [ ] 精确匹配算法
- [ ] 单层通配符匹配（`settings.*`）
- [ ] 递归通配符匹配（`settings.**`）
- [ ] 任意位置通配符匹配（`*.title`）
- [ ] 测试各种模式匹配

**Day 7: Key 过滤集成**
- [ ] 实现 `internal/keyfilter/filter.go`
- [ ] 递归遍历 JSON 树
- [ ] 构建完整 key path
- [ ] 应用包含和排除模式
- [ ] 优先级规则实现（exclude > include）
- [ ] 性能优化（缓存、提前终止）
- [ ] 测试过滤功能
- [ ] 集成到翻译引擎

---

### 6.2 Phase 2: Advanced Features (Week 5-7)

#### Week 5: Multiple Providers

**Day 1-3: Anthropic Provider**
- [ ] 实现 `internal/provider/anthropic/anthropic.go`
- [ ] HTTP 客户端封装
- [ ] API 调用和错误处理
- [ ] 测试

**Day 4-5: Google Provider**
- [ ] 实现 `internal/provider/google/google.go`
- [ ] 使用官方 SDK 或 HTTP 客户端
- [ ] 测试

**Day 6-7: 提供商切换测试**
- [ ] 更新 Factory
- [ ] CLI 选项支持
- [ ] 集成测试（各提供商）

---

#### Week 6: Agentic Translation & RTL

**Day 1-3: Agentic 翻译**
- [ ] 实现 `internal/translator/agentic.go`
- [ ] 实现 `internal/prompt/reflection.go`
- [ ] 轻量反思机制
- [ ] 选择性改进逻辑
- [ ] 测试

**Day 4-5: RTL 处理**
- [ ] 实现 `internal/rtl/processor.go`
- [ ] 实现 `internal/rtl/detector.go`
- [ ] 实现 `internal/rtl/formatter.go`
- [ ] 方向标记添加
- [ ] 标点符号转换
- [ ] 测试（阿拉伯语、希伯来语）

**Day 6-7: 并发优化**
- [ ] 优化 worker pool
- [ ] Rate limiting
- [ ] Context 超时处理
- [ ] 性能测试

---

#### Week 7: Error Handling & Validation

**Day 1-3: 增强错误处理**
- [ ] 定义错误类型
- [ ] 错误包装和传播
- [ ] 用户友好的错误消息
- [ ] 日志系统 (zerolog)

**Day 4-5: 验证器**
- [ ] 实现 `internal/validator/validator.go`
- [ ] 实现 `internal/validator/structure.go`
- [ ] 实现 `internal/validator/terminology.go`
- [ ] 实现 `internal/validator/format.go`
- [ ] 验证报告生成
- [ ] 测试

**Day 6-7: 集成测试**
- [ ] 编写 `test/integration/translate_test.go`
- [ ] 测试各种场景
- [ ] Mock AI Provider 测试

---

### 6.3 Phase 3: Polish & Documentation (Week 8-9)

#### Week 8: UI/UX & Documentation

**Day 1-3: 终端 UI 优化**
- [ ] 实现 `internal/cli/ui/progress.go` (bubbletea)
- [ ] 实现 `internal/cli/ui/prompt.go`
- [ ] 实现 `internal/cli/ui/output.go`
- [ ] 进度条和动画
- [ ] 彩色输出
- [ ] 测试

**Day 4-7: 文档编写**
- [ ] README.md 完善
- [ ] docs/architecture.md
- [ ] docs/development.md
- [ ] docs/api.md
- [ ] 代码注释完善
- [ ] GoDoc 文档

---

#### Week 9: Testing & Performance

**Day 1-3: 测试完善**
- [ ] 提高单元测试覆盖率 (>80%)
- [ ] 边界情况测试
- [ ] 错误路径测试
- [ ] Table-driven tests

**Day 4-5: E2E 测试**
- [ ] 编写 `test/e2e/cli_test.go`
- [ ] 测试完整 CLI 流程
- [ ] 测试各种选项组合

**Day 6-7: 性能优化**
- [ ] Profiling (pprof)
- [ ] 内存优化
- [ ] 并发优化
- [ ] Benchmark 测试

---

### 6.4 Phase 4: Release Preparation (Week 10-11)

#### Week 10: Build & Release

**Day 1-3: GoReleaser 配置**
- [ ] 配置 `.goreleaser.yml`
- [ ] 跨平台构建配置
- [ ] 归档和校验和
- [ ] 测试构建

**Day 4-5: 安装脚本**
- [ ] 编写 `scripts/install.sh`
- [ ] 编写 `scripts/install.ps1`
- [ ] 测试安装脚本（各平台）

**Day 6-7: GitHub Release**
- [ ] 配置 `.github/workflows/release.yml`
- [ ] 自动发布流程
- [ ] Release notes 模板
- [ ] 测试发布

---

#### Week 11: Distribution & Launch

**Day 1-2: Homebrew**
- [ ] 创建 Homebrew tap 仓库
- [ ] 编写 formula
- [ ] 测试 brew install

**Day 3-4: 文档站点（可选）**
- [ ] GitHub Pages 或 Read the Docs
- [ ] 文档整理

**Day 5-7: 发布准备**
- [ ] 最终测试（所有平台）
- [ ] 编写 CHANGELOG.md
- [ ] 准备发布公告
- [ ] v1.0.0 发布
- [ ] 社区宣传（Reddit, HN, Twitter）

---

## 7. 测试策略

### 7.1 测试层次

```
E2E Tests (端到端测试)          < 5%
  └─ CLI 完整流程测试

Integration Tests (集成测试)    15-20%
  ├─ Provider 集成测试
  ├─ 翻译流程集成测试
  └─ 文件 I/O 集成测试

Unit Tests (单元测试)           75-80%
  ├─ 术语管理
  ├─ 格式保护
  ├─ 验证器
  ├─ Prompt 生成
  └─ 工具函数
```

### 7.2 测试工具

- **testify**: 断言和 mock
- **mockery**: 自动生成 mock
- **httptest**: HTTP 测试
- **goleak**: Goroutine 泄漏检测

### 7.3 测试示例

**单元测试示例**:
```go
package terminology_test

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "github.com/hikanner/jta/internal/terminology"
)

func TestDetectTerms(t *testing.T) {
    // 创建 mock provider
    mockProvider := new(MockAIProvider)
    mockProvider.On("Complete", mock.Anything, mock.Anything).
        Return(&provider.CompletionResponse{
            Content: `{
                "preserveTerms": ["API", "OAuth"],
                "consistentTerms": ["credits", "premium"]
            }`,
        }, nil)
    
    // 创建检测器
    detector := terminology.NewDetector(mockProvider)
    
    // 执行检测
    terms, err := detector.DetectTerms(context.Background(), 
        []string{"You have 10 credits for API access"}, "en")
    
    // 断言
    assert.NoError(t, err)
    assert.Len(t, terms, 4)
    
    // 验证 preserve terms
    preserveTerms := filterByType(terms, terminology.TermTypePreserve)
    assert.ElementsMatch(t, []string{"API", "OAuth"}, getTermNames(preserveTerms))
    
    // 验证 consistent terms
    consistentTerms := filterByType(terms, terminology.TermTypeConsistent)
    assert.ElementsMatch(t, []string{"credits", "premium"}, getTermNames(consistentTerms))
    
    // 验证 mock 调用
    mockProvider.AssertExpectations(t)
}
```

**集成测试示例**:
```go
package integration_test

import (
    "context"
    "testing"
    "github.com/stretchr/testify/require"
    "github.com/hikanner/jta/internal/app"
    "github.com/hikanner/jta/internal/config"
)

func TestFullTranslationFlow(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }
    
    // 创建配置
    cfg := &config.Config{
        Provider: "openai",
        Model:    "gpt-4o",
        APIKey:   getTestAPIKey(t),
    }
    
    // 创建应用
    app, err := app.NewApp(cfg)
    require.NoError(t, err)
    
    // 准备输入
    sourceData := map[string]interface{}{
        "app": map[string]interface{}{
            "name": "MyApp",
            "description": "A powerful tool",
        },
        "messages": map[string]interface{}{
            "welcome": "Welcome, {username}!",
        },
    }
    
    input := translator.TranslationInput{
        Source:     sourceData,
        SourceLang: "en",
        TargetLang: "zh",
        Options: translator.TranslationOptions{
            BatchSize:   20,
            Concurrency: 3,
        },
    }
    
    // 执行翻译
    result, err := app.TranslationService.Translate(context.Background(), input)
    require.NoError(t, err)
    
    // 验证结果
    require.NotNil(t, result.Target)
    require.Equal(t, result.Stats.TotalItems, result.Stats.SuccessItems)
    
    // 验证结构
    appData, ok := result.Target["app"].(map[string]interface{})
    require.True(t, ok)
    require.Contains(t, appData, "name")
    require.Contains(t, appData, "description")
    
    // 验证格式保护（占位符）
    messages, ok := result.Target["messages"].(map[string]interface{})
    require.True(t, ok)
    welcome := messages["welcome"].(string)
    require.Contains(t, welcome, "{username}")  // 占位符保留
}
```

**E2E 测试示例**:
```go
package e2e_test

import (
    "os"
    "os/exec"
    "path/filepath"
    "testing"
    "github.com/stretchr/testify/require"
)

func TestCLITranslate(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping e2e test")
    }
    
    // 创建临时目录
    tmpDir := t.TempDir()
    
    // 准备测试文件
    sourceFile := filepath.Join(tmpDir, "en.json")
    err := os.WriteFile(sourceFile, []byte(`{
        "app": {
            "name": "MyApp",
            "description": "A powerful tool"
        }
    }`), 0644)
    require.NoError(t, err)
    
    // 执行 CLI 命令（无子命令）
    cmd := exec.Command("jta", sourceFile, 
        "--to", "zh",
        "--output", tmpDir,
        "-y")
    cmd.Env = append(os.Environ(), "OPENAI_API_KEY="+getTestAPIKey(t))
    
    output, err := cmd.CombinedOutput()
    require.NoError(t, err, "CLI output: %s", output)
    
    // 验证输出文件存在
    targetFile := filepath.Join(tmpDir, "zh.json")
    require.FileExists(t, targetFile)
    
    // 验证内容
    data, err := os.ReadFile(targetFile)
    require.NoError(t, err)
    require.Contains(t, string(data), "MyApp")  // 保留术语
    require.NotContains(t, string(data), "A powerful tool")  // 已翻译
}
```

### 7.4 测试覆盖率目标

| 包 | 目标覆盖率 |
|----|-----------|
| `internal/terminology` | > 90% |
| `internal/format` | > 90% |
| `internal/translator` | > 85% |
| `internal/validator` | > 85% |
| `internal/provider` | > 70% |
| `internal/cli` | > 60% |
| **总体** | **> 80%** |

---

## 8. 构建与发布

### 8.1 Makefile

```makefile
# Makefile

.PHONY: help build test lint clean install

# 默认目标
help:
	@echo "Available targets:"
	@echo "  build       - Build the binary"
	@echo "  test        - Run tests"
	@echo "  test-cover  - Run tests with coverage"
	@echo "  lint        - Run linters"
	@echo "  clean       - Clean build artifacts"
	@echo "  install     - Install binary to GOPATH/bin"
	@echo "  release     - Build release binaries for all platforms"

# 构建
build:
	go build -o bin/jta ./cmd/jta

# 测试
test:
	go test -v -race ./...

# 测试覆盖率
test-cover:
	go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html

# Lint
lint:
	golangci-lint run ./...

# 清理
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

# 安装
install:
	go install ./cmd/jta

# 本地发布测试
release-snapshot:
	goreleaser release --snapshot --clean

# 发布
release:
	goreleaser release --clean
```

### 8.2 GoReleaser 配置

`.goreleaser.yml`:
```yaml
# .goreleaser.yml
version: 2

before:
  hooks:
    - go mod tidy
    - go generate ./...

builds:
  - id: jta
    main: ./cmd/jta
    binary: jta
    env:
      - CGO_ENABLED=0
    goos:
      - linux
      - darwin
      - windows
    goarch:
      - amd64
      - arm64
    ldflags:
      - -s -w
      - -X github.com/hikanner/jta/internal/version.Version={{.Version}}
      - -X github.com/hikanner/jta/internal/version.Commit={{.Commit}}
      - -X github.com/hikanner/jta/internal/version.Date={{.Date}}

archives:
  - id: jta
    format: tar.gz
    format_overrides:
      - goos: windows
        format: zip
    name_template: "jta_{{ .Version }}_{{ .Os }}_{{ .Arch }}"
    files:
      - README.md
      - LICENSE
      - CHANGELOG.md

checksum:
  name_template: "checksums.txt"

snapshot:
  version_template: "{{ incpatch .Version }}-next"

changelog:
  sort: asc
  filters:
    exclude:
      - "^docs:"
      - "^test:"
      - "^chore:"

brews:
  - name: jta
    repository:
      owner: hikanner
      name: homebrew-jta
    homepage: https://github.com/hikanner/jta
    description: "AI-powered JSON translation with terminology management"
    license: MIT
    install: |
      bin.install "jta"
    test: |
      system "#{bin}/jta", "--version"

release:
  github:
    owner: hikanner
    name: jta
  draft: false
  prerelease: auto
```

### 8.3 安装脚本

**Linux/macOS** (`scripts/install.sh`):
```bash
#!/bin/sh
set -e

# Jta installer script

REPO="hikanner/jta"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

# Detect OS
OS="$(uname -s)"
case "$OS" in
    Darwin) OS="darwin" ;;
    Linux)  OS="linux" ;;
    *)      echo "Unsupported OS: $OS"; exit 1 ;;
esac

# Detect architecture
ARCH="$(uname -m)"
case "$ARCH" in
    x86_64)  ARCH="amd64" ;;
    arm64)   ARCH="arm64" ;;
    aarch64) ARCH="arm64" ;;
    *)       echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

# Get latest version
VERSION=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$VERSION" ]; then
    echo "Failed to get latest version"
    exit 1
fi

echo "Installing jta $VERSION..."

# Download URL
URL="https://github.com/$REPO/releases/download/$VERSION/jta_${VERSION#v}_${OS}_${ARCH}.tar.gz"

# Create temp directory
TMP_DIR=$(mktemp -d)
trap "rm -rf $TMP_DIR" EXIT

# Download and extract
echo "Downloading from $URL..."
curl -sL "$URL" | tar -xz -C "$TMP_DIR"

# Install
sudo mv "$TMP_DIR/jta" "$INSTALL_DIR/jta"
sudo chmod +x "$INSTALL_DIR/jta"

echo "✅ jta installed successfully to $INSTALL_DIR/jta"
echo ""
echo "Run 'jta --version' to verify installation"
```

**Windows** (`scripts/install.ps1`):
```powershell
# Jta installer script for Windows

$ErrorActionPreference = "Stop"

$Repo = "hikanner/jta"
$InstallDir = "$env:ProgramFiles\jta"

Write-Host "Installing jta..."

# Get latest version
$Release = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest"
$Version = $Release.tag_name

Write-Host "Version: $Version"

# Download URL
$Arch = "amd64"  # Windows only supports amd64 for now
$Asset = "jta_$($Version.TrimStart('v'))_windows_$Arch.zip"
$DownloadUrl = $Release.assets | Where-Object { $_.name -eq $Asset } | Select-Object -ExpandProperty browser_download_url

if (-not $DownloadUrl) {
    Write-Error "Failed to find download URL for $Asset"
    exit 1
}

# Create temp directory
$TempDir = New-Item -ItemType Directory -Path "$env:TEMP\jta-install-$(Get-Random)"

try {
    # Download
    $ZipPath = Join-Path $TempDir "jta.zip"
    Write-Host "Downloading from $DownloadUrl..."
    Invoke-WebRequest -Uri $DownloadUrl -OutFile $ZipPath

    # Extract
    Expand-Archive -Path $ZipPath -DestinationPath $TempDir -Force

    # Create install directory
    if (-not (Test-Path $InstallDir)) {
        New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
    }

    # Copy binary
    Copy-Item -Path (Join-Path $TempDir "jta.exe") -Destination $InstallDir -Force

    # Add to PATH
    $CurrentPath = [Environment]::GetEnvironmentVariable("Path", "Machine")
    if ($CurrentPath -notlike "*$InstallDir*") {
        [Environment]::SetEnvironmentVariable(
            "Path",
            "$CurrentPath;$InstallDir",
            "Machine"
        )
        Write-Host "Added $InstallDir to PATH"
    }

    Write-Host "✅ jta installed successfully to $InstallDir\jta.exe"
    Write-Host ""
    Write-Host "Run 'jta --version' to verify installation (restart terminal if needed)"

} finally {
    # Cleanup
    Remove-Item -Path $TempDir -Recurse -Force
}
```

### 8.4 GitHub Actions

**测试 CI** (`.github/workflows/test.yml`):
```yaml
name: Test

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    strategy:
      matrix:
        os: [ubuntu-latest, macos-latest, windows-latest]
        go: ['1.25']
    
    runs-on: ${{ matrix.os }}
    
    steps:
    - uses: actions/checkout@v4
    
    - name: Set up Go
      uses: actions/setup-go@v5
      with:
        go-version: ${{ matrix.go }}
    
    - name: Cache Go modules
      uses: actions/cache@v4
      with:
        path: ~/go/pkg/mod
        key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
        restore-keys: |
          ${{ runner.os }}-go-
    
    - name: Install dependencies
      run: go mod download
    
    - name: Run tests
      run: go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
    
    - name: Upload coverage
      uses: codecov/codecov-action@v4
      with:
        file: ./coverage.out
        flags: unittests
        name: codecov-${{ matrix.os }}

  lint:
    runs-on: ubuntu-latest
    
    steps:
    - uses: actions/checkout@v4
    
    - name: Set up Go
      uses: actions/setup-go@v5
      with:
        go-version: '1.25'
    
    - name: Run golangci-lint
      uses: golangci/golangci-lint-action@v6
      with:
        version: latest
```

**发布 CI** (`.github/workflows/release.yml`):
```yaml
name: Release

on:
  push:
    tags:
      - 'v*'

permissions:
  contents: write

jobs:
  goreleaser:
    runs-on: ubuntu-latest
    
    steps:
    - uses: actions/checkout@v4
      with:
        fetch-depth: 0
    
    - name: Set up Go
      uses: actions/setup-go@v5
      with:
        go-version: '1.25'
    
    - name: Run GoReleaser
      uses: goreleaser/goreleaser-action@v6
      with:
        version: latest
        args: release --clean
      env:
        GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

---

## 9. 性能优化

### 9.1 优化策略

**1. 批量处理**
- 批次大小: 20-50 个文本
- 减少 API 调用次数

**2. 并发控制**
- Worker pool 模式
- Goroutines + errgroup
- 并发数: 3-5 个请求

**3. 连接复用**
- HTTP/2 keep-alive
- 连接池

**4. 内存优化**
- 流式处理大文件
- 避免不必要的复制
- 使用 sync.Pool 复用对象

**5. 缓存策略**
- 术语翻译缓存
- API 响应缓存（可选）

### 9.2 Benchmark 示例

```go
func BenchmarkBatchTranslation(b *testing.B) {
    // 准备数据
    items := make([]batch.BatchItem, 100)
    for i := 0; i < 100; i++ {
        items[i] = batch.BatchItem{
            Key:  fmt.Sprintf("key%d", i),
            Text: "Sample text for translation",
        }
    }
    
    // 创建处理器（使用 mock provider）
    processor := batch.NewBatchProcessor(mockProvider, 20, 3)
    
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        _, err := processor.ProcessBatches(context.Background(), items)
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

### 9.3 性能目标

| 指标 | 目标 |
|------|------|
| **500 文本翻译** | < 3 分钟 |
| **API 调用次数** | < 30 次 |
| **内存使用** | < 100 MB |
| **并发效率** | > 80% |
| **CPU 使用** | < 50% (平均) |

---

## 10. 总结

### 10.1 关键技术决策

1. **Golang 1.25+**: 性能、并发、单一二进制
2. **Interface-driven**: 可扩展、可测试、解耦
3. **Design Patterns**: Strategy, Factory, Repository, Decorator, Chain
4. **Cobra + Viper**: 成熟的 CLI 框架
5. **GoReleaser**: 自动化跨平台构建和发布

### 10.2 核心优势

- ✅ **单一二进制**: 无依赖，易部署
- ✅ **高性能**: Goroutines 并发，快速翻译
- ✅ **跨平台**: Mac/Linux/Windows 原生支持
- ✅ **可维护**: 清晰的架构和设计模式
- ✅ **可扩展**: Interface 驱动，易于添加新功能
- ✅ **易测试**: Mock 和依赖注入

### 10.3 发布里程碑

| 版本 | 时间 | 功能 |
|------|------|------|
| **v1.0.0** | Week 11 | 核心功能完整：翻译、术语管理、增量翻译、Key 过滤、格式保护、RTL 支持 |
| **v1.1.0** | +4 weeks | 翻译质量评分、自定义 prompt 模板 |
| **v2.0.0** | +8 weeks | 高级功能、插件系统 |

---

**文档版本**: v1.0  
**最后更新**: 2025-10-21  
**作者**: 开发团队
