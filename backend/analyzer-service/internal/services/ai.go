package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type AIRequestMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type AIChatPayload struct {
	Model    string             `json:"model"`
	Messages []AIRequestMessage `json:"messages"`
}

type AIChatResponseChoice struct {
	Message struct {
		Content string `json:"content"`
	} `json:"message"`
}

type AIChatResponse struct {
	Choices []AIChatResponseChoice `json:"choices"`
}

// QuerySuggestion is a GORM/JSON mapping for recommended DBA adjustments
type QuerySuggestion struct {
	Type        string `json:"type"`        // 'index', 'rewrite', 'config', 'other'
	Sql         string `json:"sql"`         // executable SQL statement if applicable
	Description string `json:"description"` // reasoning
}

// QueryAnalysisResult matches the JSON response structure from the AI
type QueryAnalysisResult struct {
	Issue       string            `json:"issue"`
	Rationale   string            `json:"rationale"`
	Suggestions []QuerySuggestion `json:"suggestions"`
}

// AnalyzeQueryWithAI sends the query, plan, and rules to the configured AI provider
func AnalyzeQueryWithAI(endpoint string, apiKey string, model string, dbType string, query string, explainPlan string, rules string) (*QueryAnalysisResult, error) {
	systemPrompt := `You are an expert Database Administrator (DBA). Analyze the user's database query and its EXPLAIN plan.
Identify any performance bottlenecks (such as sequential scans, high cost, missing indices, or unoptimized joins).
Provide clear suggestions to fix the issues (e.g. creating indexes or rewriting the query).

Return your response in a strict, valid JSON format matching this schema:
{
  "issue": "Summary of the main performance bottleneck",
  "rationale": "Brief technical explanation of why the bottleneck exists using the explain plan details",
  "suggestions": [
    {
      "type": "index", // use 'index' for CREATE INDEX, 'rewrite' for query rewriting, 'config' for configuration changes, or 'other'
      "sql": "CREATE INDEX idx_name ON table_name(column_name);", // Provide the executable SQL command (leave empty if none)
      "description": "Explanation of what this suggestion does"
    }
  ]
}
Return ONLY valid JSON. Do not include markdown code block formatting (e.g. do not wrap in triple backticks) or additional conversational text.`

	userPrompt := fmt.Sprintf("Database Type: %s\n\nQuery:\n%s\n\nExplain Plan:\n%s\n\n", dbType, query, explainPlan)
	if rules != "" {
		userPrompt += fmt.Sprintf("Additional Optimization Rules / Instructions:\n%s\n", rules)
	}

	payload := AIChatPayload{
		Model: model,
		Messages: []AIRequestMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
	}

	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal AI request: %w", err)
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request to AI endpoint failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("AI endpoint returned status %d. Response: %s", resp.StatusCode, string(bodyBytes))
	}

	var chatResp AIChatResponse
	if err := json.Unmarshal(bodyBytes, &chatResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal AI response: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return nil, fmt.Errorf("AI returned no suggestions choices. Body: %s", string(bodyBytes))
	}

	content := chatResp.Choices[0].Message.Content

	// Clean JSON if the LLM output was wrapped in markdown code blocks
	cleanedContent := strings.TrimSpace(content)
	if strings.HasPrefix(cleanedContent, "```") {
		cleanedContent = strings.TrimPrefix(cleanedContent, "```json")
		cleanedContent = strings.TrimPrefix(cleanedContent, "```")
		cleanedContent = strings.TrimSuffix(cleanedContent, "```")
		cleanedContent = strings.TrimSpace(cleanedContent)
	}

	var analysisResult QueryAnalysisResult
	if err := json.Unmarshal([]byte(cleanedContent), &analysisResult); err != nil {
		return nil, fmt.Errorf("failed to parse AI JSON response: %w. Raw output was: %s", err, content)
	}

	return &analysisResult, nil
}
