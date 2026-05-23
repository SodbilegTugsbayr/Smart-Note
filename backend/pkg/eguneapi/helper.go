package eguneapi

import (
	"fmt"

	"github.com/openai/openai-go/v3"
)

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
							"description": "Exactly 4 answer option strings. One option must be the correct answer.",
							"minItems":    4,
							"maxItems":    4,
							"items": map[string]any{
								"type": "string",
							},
						},
						"correct_answer": map[string]any{
							"type":        "string",
							"description": "The correct answer copied exactly from one of the options. Do not add explanation or alternative wording.",
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

Тестийн хатуу дүрэм:
- Quiz бүрийн **options** массив яг 4 ширхэг хариултын хувилбартай байна
- Quiz бүрийн **correct_answer** нь **options** доторх нэг string-ийг үсэг, тэмдэгт, зай, нөхцөлөөр нь ЯГ ХУУЛСАН утга байна
- **correct_answer** дээр options-д байхгүй дэлгэрэнгүй тайлбар, өөр найруулга, нэмэлт өгүүлбэр бичиж болохгүй
- Хэрэв зөв хариулт дэлгэрэнгүй байх шаардлагатай бол тэр дэлгэрэнгүй өгүүлбэрийг options-ийн нэг хувилбар болгож оруулаад correct_answer-д яг тэр хувилбарыг хуулна
- JSON гаргахаасаа өмнө quiz бүр дээр "correct_answer ∈ options" нөхцөлийг өөрөө шалгана

Буруу жишээ:
{
  "question": "Шаардлагын хүчинтэй байдлыг шалгах гэдэг нь юуг илэрхийлдэг вэ?",
  "options": ["Хэрэглэгчийн хүсэл", "Системийн зорилго", "Баримт бичгийн бүтэц", "Хэрэглэгчийн бодит хэрэгцээ"],
  "correct_answer": "Хэрэглэгчийн бодит хэрэгцээг тусгасан эсэх"
}

Зөв жишээ:
{
  "question": "Шаардлагын хүчинтэй байдлыг шалгах гэдэг нь юуг илэрхийлдэг вэ?",
  "options": ["Хэрэглэгчийн хүсэл", "Системийн зорилго", "Баримт бичгийн бүтэц", "Хэрэглэгчийн бодит хэрэгцээг тусгасан эсэх"],
  "correct_answer": "Хэрэглэгчийн бодит хэрэгцээг тусгасан эсэх"
}

Заавар:
- Бүх гарчиг, тайлбар, асуулт, хариултыг МОНГОЛ хэлээр бичнэ үү
- Зөвхөн JSON хариу өгнө үү — нэмэлт текст, тайлбар оруулахгүй`, rawContent)
}

func prepareQuestionPrompt(courseContext, question string) string {
	return fmt.Sprintf(`Та Smart Note системийн сургалтын туслах чатбот байна.
Доорх хичээлийн тэмдэглэл дээр тулгуурлан хэрэглэгчийн асуултад Монгол хэлээр, товч бөгөөд ойлгомжтой хариулна уу.
Хариулт зөвхөн өгөгдсөн тэмдэглэлд тулгуурлах ёстой. Хэрэв тэмдэглэлд хангалттай мэдээлэл байхгүй бол түүнийгээ шууд хэлнэ үү.

--- ХИЧЭЭЛИЙН ТЭМДЭГЛЭЛ ---
%s
--- ТЭМДЭГЛЭЛ ТӨГСӨВ ---

Асуулт: %s`, courseContext, question)
}
