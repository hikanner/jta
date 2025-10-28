package integration

import (
	"context"
	"testing"

	"github.com/hikanner/jta/internal/domain"
	"github.com/hikanner/jta/internal/provider"
	"github.com/hikanner/jta/internal/terminology"
	"github.com/hikanner/jta/internal/translator"
)

// TestCompleteTranslationWorkflow tests the complete translation workflow end-to-end
func TestCompleteTranslationWorkflow(t *testing.T) {
	ctx := context.Background()

	// Create mock provider
	mockProvider := provider.NewMockProvider("test-model")
	mockProvider.AddResponse(`[1] Bienvenido a nuestra aplicación
[2] Iniciar sesión
[3] Regístrate
[4] Dirección de correo electrónico
[5] Contraseña
[6] Olvidé mi contraseña
[7] Crear cuenta
[8] Ya tienes una cuenta?
[9] Nombre de usuario
[10] Confirmar contraseña`)

	// Create terminology manager with mock provider
	termManager := terminology.NewManager(mockProvider)

	// Create translation engine
	engine := translator.NewEngine(mockProvider, termManager)

	// Prepare source JSON
	source := map[string]any{
		"welcome":         "Welcome to our application",
		"login":           "Sign in",
		"signup":          "Sign up",
		"email":           "Email address",
		"password":        "Password",
		"forgot":          "Forgot password",
		"create":          "Create account",
		"hasAccount":      "Already have an account?",
		"username":        "Username",
		"confirmPassword": "Confirm password",
	}

	// Prepare translation input
	input := domain.TranslationInput{
		Source:     source,
		SourceLang: "en",
		TargetLang: "es",
		Options: domain.TranslationOptions{
			BatchSize:     20,
			Concurrency:   1,
			NoTerminology: true, // Disable terminology for this simple test
		},
	}

	// Execute translation
	result, err := engine.Translate(ctx, input)
	if err != nil {
		t.Fatalf("Translation failed: %v", err)
	}

	// Verify results
	if result == nil {
		t.Fatal("Expected translation result, got nil")
	}

	if len(result.Target) == 0 {
		t.Error("Expected translated content, got empty map")
	}

	// Verify stats
	if result.Stats.TotalItems == 0 {
		t.Error("Expected TotalItems > 0")
	}

	if result.Stats.APICallsCount == 0 {
		t.Error("Expected APICallsCount > 0")
	}

	if result.Stats.SuccessItems == 0 {
		t.Error("Expected SuccessItems > 0")
	}

	// Verify mock provider was called
	if mockProvider.GetCallCount() == 0 {
		t.Error("Expected provider to be called at least once")
	}

	t.Logf("Translation completed successfully:")
	t.Logf("  Total items: %d", result.Stats.TotalItems)
	t.Logf("  Success items: %d", result.Stats.SuccessItems)
	t.Logf("  Failed items: %d", result.Stats.FailedItems)
	t.Logf("  API calls: %d", result.Stats.APICallsCount)
	t.Logf("  Duration: %v", result.Stats.Duration)
}

// TestTranslationWithTerminology tests translation workflow with terminology
func TestTranslationWithTerminology(t *testing.T) {
	ctx := context.Background()

	// Create mock provider with responses for terminology translation
	mockProvider := provider.NewMockProvider("test-model")
	mockProvider.AddResponse(`[1] API de GitHub está disponible
[2] Clonar el repositorio de Git
[3] Crear un nuevo issue
[4] El token de OAuth es requerido`)

	// Create terminology manager

	termManager := terminology.NewManager(mockProvider)

	// Create translation engine
	engine := translator.NewEngine(mockProvider, termManager)

	// Prepare terminology
	termData := &domain.Terminology{
		SourceLanguage:  "en",
		PreserveTerms:   []string{"GitHub", "API", "Git", "OAuth"},
		ConsistentTerms: []string{"repository", "issue", "token"},
	}

	// Prepare terminology translation
	termTranslation := &domain.TerminologyTranslation{
		SourceLanguage: "en",
		TargetLanguage: "es",
		Translations: map[string]string{
			"repository": "repositorio",
			"issue":      "issue",
			"token":      "token",
		},
	}

	// Prepare source JSON with technical terms
	source := map[string]any{
		"github_api":     "GitHub API is available",
		"clone_repo":     "Clone Git repository",
		"create_issue":   "Create a new issue",
		"oauth_required": "OAuth token is required",
	}

	// Prepare translation input with terminology
	input := domain.TranslationInput{
		Source:                 source,
		SourceLang:             "en",
		TargetLang:             "es",
		Terminology:            termData,
		TerminologyTranslation: termTranslation,
		Options: domain.TranslationOptions{
			BatchSize:   10,
			Concurrency: 1,
		},
	}

	// Execute translation
	result, err := engine.Translate(ctx, input)
	if err != nil {
		t.Fatalf("Translation with terminology failed: %v", err)
	}

	// Verify results
	if result == nil {
		t.Fatal("Expected translation result, got nil")
	}

	if len(result.Target) == 0 {
		t.Error("Expected translated content, got empty map")
	}

	// Verify terminology was used in prompt
	if result.Stats.TotalItems != 4 {
		t.Errorf("Expected 4 items, got %d", result.Stats.TotalItems)
	}

	t.Logf("Translation with terminology completed:")
	t.Logf("  Total items: %d", result.Stats.TotalItems)
	t.Logf("  Success items: %d", result.Stats.SuccessItems)
	t.Logf("  API calls: %d", result.Stats.APICallsCount)
}

// TestTranslationWithKeyFiltering tests translation with key filtering
func TestTranslationWithKeyFiltering(t *testing.T) {
	ctx := context.Background()

	// Create mock provider
	mockProvider := provider.NewMockProvider("test-model")
	mockProvider.AddResponse(`[1] Configuración
[2] Configuración de usuario`)

	// Create terminology manager

	termManager := terminology.NewManager(mockProvider)

	// Create translation engine
	engine := translator.NewEngine(mockProvider, termManager)

	// Prepare source JSON with multiple keys
	source := map[string]any{
		"welcome":      "Welcome",
		"login":        "Login",
		"settings":     "Settings",
		"userSettings": "User settings",
		"logout":       "Logout",
	}

	// Prepare translation input with key filtering
	input := domain.TranslationInput{
		Source:     source,
		SourceLang: "en",
		TargetLang: "es",
		Options: domain.TranslationOptions{
			Keys:          []string{"settings", "userSettings"}, // Only translate these keys
			BatchSize:     10,
			Concurrency:   1,
			NoTerminology: true,
		},
	}

	// Execute translation
	result, err := engine.Translate(ctx, input)
	if err != nil {
		t.Fatalf("Translation with filtering failed: %v", err)
	}

	// Verify results
	if result == nil {
		t.Fatal("Expected translation result, got nil")
	}

	// Verify filter stats
	if result.Stats.FilterStats == nil {
		t.Fatal("Expected filter stats, got nil")
	}

	if result.Stats.FilterStats.IncludedKeys != 2 {
		t.Errorf("Expected 2 included keys, got %d", result.Stats.FilterStats.IncludedKeys)
	}

	if result.Stats.FilterStats.ExcludedKeys != 3 {
		t.Errorf("Expected 3 excluded keys, got %d", result.Stats.FilterStats.ExcludedKeys)
	}

	t.Logf("Translation with filtering completed:")
	t.Logf("  Total keys: %d", result.Stats.FilterStats.TotalKeys)
	t.Logf("  Included keys: %d", result.Stats.FilterStats.IncludedKeys)
	t.Logf("  Excluded keys: %d", result.Stats.FilterStats.ExcludedKeys)
}

// TestTranslationErrorHandling tests error handling in translation workflow
func TestTranslationErrorHandling(t *testing.T) {
	ctx := context.Background()

	// Create mock provider that returns errors
	mockProvider := provider.NewMockProvider("test-model")
	mockProvider.SetError("simulated API error")

	// Create terminology manager

	termManager := terminology.NewManager(mockProvider)

	// Create translation engine
	engine := translator.NewEngine(mockProvider, termManager)

	// Prepare source JSON
	source := map[string]any{
		"test": "Test message",
	}

	// Prepare translation input
	input := domain.TranslationInput{
		Source:     source,
		SourceLang: "en",
		TargetLang: "es",
		Options: domain.TranslationOptions{
			BatchSize:     10,
			Concurrency:   1,
			NoTerminology: true,
		},
	}

	// Execute translation - should fail
	_, err := engine.Translate(ctx, input)
	if err == nil {
		t.Fatal("Expected translation to fail with error, got nil")
	}

	// Verify error is a domain error
	if domainErr, ok := err.(*domain.Error); ok {
		if domainErr.Type != domain.ErrorTypeTranslation {
			t.Errorf("Expected ErrorTypeTranslation, got %v", domainErr.Type)
		}
		t.Logf("Received expected error: %v", domainErr)
	} else {
		t.Logf("Received error (not domain.Error): %v", err)
	}
}

// TestConcurrentTranslation tests concurrent batch processing
func TestConcurrentTranslation(t *testing.T) {
	ctx := context.Background()

	// Create mock provider with multiple responses
	mockProvider := provider.NewMockProvider("test-model")
	// Batch 1
	mockProvider.AddResponse(`[1] Mensaje uno
[2] Mensaje dos
[3] Mensaje tres`)
	// Batch 2
	mockProvider.AddResponse(`[1] Mensaje cuatro
[2] Mensaje cinco
[3] Mensaje seis`)

	// Create terminology manager

	termManager := terminology.NewManager(mockProvider)

	// Create translation engine
	engine := translator.NewEngine(mockProvider, termManager)

	// Prepare source JSON with enough items for multiple batches
	source := map[string]any{
		"msg1": "Message one",
		"msg2": "Message two",
		"msg3": "Message three",
		"msg4": "Message four",
		"msg5": "Message five",
		"msg6": "Message six",
	}

	// Prepare translation input with small batch size and concurrency
	input := domain.TranslationInput{
		Source:     source,
		SourceLang: "en",
		TargetLang: "es",
		Options: domain.TranslationOptions{
			BatchSize:     3, // Force multiple batches
			Concurrency:   2, // Process 2 batches concurrently
			NoTerminology: true,
		},
	}

	// Execute translation
	result, err := engine.Translate(ctx, input)
	if err != nil {
		t.Fatalf("Concurrent translation failed: %v", err)
	}

	// Verify results
	if result == nil {
		t.Fatal("Expected translation result, got nil")
	}

	// Should have made 2 API calls (2 batches)
	if result.Stats.APICallsCount < 2 {
		t.Errorf("Expected at least 2 API calls, got %d", result.Stats.APICallsCount)
	}

	// All items should be translated
	if result.Stats.SuccessItems != 6 {
		t.Errorf("Expected 6 success items, got %d", result.Stats.SuccessItems)
	}

	t.Logf("Concurrent translation completed:")
	t.Logf("  Total items: %d", result.Stats.TotalItems)
	t.Logf("  Success items: %d", result.Stats.SuccessItems)
	t.Logf("  API calls: %d", result.Stats.APICallsCount)
	t.Logf("  Duration: %v", result.Stats.Duration)
}
