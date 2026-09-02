package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

type OCRTicket struct {
	Number   string `json:"number"`
	Quantity int    `json:"quantity"`
}

type OCRResponse struct {
	Tickets  []OCRTicket `json:"tickets"`
	Warnings []string    `json:"warnings"`
}



type OcrService struct {
	client openai.Client
}

func NewOcrService(apiKey string) *OcrService {
	return &OcrService{
		client: openai.NewClient(option.WithAPIKey(apiKey)),
	}
}

// Hardcoded JSON Schema for Structured Outputs.
// Required to be Strict: true, which mandates additionalProperties: false.
var ocrResponseSchema = map[string]interface{}{
	"type": "object",
	"properties": map[string]interface{}{
		"tickets": map[string]interface{}{
			"type": "array",
			"items": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"number": map[string]interface{}{
						"type":        "string",
						"description": "The parsed lottery ticket number. Must be exactly 3 or 6 digits.",
					},
					"quantity": map[string]interface{}{
						"type":        "integer",
						"description": "The quantity/volume of the ticket. Defaults to 1 if not specified.",
					},
				},
				"required":             []string{"number", "quantity"},
				"additionalProperties": false,
			},
			"description": "The list of valid lottery tickets found in the image.",
		},
		"warnings": map[string]interface{}{
			"type": "array",
			"items": map[string]interface{}{
				"type": "string",
				"enum": []string{"image_blurry", "number_partially_hidden", "multiple_tickets_detected", "low_confidence", "not_lottery_ticket"},
			},
			"description": "Any warnings about image quality, partially visible numbers, excessive count of tickets (> 5), or if the content is not a Thai Government Lottery ticket.",
		},
	},
	"required":             []string{"tickets", "warnings"},
	"additionalProperties": false,
}

func (s *OcrService) PerformOCR(ctx context.Context, base64DataURL string) (*OCRResponse, []byte, error) {
	// System prompt explaining the lottery ticket parsing requirements
	systemPrompt := `You are an expert AI agent specializing in extracting ticket numbers and quantities from photos of Thai Government Lottery (สลากกินแบ่งรัฐบาล) tickets.

Your tasks:
1. Detect and transcribe all lottery ticket numbers.
   - Thai lottery tickets can be 6-digit (standard lottery) or 3-digit (N3 lottery).
   - Ensure you read every digit correctly. If you are uncertain about any digit, add "low_confidence" to the warnings list.
2. Determine the quantity/amount of each ticket. Look for multipliers like "x2", "2 ใบ", "จำนวน 2 ใบ" or similar indicators on the ticket. If not found, default quantity to 1.
3. Assess the photo quality and content:
   - If the image is blurry, add "image_blurry" to warnings.
   - If a ticket is partially covered, cropped, or digits are cut off, add "number_partially_hidden" to warnings.
   - If you detect more than 5 distinct ticket items, add "multiple_tickets_detected" to warnings.
   - If the image is not a Thai Government Lottery ticket at all, add "not_lottery_ticket" to warnings.

Only return data conforming strictly to the provided JSON Schema.`

	// ChatCompletion params using official SDK structure
	params := openai.ChatCompletionNewParams{
		Model: openai.ChatModelGPT4oMini,
		ResponseFormat: openai.ChatCompletionNewParamsResponseFormatUnion{
			OfJSONSchema: &openai.ResponseFormatJSONSchemaParam{
				JSONSchema: openai.ResponseFormatJSONSchemaJSONSchemaParam{
					Name:        "ocr_response",
					Description: openai.String("Schema for parsed lottery ticket data"),
					Strict:      openai.Bool(true),
					Schema:      ocrResponseSchema,
				},
			},
		},
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(systemPrompt),
			openai.UserMessage([]openai.ChatCompletionContentPartUnionParam{
				{
					OfImageURL: &openai.ChatCompletionContentPartImageParam{
						ImageURL: openai.ChatCompletionContentPartImageImageURLParam{
							URL: base64DataURL,
						},
					},
				},
			}),
		},
	}

	resp, err := s.client.Chat.Completions.New(ctx, params)
	if err != nil {
		return nil, nil, fmt.Errorf("openai completions API error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, nil, fmt.Errorf("openai returned empty completions choices")
	}

	contentJSON := resp.Choices[0].Message.Content
	rawResponseBytes := []byte(contentJSON)

	var ocrResp OCRResponse
	if err := json.Unmarshal(rawResponseBytes, &ocrResp); err != nil {
		return nil, rawResponseBytes, fmt.Errorf("failed to unmarshal openai ocr response: %w (content: %s)", err, contentJSON)
	}

	return &ocrResp, rawResponseBytes, nil
}
