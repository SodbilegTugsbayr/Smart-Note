package eguneapi

import (
	"github.com/SodbilegTugsbayr/Smart-Note/backend/pkg/noteman"
	"github.com/SodbilegTugsbayr/Smart-Note/backend/pkg/quizman"
)

type GeneratedOutput struct {
	Note    noteman.Note   `json:"note"`
	Quizzes []quizman.Quiz `json:"quizzes"`
}

type rawQuiz struct {
	Question      string   `json:"question"`
	Options       []string `json:"options"`
	CorrectAnswer string   `json:"correct_answer"`
}

type rawOutput struct {
	Note    *noteman.Note `json:"note"`
	Quizzes []rawQuiz     `json:"quizzes"`
}
