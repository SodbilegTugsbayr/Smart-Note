package eguneapi

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/SodbilegTugsbayr/Smart-Note/backend/pkg/quizman"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

type EguneService struct {
	BaseURL string
	APIKey  string
	Model   string
}

func NewService(baseUrl, apiKey, model string) *EguneService {
	return &EguneService{
		BaseURL: baseUrl,
		APIKey:  apiKey,
		Model:   model,
	}
}

func (s *EguneService) GenerateNote(rawContent string) (*GeneratedOutput, error) {
	client := openai.NewClient(
		option.WithBaseURL(s.BaseURL),
		option.WithAPIKey(s.APIKey),
	)

	chatCompletion, err := client.Chat.Completions.New(context.Background(), openai.ChatCompletionNewParams{
		Model: s.Model,
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
	if raw.Note == nil {
		return nil, fmt.Errorf("note is missing from generated output")
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

func (s *EguneService) AnswerQuestion(courseContext, question string) (string, error) {
	client := openai.NewClient(
		option.WithBaseURL(s.BaseURL),
		option.WithAPIKey(s.APIKey),
	)

	chatCompletion, err := client.Chat.Completions.New(context.Background(), openai.ChatCompletionNewParams{
		Model: s.Model,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(prepareQuestionPrompt(courseContext, question)),
		},
	})
	if err != nil {
		return "", fmt.Errorf("chat completion error: %w", err)
	}

	if len(chatCompletion.Choices) == 0 {
		return "", fmt.Errorf("no choices returned from chat completion")
	}

	return chatCompletion.Choices[0].Message.Content, nil
}
