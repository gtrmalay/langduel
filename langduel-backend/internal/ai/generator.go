package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	OpenRouterBaseURL = "https://openrouter.ai/api/v1"
	DefaultModel      = "google/gemini-2.0-flash-001"
	GenerationTimeout = 120 * time.Second
)

type Generator struct {
	apiKey string
	model  string
	client *http.Client
}

type PhraseResponse struct {
	Prompt  string   `json:"prompt"`
	Answers []string `json:"answers"`
}

func NewGenerator() *Generator {
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	model := os.Getenv("OPENROUTER_MODEL")
	if model == "" {
		model = DefaultModel
	}

	return &Generator{
		apiKey: apiKey,
		model:  model,
		client: &http.Client{
			Timeout: GenerationTimeout,
		},
	}
}

func (g *Generator) GeneratePhrases(ctx context.Context, topic, difficulty, langFrom, langTo string, count int) ([]PhraseResponse, error) {
	if g.apiKey == "" {
		return nil, fmt.Errorf("OPENROUTER_API_KEY not set")
	}

	messages := g.buildMessages(topic, difficulty, langFrom, langTo, count)
	ruToEn := langFrom == "ru" && langTo == "en"

	reqBody := map[string]interface{}{
		"model":       g.model,
		"messages":    messages,
		"temperature": 0.7,
		"max_tokens":  8000,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", OpenRouterBaseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+g.apiKey)
	req.Header.Set("HTTP-Referer", "https://langduel.game")
	req.Header.Set("X-Title", "LangDuel")

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	content := strings.TrimSpace(result.Choices[0].Message.Content)

	// Remove markdown code blocks if present
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimPrefix(content, "json")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	phrases, err := g.parsePhrases(content, ruToEn)
	if err != nil {
		// Log raw response to help debug parse failures
		preview := content
		if len(preview) > 500 {
			preview = preview[:500] + "..."
		}
		return nil, fmt.Errorf("failed to parse valid phrases (raw: %s): %w", preview, err)
	}
	return phrases, nil
}

func (g *Generator) buildMessages(topic, difficulty, langFrom, langTo string, count int) []map[string]string {
	topicName := topic
	if topic == "default" {
		topicName = "everyday"
	} else if topic == "slang" {
		topicName = "casual slang"
	}

	ruToEn := langFrom == "ru" && langTo == "en"

	var userPrompt string
	if ruToEn {
		userPrompt = fmt.Sprintf(`Generate a JSON vocabulary list for a Russian-to-English translation game.

TOPIC: "%s"
Generate exactly %d entries. Each entry: a Russian word as the prompt, English translations as answers.

RULES:
- "prompt": real Russian word/phrase (Cyrillic only, no Latin)
- "answers": ALL English translations a player might write — put the most common one FIRST, then synonyms, colloquial forms
- Minimum 5 answers per entry, aim for 7-8
- No duplicate prompts, no duplicate answers within one entry
- Every prompt must be a real Russian word related to the topic

❌ BAD — prompt is not a real word for the topic, answers are incomplete:
{"prompt":"кошелёк","answers":["wallet"]}

✅ GOOD — complete synonym coverage:
{"prompt":"кошелёк","answers":["wallet","purse","billfold","pocketbook","coin purse"]}
{"prompt":"обувь","answers":["shoes","footwear","boots","sneakers","slippers","sandals"]}
{"prompt":"зонт","answers":["umbrella","brolly","parasol","rain umbrella"]}

Output ONLY a valid JSON array, no markdown fences, no explanation:
[{"prompt":"...","answers":["...","...","...","...","..."]}]`, topicName, count)
	} else {
		userPrompt = fmt.Sprintf(`Generate a JSON vocabulary list for an English-to-Russian translation game.

TOPIC: "%s"
Generate exactly %d entries. Each entry: a real English word as the prompt, Russian translations as answers.

RULES:
- "prompt": real English word/phrase (Latin only, no Cyrillic). NEVER use transliterations of Russian words (e.g. "banya", "blini", "kvass" are only OK if the topic is specifically about Russian culture).
- "answers": ALL Russian translations a player might write — put the most common/obvious one FIRST, then synonyms, colloquial forms, diminutives
- Minimum 5 answers per entry, aim for 7-8
- No duplicate prompts, no duplicate answers within one entry
- Every prompt must be a real English dictionary word

❌ BAD — wrong language in prompt, missing obvious translations:
{"prompt":"кошмар","answers":["nightmare"]}
{"prompt":"nightmare","answers":["ужасный сон","страшный сон"]}

✅ GOOD — correct prompt language, first answer is the most obvious Russian word:
{"prompt":"nightmare","answers":["кошмар","страшный сон","ужас","кошмарный сон","жуть"]}
{"prompt":"wallet","answers":["кошелёк","кошелек","бумажник","портмоне","кошель"]}
{"prompt":"homework","answers":["домашняя работа","домашнее задание","домашка","задание на дом","уроки"]}
{"prompt":"umbrella","answers":["зонт","зонтик","зонтище","зонтик от дождя"]}

Output ONLY a valid JSON array, no markdown fences, no explanation:
[{"prompt":"...","answers":["...","...","...","...","..."]}]`, topicName, count)
	}

	return []map[string]string{
		{"role": "user", "content": userPrompt},
	}
}

func (g *Generator) parsePhrases(content string, ruToEn bool) ([]PhraseResponse, error) {
	if content == "" {
		return nil, fmt.Errorf("empty response")
	}

	containsCyrillic := func(s string) bool {
		for _, r := range s {
			if (r >= 'а' && r <= 'я') || (r >= 'А' && r <= 'Я') || r == 'ё' || r == 'Ё' {
				return true
			}
		}
		return false
	}

	containsLatin := func(s string) bool {
		for _, r := range s {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				return true
			}
		}
		return false
	}

	// promptOK: для en→ru промпт должен быть латиницей, для ru→en — кириллицей
	promptOK := func(s string) bool {
		if ruToEn {
			return containsCyrillic(s) && !containsLatin(s)
		}
		return containsLatin(s) && !containsCyrillic(s)
	}

	// answerOK: для en→ru ответы кириллицей, для ru→en — латиницей
	answerOK := func(s string) bool {
		if ruToEn {
			return containsLatin(s)
		}
		return containsCyrillic(s)
	}

	tryParse := func(data string) ([]PhraseResponse, bool) {
		var phrases []PhraseResponse
		if err := json.Unmarshal([]byte(data), &phrases); err != nil {
			return nil, false
		}

		validPhrases := make([]PhraseResponse, 0)
		seenPrompts := make(map[string]bool)
		for _, p := range phrases {
			if p.Prompt == "" {
				continue
			}
			if len(p.Answers) == 0 {
				continue
			}

			validAnswers := make([]string, 0)
			seenAnswers := make(map[string]bool)
			for _, a := range p.Answers {
				lower := strings.ToLower(strings.TrimSpace(a))
				if lower != "" && answerOK(lower) && !seenAnswers[lower] {
					seenAnswers[lower] = true
					validAnswers = append(validAnswers, lower)
				}
			}

			promptLower := strings.ToLower(strings.TrimSpace(p.Prompt))

			if len(validAnswers) >= 1 && promptOK(promptLower) && !seenPrompts[promptLower] {
				seenPrompts[promptLower] = true
				validPhrases = append(validPhrases, PhraseResponse{
					Prompt:  promptLower,
					Answers: validAnswers,
				})
			}
		}

		return validPhrases, len(validPhrases) > 0
	}

	if phrases, ok := tryParse(content); ok {
		return phrases, nil
	}

	jsonStr := extractJSONArray(content)
	if jsonStr == "" {
		return nil, fmt.Errorf("no JSON found in response")
	}

	if phrases, ok := tryParse(jsonStr); ok {
		return phrases, nil
	}

	return nil, fmt.Errorf("failed to parse valid phrases")
}

func extractJSONArray(content string) string {
	content = strings.TrimSpace(content)

	// Try to find array
	startIdx := -1
	for i := 0; i < len(content); i++ {
		if content[i] == '[' {
			startIdx = i
			break
		}
		// Skip markdown code blocks
		if i+4 < len(content) && content[i:i+4] == "```" {
			i += 3
		}
	}

	if startIdx == -1 {
		return ""
	}

	// Count brackets to find matching end
	depth := 0
	inString := false
	escape := false

	for i := startIdx; i < len(content); i++ {
		c := content[i]

		if escape {
			escape = false
			continue
		}

		if c == '\\' && inString {
			escape = true
			continue
		}

		if c == '"' {
			inString = !inString
			continue
		}

		if !inString {
			if c == '[' || c == '{' {
				depth++
			} else if c == ']' || c == '}' {
				depth--
				if depth == 0 {
					return content[startIdx : i+1]
				}
			}
		}
	}

	return content[startIdx:]
}
