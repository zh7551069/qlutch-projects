package airuntime

import (
	"bytes"
	"encoding/json"
	"strings"
)

type flexInt int

func (f *flexInt) UnmarshalJSON(raw []byte) error {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || string(raw) == "null" {
		*f = 0
		return nil
	}
	var whole int
	if err := json.Unmarshal(raw, &whole); err == nil {
		*f = flexInt(whole)
		return nil
	}
	var number float64
	if err := json.Unmarshal(raw, &number); err != nil {
		return err
	}
	*f = flexInt(number)
	return nil
}

func parseChat(raw []byte) (string, int, error) {
	var body struct {
		Choices []struct {
			Message struct {
				Content json.RawMessage `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     flexInt `json:"prompt_tokens"`
			CompletionTokens flexInt `json:"completion_tokens"`
			TotalTokens      flexInt `json:"total_tokens"`
			InputTokens      flexInt `json:"input_tokens"`
			OutputTokens     flexInt `json:"output_tokens"`
		} `json:"usage"`
		Error *struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(raw, &body); err != nil {
		return "", 0, errParse
	}
	if body.Error != nil && body.Error.Code != "" && len(body.Choices) == 0 {
		return "", 0, errCall
	}
	if len(body.Choices) == 0 {
		return "", 0, errParse
	}
	text := contentText(body.Choices[0].Message.Content)
	if strings.TrimSpace(text) == "" {
		return "", 0, errParse
	}
	tokens := pickUsage(body.Usage.TotalTokens, body.Usage.PromptTokens, body.Usage.CompletionTokens, body.Usage.InputTokens, body.Usage.OutputTokens)
	return text, tokens, nil
}

func parseMultimodal(raw []byte) (string, int, error) {
	var body struct {
		Code   string `json:"code"`
		Output struct {
			Text    string `json:"text"`
			Choices []struct {
				Message struct {
					Content json.RawMessage `json:"content"`
				} `json:"message"`
			} `json:"choices"`
		} `json:"output"`
		Usage struct {
			InputTokens  flexInt `json:"input_tokens"`
			OutputTokens flexInt `json:"output_tokens"`
			TotalTokens  flexInt `json:"total_tokens"`
		} `json:"usage"`
	}
	if err := json.Unmarshal(raw, &body); err != nil {
		return "", 0, errParse
	}
	var text string
	if len(body.Output.Choices) > 0 {
		text = contentText(body.Output.Choices[0].Message.Content)
	}
	if strings.TrimSpace(text) == "" {
		text = body.Output.Text
	}
	if body.Code != "" && strings.TrimSpace(text) == "" {
		return "", 0, errCall
	}
	if strings.TrimSpace(text) == "" {
		return "", 0, errParse
	}
	tokens := pickUsage(body.Usage.TotalTokens, 0, 0, body.Usage.InputTokens, body.Usage.OutputTokens)
	return text, tokens, nil
}

func contentText(raw json.RawMessage) string {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || string(raw) == "null" {
		return ""
	}
	switch raw[0] {
	case '"':
		var text string
		if err := json.Unmarshal(raw, &text); err != nil {
			return ""
		}
		return text
	case '[':
		var parts []struct {
			Text string `json:"text"`
		}
		if err := json.Unmarshal(raw, &parts); err != nil {
			return ""
		}
		var b strings.Builder
		for _, part := range parts {
			b.WriteString(part.Text)
		}
		return b.String()
	default:
		var one struct {
			Text string `json:"text"`
		}
		if err := json.Unmarshal(raw, &one); err != nil {
			return ""
		}
		return one.Text
	}
}

func pickUsage(total, prompt, completion, input, output flexInt) int {
	if total > 0 {
		return int(total)
	}
	if prompt > 0 || completion > 0 {
		return positive(prompt) + positive(completion)
	}
	return positive(input) + positive(output)
}

func positive(value flexInt) int {
	if value <= 0 {
		return 0
	}
	return int(value)
}

func parseMap(text string) (MapTree, bool) {
	raw := extractJSONObject(text)
	if raw == "" {
		return MapTree{}, false
	}
	var probe map[string]json.RawMessage
	if err := json.Unmarshal([]byte(raw), &probe); err != nil {
		return MapTree{}, false
	}
	catalogsRaw, ok := probe["catalogs"]
	if !ok {
		return MapTree{}, false
	}
	var catalogs []struct {
		Title    string `json:"title"`
		Chapters []struct {
			Title  string `json:"title"`
			Points []struct {
				Title string `json:"title"`
			} `json:"points"`
		} `json:"chapters"`
	}
	if err := json.Unmarshal(catalogsRaw, &catalogs); err != nil {
		return MapTree{}, false
	}
	tree := MapTree{Catalogs: make([]Catalog, 0, len(catalogs))}
	for _, catalog := range catalogs {
		item := Catalog{Title: catalog.Title, Chapters: make([]Chapter, 0, len(catalog.Chapters))}
		for _, chapter := range catalog.Chapters {
			next := Chapter{Title: chapter.Title, Points: make([]Point, 0, len(chapter.Points))}
			for _, point := range chapter.Points {
				next.Points = append(next.Points, Point{Title: point.Title})
			}
			item.Chapters = append(item.Chapters, next)
		}
		tree.Catalogs = append(tree.Catalogs, item)
	}
	return tree, true
}

func parseQuestions(text string) ([]Question, bool) {
	raw := extractJSONObject(text)
	if raw == "" {
		return nil, false
	}
	var probe map[string]json.RawMessage
	if err := json.Unmarshal([]byte(raw), &probe); err != nil {
		return nil, false
	}
	questionsRaw, ok := probe["questions"]
	if !ok {
		return nil, false
	}
	var questions []struct {
		Stem string `json:"stem"`
	}
	if err := json.Unmarshal(questionsRaw, &questions); err != nil {
		return nil, false
	}
	out := make([]Question, 0, len(questions))
	for _, question := range questions {
		out = append(out, Question{Stem: question.Stem})
	}
	return out, true
}

func extractJSONObject(text string) string {
	text = strings.TrimSpace(text)
	if strings.HasPrefix(text, "```") {
		text = strings.TrimPrefix(text, "```json")
		text = strings.TrimPrefix(text, "```")
		if end := strings.LastIndex(text, "```"); end >= 0 {
			text = text[:end]
		}
		text = strings.TrimSpace(text)
	}
	start := strings.Index(text, "{")
	end := strings.LastIndex(text, "}")
	if start < 0 || end <= start {
		return ""
	}
	return text[start : end+1]
}
