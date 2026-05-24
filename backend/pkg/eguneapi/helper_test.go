package eguneapi

import (
	"strings"
	"testing"
)

func TestPreparePromptIncludesRawContentAndStrictQuizRules(t *testing.T) {
	prompt := preparePrompt("raw lesson content")

	for _, want := range []string{"raw lesson content", "correct_answer", "options"} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("preparePrompt() missing %q", want)
		}
	}
}

func TestPrepareQuestionPromptIncludesCourseContextAndQuestion(t *testing.T) {
	prompt := prepareQuestionPrompt("course context", "What is the main idea?")

	for _, want := range []string{"course context", "What is the main idea?"} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("prepareQuestionPrompt() missing %q", want)
		}
	}
}
