package analyzer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
	"scam-checker-bot/internal/models"
)

const systemInstruction = `You are an expert cybersecurity scam detection assistant.
A user received a message claiming to be from customer support.
Analyze the message or screenshot and identify social engineering tactics
(urgency, impersonation, fear, unusual requests for credentials/OTP/gift cards/crypto).
Return ONLY valid JSON matching the requested schema — no markdown fences, no extra text.`

// GeminiAnalyzer wraps the Google Generative AI (Gemini) client.
type GeminiAnalyzer struct {
	client *genai.Client
	model  *genai.GenerativeModel
}

// NewGeminiAnalyzer creates and configures a GeminiAnalyzer using an AI Studio API key.
func NewGeminiAnalyzer(ctx context.Context, apiKey string) (*GeminiAnalyzer, error) {
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("gemini client: %w", err)
	}

	model := client.GenerativeModel("gemini-1.5-flash")
	model.SystemInstruction = &genai.Content{
		Parts: []genai.Part{genai.Text(systemInstruction)},
	}
	model.ResponseMIMEType = "application/json"
	model.ResponseSchema = &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"risk_score":     {Type: genai.TypeInteger},
			"risk_level":     {Type: genai.TypeString, Enum: []string{"LOW", "SUSPICIOUS", "HIGH"}},
			"claimed_entity": {Type: genai.TypeString},
			"warning_signs": {
				Type:  genai.TypeArray,
				Items: &genai.Schema{Type: genai.TypeString},
			},
			"safe_verification_steps": {
				Type:  genai.TypeArray,
				Items: &genai.Schema{Type: genai.TypeString},
			},
		},
		Required: []string{"risk_score", "risk_level", "claimed_entity", "warning_signs", "safe_verification_steps"},
	}

	return &GeminiAnalyzer{client: client, model: model}, nil
}

// Close releases the underlying client resources.
func (a *GeminiAnalyzer) Close() {
	a.client.Close()
}

// parseReport extracts the first text part from a Gemini response and unmarshals it.
func parseReport(resp *genai.GenerateContentResponse) (*models.AnalysisReport, error) {
	for _, cand := range resp.Candidates {
		if cand.Content == nil {
			continue
		}
		for _, part := range cand.Content.Parts {
			if txt, ok := part.(genai.Text); ok {
				var r models.AnalysisReport
				if err := json.Unmarshal([]byte(txt), &r); err != nil {
					return nil, fmt.Errorf("json unmarshal: %w", err)
				}
				return &r, nil
			}
		}
	}
	return nil, fmt.Errorf("no usable response from Gemini")
}

// AnalyzeTextMessage sends a plain-text message to Gemini for scam analysis.
func (a *GeminiAnalyzer) AnalyzeTextMessage(ctx context.Context, text string) (*models.AnalysisReport, error) {
	prompt := genai.Text(fmt.Sprintf(
		"Analyze this customer-support message for scam indicators:\n\n%s", text,
	))
	resp, err := a.model.GenerateContent(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("gemini generate (text): %w", err)
	}
	return parseReport(resp)
}

// AnalyzeImageMessage sends a screenshot to Gemini for OCR + scam analysis.
func (a *GeminiAnalyzer) AnalyzeImageMessage(ctx context.Context, imageBytes []byte, mimeType string) (*models.AnalysisReport, error) {
	parts := []genai.Part{
		genai.ImageData(mimeType, imageBytes),
		genai.Text("Examine this screenshot. Extract the text and analyze whether this support message is a scam."),
	}
	resp, err := a.model.GenerateContent(ctx, parts...)
	if err != nil {
		return nil, fmt.Errorf("gemini generate (image): %w", err)
	}
	return parseReport(resp)
}
