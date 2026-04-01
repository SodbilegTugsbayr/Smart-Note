package eguneapi

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/SodbilegTugsbayr/Smart-Note/backend/pkg/noteman"
	"github.com/SodbilegTugsbayr/Smart-Note/backend/pkg/quizman"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

var (
	baseURL = "http://localhost:8080/api/v1"
	apiKey  = "secret"
	model   = "model"
)

func SetBaseURL(url string) { baseURL = url }
func SetAPIKey(key string)  { apiKey = key }
func SetModel(m string)     { model = m }

// GeneratedOutput holds the combined LLM response
type GeneratedOutput struct {
	Note    noteman.Note   `json:"note"`
	Quizzes []quizman.Quiz `json:"quizzes"`
}

func getSchema() openai.ChatCompletionNewParamsResponseFormatUnion {
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"note": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"title": map[string]any{
						"type":        "string",
						"description": "Title of the note",
					},
					"summary": map[string]any{
						"type":        "string",
						"description": "A concise summary of the content",
					},
					"key_concepts": map[string]any{
						"type":        "array",
						"description": "List of key concepts extracted from the content",
						"minItems":    3,
						"maxItems":    10,
						"items": map[string]any{
							"type": "object",
							"properties": map[string]any{
								"concept": map[string]any{
									"type":        "string",
									"description": "The key concept or term",
								},
								"definition": map[string]any{
									"type":        "string",
									"description": "Clear definition or explanation of the concept",
								},
							},
							"required":             []string{"concept", "definition"},
							"additionalProperties": false,
						},
					},
					"flash_cards": map[string]any{
						"type":        "array",
						"description": "Flashcards for studying the content",
						"minItems":    3,
						"maxItems":    10,
						"items": map[string]any{
							"type": "object",
							"properties": map[string]any{
								"question": map[string]any{
									"type":        "string",
									"description": "The flashcard question",
								},
								"answer": map[string]any{
									"type":        "string",
									"description": "The flashcard answer",
								},
							},
							"required":             []string{"question", "answer"},
							"additionalProperties": false,
						},
					},
				},
				"required":             []string{"title", "summary", "key_concepts", "flash_cards"},
				"additionalProperties": false,
			},
			"quizzes": map[string]any{
				"type":        "array",
				"description": "Multiple choice quiz questions based on the content",
				"minItems":    3,
				"maxItems":    10,
				"items": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"question": map[string]any{
							"type":        "string",
							"description": "The quiz question",
						},
						"options": map[string]any{
							"type":        "array",
							"description": "Exactly 4 answer options (A, B, C, D)",
							"minItems":    4,
							"maxItems":    4,
							"items": map[string]any{
								"type": "string",
							},
						},
						"correct_answer": map[string]any{
							"type":        "string",
							"description": "The correct answer, must exactly match one of the options",
						},
					},
					"required":             []string{"question", "options", "correct_answer"},
					"additionalProperties": false,
				},
			},
		},
		"required":             []string{"note", "quizzes"},
		"additionalProperties": false,
	}

	schemaParam := openai.ResponseFormatJSONSchemaJSONSchemaParam{
		Name:        "note_and_quizzes",
		Description: openai.String("Schema for generating a structured note with key concepts, flashcards, and quizzes"),
		Schema:      schema,
		Strict:      openai.Bool(true),
	}

	return openai.ChatCompletionNewParamsResponseFormatUnion{
		OfJSONSchema: &openai.ResponseFormatJSONSchemaParam{
			JSONSchema: schemaParam,
		},
	}
}

func preparePrompt(rawContent string) string {
	return fmt.Sprintf(`Та боловсролын контент шинжлэгч байна. Доорх материалыг уншаад дараах бүтэцтэй сургалтын тэмдэглэл үүсгэнэ үү.

--- МАТЕРИАЛ ---
%s
--- МАТЕРИАЛ ТӨГСӨВ ---

Даалгавар:
1. **note.title** — Материалын агуулгыг илэрхийлэх товч гарчиг бичнэ үү
2. **note.summary** — Материалын үндсэн санааг 3-5 өгүүлбэрт товчлон тайлбарлана уу
3. **note.key_concepts** — Материалаас 5-10 чухал ойлголт, нэр томьёог тодорхойлолтын хамт гаргана уу
4. **note.flash_cards** — Суралцагчид цээжлэхэд туслах 5-10 асуулт-хариулт карт үүсгэнэ үү
5. **quizzes** — Материалын ойлголтыг шалгах 5-10 олон сонголттой тест асуулт бэлтгэнэ үү; дөрвөн хувилбар (options) өгч, correct_answer нь options-ийн аль нэгтэй яг таарах ёстой

Заавар:
- Бүх гарчиг, тайлбар, асуулт, хариултыг МОНГОЛ хэлээр бичнэ үү
- Зөвхөн JSON хариу өгнө үү — нэмэлт текст, тайлбар оруулахгүй`, rawContent)
}

// rawQuiz is used only for unmarshaling the LLM response (includes correct_answer)
type rawQuiz struct {
	Question      string   `json:"question"`
	Options       []string `json:"options"`
	CorrectAnswer string   `json:"correct_answer"`
}

type rawOutput struct {
	Note    *noteman.Note `json:"note"`
	Quizzes []rawQuiz     `json:"quizzes"`
}

func GenerateNote(rawContent string) (*GeneratedOutput, error) {
	client := openai.NewClient(
		option.WithBaseURL(baseURL),
		option.WithAPIKey(apiKey),
	)

	chatCompletion, err := client.Chat.Completions.New(context.Background(), openai.ChatCompletionNewParams{
		Model: model,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(preparePrompt(rawContent)),
		},
		ResponseFormat: getSchema(),
	})
	if err != nil {
		return nil, fmt.Errorf("chat completion error: %w", err)
	}

	if len(chatCompletion.Choices) == 0 {
		return nil, fmt.Errorf("no choices returned from chat completion")
	}

	content := chatCompletion.Choices[0].Message.Content

	var raw rawOutput
	if err := json.Unmarshal([]byte(content), &raw); err != nil {
		return nil, fmt.Errorf("error unmarshaling content: %w", err)
	}

	raw.Note.RawContent = rawContent

	quizzes := make([]quizman.Quiz, 0, len(raw.Quizzes))
	for _, rq := range raw.Quizzes {
		quizzes = append(quizzes, quizman.Quiz{
			NoteID:        0,
			Question:      rq.Question,
			Options:       rq.Options,
			CorrectAnswer: rq.CorrectAnswer,
		})
	}

	return &GeneratedOutput{
		Note:    *raw.Note,
		Quizzes: quizzes,
	}, nil
}
