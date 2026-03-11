package translator

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/sashabaranov/go-openai"
)

// Translator handles translation using Doubao API
type Translator struct {
	client *openai.Client
	model  string
}

// TranslationResult holds the translation result for one text
type TranslationResult struct {
	Text   string            `json:"text"`
	ID     string            `json:"id"`
	Zh     string            `json:"zh"`
	En     string            `json:"en"`
	Id     string            `json:"id_lang"` // id is a keyword in Go
	Th     string            `json:"th"`
	Vi     string            `json:"vi"`
	Ms     string            `json:"ms"`
	Error  error             `json:"-"`
}

// New creates a new Translator
func New(apiKey, baseURL, model string) *Translator {
	config := openai.DefaultConfig(apiKey)
	config.BaseURL = baseURL
	client := openai.NewClientWithConfig(config)

	return &Translator{
		client: client,
		model:  model,
	}
}

// TranslateTexts translates multiple texts in parallel
func (t *Translator) TranslateTexts(ctx context.Context, texts []string, ids []string) ([]TranslationResult, error) {
	if len(texts) == 0 {
		return nil, nil
	}

	// Ensure ids slice has the same length as texts
	if ids == nil {
		ids = make([]string, len(texts))
	} else if len(ids) < len(texts) {
		// Extend ids slice to match texts length
		extendedIDs := make([]string, len(texts))
		copy(extendedIDs, ids)
		ids = extendedIDs
	}

	// Process in batches to avoid rate limits
	const batchSize = 10
	var results []TranslationResult
	var mu sync.Mutex
	var wg sync.WaitGroup
	errChan := make(chan error, len(texts))

	for i := 0; i < len(texts); i += batchSize {
		end := i + batchSize
		if end > len(texts) {
			end = len(texts)
		}

		batchTexts := texts[i:end]
		batchIDs := ids[i:end]

		wg.Add(1)
		go func(textsBatch, idsBatch []string) {
			defer wg.Done()

			batchResults, err := t.translateBatch(ctx, textsBatch, idsBatch)
			if err != nil {
				errChan <- fmt.Errorf("error translating batch: %w", err)
				return
			}

			mu.Lock()
			results = append(results, batchResults...)
			mu.Unlock()
		}(batchTexts, batchIDs)
	}

	wg.Wait()
	close(errChan)

	// Check for errors
	var errs []string
	for err := range errChan {
		errs = append(errs, err.Error())
	}
	if len(errs) > 0 {
		return results, fmt.Errorf("translation errors: %s", strings.Join(errs, "; "))
	}

	return results, nil
}

// translateBatch translates a batch of texts
func (t *Translator) translateBatch(ctx context.Context, texts []string, ids []string) ([]TranslationResult, error) {
	// Build prompt for batch translation
	prompt := buildBatchPrompt(texts)

	resp, err := t.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: t.model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "You are a professional translator. Translate the given Chinese texts to multiple languages. Return results in strict JSON format.",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		Temperature: 0.3,
	})

	if err != nil {
		return nil, fmt.Errorf("API error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from API")
	}

	// Parse the response
	content := resp.Choices[0].Message.Content

	// Try to extract JSON from markdown code blocks if present
	content = extractJSON(content)

	var translations []TranslationResult
	if err := json.Unmarshal([]byte(content), &translations); err != nil {
		// Try alternative format - array of objects
		var altFormat []map[string]string
		if err2 := json.Unmarshal([]byte(content), &altFormat); err2 == nil {
			for i, item := range altFormat {
				if i < len(texts) {
					translations = append(translations, TranslationResult{
						Text: texts[i],
						ID:   ids[i],
						Zh:   item["zh"],
						En:   item["en"],
						Id:   item["id"],
						Th:   item["th"],
						Vi:   item["vi"],
						Ms:   item["ms"],
					})
				}
			}
		} else {
			return nil, fmt.Errorf("failed to parse response: %w, content: %s", err, content)
		}
	}

	// Ensure IDs are set
	for i := range translations {
		if i < len(ids) {
			translations[i].ID = ids[i]
			translations[i].Text = texts[i]
		}
	}

	return translations, nil
}

// buildBatchPrompt builds the prompt for batch translation
func buildBatchPrompt(texts []string) string {
	var sb strings.Builder
	sb.WriteString("Translate the following Chinese texts to multiple languages. ")
	sb.WriteString("Return a JSON array where each object has the following structure:\n")
	sb.WriteString(`{"zh": "original chinese", "en": "english", "id": "indonesian", "th": "thai", "vi": "vietnamese", "ms": "malay"}`)
	sb.WriteString("\n\nTexts to translate:\n")

	for i, text := range texts {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, text))
	}

	sb.WriteString("\nRespond with ONLY the JSON array, no other text.")
	return sb.String()
}

// extractJSON extracts JSON from markdown code blocks
func extractJSON(content string) string {
	content = strings.TrimSpace(content)

	// Check for markdown code blocks
	if strings.HasPrefix(content, "```json") {
		content = strings.TrimPrefix(content, "```json")
		content = strings.TrimSuffix(content, "```")
		return strings.TrimSpace(content)
	}

	if strings.HasPrefix(content, "```") {
		content = strings.TrimPrefix(content, "```")
		content = strings.TrimSuffix(content, "```")
		return strings.TrimSpace(content)
	}

	return content
}

// TranslateSingle translates a single text
func (t *Translator) TranslateSingle(ctx context.Context, text string) (*TranslationResult, error) {
	results, err := t.TranslateTexts(ctx, []string{text}, []string{""})
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("no translation result")
	}
	return &results[0], nil
}
