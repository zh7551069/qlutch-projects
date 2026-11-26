package airuntime

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const (
	chatCompletionsPath      = "/compatible-mode/v1/chat/completions"
	multimodalGenerationPath = "/api/v1/services/aigc/multimodal-generation/generation"
	maxResponseBytes         = 1 << 20
)

var (
	errCall       = errors.New("provider call failed")
	errParse      = errors.New("provider response unusable")
	errShortAudio = errors.New("short audio limit")
	errMaterial   = errors.New("material bytes are not sent")
)

type modelPair struct {
	primary  string
	fallback string
}

// httpProvider is the single adapter installed for S01 and S02.
// It does not change a job, write a map, or record a price.
type httpProvider struct {
	baseURL string
	key     string
	ocr     modelPair
	asr     modelPair
	text    modelPair
	http    *http.Client
}

func (p *httpProvider) Complete(ctx context.Context, req ProviderRequest) (ProviderResponse, error) {
	if err := ctx.Err(); err != nil {
		return ProviderResponse{}, err
	}
	if p == nil {
		return ProviderResponse{}, errCall
	}
	switch {
	case req.SceneCode == SceneS02 || req.Subclass == SubclassPrefetch:
		return p.completeS02(ctx, req)
	case req.Subclass == SubclassASR:
		return p.completeASR(ctx, req)
	case req.Subclass == SubclassOCR:
		return p.completeOCR(ctx, req)
	case req.Subclass == SubclassText:
		return p.completeText(ctx, req, string(req.Fragment))
	default:
		return ProviderResponse{}, ErrSubclassRequired
	}
}

func (p *httpProvider) completeText(ctx context.Context, req ProviderRequest, user string) (ProviderResponse, error) {
	text, tokens, err := p.completeChat(ctx, p.text, mapMessages(req.TemplateVersion, user))
	if err != nil {
		return ProviderResponse{}, err
	}
	tree, ok := parseMap(text)
	if !ok {
		return ProviderResponse{}, errParse
	}
	return ProviderResponse{Tokens: tokens, Map: tree}, nil
}

func (p *httpProvider) completeOCR(ctx context.Context, req ProviderRequest) (ProviderResponse, error) {
	if len(req.Fragment) == 0 {
		return ProviderResponse{}, errCall
	}
	dataURL := "data:" + imageMIME(req.Fragment) + ";base64," + base64.StdEncoding.EncodeToString(req.Fragment)
	messages := []chatMessage{
		{Role: "system", Content: mapInstruction(req.TemplateVersion)},
		{Role: "user", Content: []imagePart{{
			Type:     "image_url",
			ImageURL: imageURL{URL: dataURL},
		}}},
	}
	text, tokens, err := p.completeChat(ctx, p.ocr, messages)
	if err != nil {
		return ProviderResponse{}, err
	}
	tree, ok := parseMap(text)
	if !ok {
		return ProviderResponse{}, errParse
	}
	return ProviderResponse{Tokens: tokens, Map: tree}, nil
}

func (p *httpProvider) completeASR(ctx context.Context, req ProviderRequest) (ProviderResponse, error) {
	if len(req.Fragment) == 0 {
		return ProviderResponse{}, errCall
	}
	if exceedsShortAudioLimit(req.Fragment) {
		return ProviderResponse{}, errShortAudio
	}
	text, tokens, err := p.completeMultimodal(ctx, p.asr, req.Fragment)
	if err != nil {
		return ProviderResponse{}, err
	}
	if tree, ok := parseMap(text); ok {
		return ProviderResponse{Tokens: tokens, Map: tree}, nil
	}
	if strings.TrimSpace(text) == "" {
		return ProviderResponse{}, errParse
	}
	structured, extra, err := p.completeChat(ctx, p.text, mapMessages(req.TemplateVersion, text))
	if err != nil {
		return ProviderResponse{}, err
	}
	tree, ok := parseMap(structured)
	if !ok {
		return ProviderResponse{}, errParse
	}
	return ProviderResponse{Tokens: tokens + extra, Map: tree}, nil
}

func (p *httpProvider) completeS02(ctx context.Context, req ProviderRequest) (ProviderResponse, error) {
	if len(req.Fragment) != 0 {
		return ProviderResponse{}, errMaterial
	}
	user := "节点标题：" + req.NodeTitle + "\n题量：" + strconv.Itoa(req.QuestionLimit)
	text, tokens, err := p.completeChat(ctx, p.text, questionMessages(req.TemplateVersion, req.QuestionLimit, user))
	if err != nil {
		return ProviderResponse{}, err
	}
	questions, ok := parseQuestions(text)
	if !ok {
		return ProviderResponse{}, errParse
	}
	return ProviderResponse{Tokens: tokens, Questions: questions}, nil
}

func (p *httpProvider) completeChat(ctx context.Context, models modelPair, messages []chatMessage) (string, int, error) {
	status, raw, err := p.postModels(ctx, chatCompletionsPath, models, func(model string) any {
		return chatPayload{Model: model, Messages: messages}
	})
	if err != nil {
		return "", 0, err
	}
	if status < 200 || status >= 300 {
		return "", 0, errCall
	}
	return parseChat(raw)
}

func (p *httpProvider) completeMultimodal(ctx context.Context, models modelPair, fragment []byte) (string, int, error) {
	dataURL := "data:" + audioMIME(fragment) + ";base64," + base64.StdEncoding.EncodeToString(fragment)
	status, raw, err := p.postModels(ctx, multimodalGenerationPath, models, func(model string) any {
		return mmPayload{
			Model: model,
			Input: mmInput{Messages: []mmMessage{{
				Role:    "user",
				Content: []mmPart{{Audio: dataURL}},
			}}},
		}
	})
	if err != nil {
		return "", 0, err
	}
	if status < 200 || status >= 300 {
		return "", 0, errCall
	}
	return parseMultimodal(raw)
}

func (p *httpProvider) postModels(ctx context.Context, path string, models modelPair, payload func(string) any) (int, []byte, error) {
	status, raw, err := p.post(ctx, path, payload(models.primary))
	if err != nil || !responseUnavailable(status, raw) {
		return status, raw, err
	}
	if models.fallback == "" || models.fallback == models.primary {
		return status, raw, nil
	}
	return p.post(ctx, path, payload(models.fallback))
}

func (p *httpProvider) post(ctx context.Context, path string, payload any) (int, []byte, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return 0, nil, errCall
	}
	endpoint, err := p.endpoint(path)
	if err != nil {
		return 0, nil, errCall
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return 0, nil, errCall
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.key)
	client := p.http
	if client == nil {
		client = &http.Client{CheckRedirect: refuseRedirect}
	}
	resp, err := client.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return 0, nil, ctx.Err()
		}
		return 0, nil, errCall
	}
	defer func() { _ = resp.Body.Close() }()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes))
	if err != nil {
		return 0, nil, errCall
	}
	return resp.StatusCode, raw, nil
}

func (p *httpProvider) endpoint(path string) (string, error) {
	parsed, err := url.Parse(p.baseURL)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return "", errCall
	}
	parsed.User = nil
	parsed.RawQuery = ""
	parsed.Fragment = ""
	parsed.Path = strings.TrimRight(parsed.Path, "/") + path
	return parsed.String(), nil
}

func mapMessages(version, user string) []chatMessage {
	return []chatMessage{
		{Role: "system", Content: mapInstruction(version)},
		{Role: "user", Content: user},
	}
}

func questionMessages(version string, limit int, user string) []chatMessage {
	return []chatMessage{
		{Role: "system", Content: questionInstruction(version, limit)},
		{Role: "user", Content: user},
	}
}

func mapInstruction(version string) string {
	return "模板 " + version + "。只根据这一次给出的单个片段输出 JSON，不要输出其它文字。对象形状：{\"catalogs\":[{\"title\":\"\",\"chapters\":[{\"title\":\"\",\"points\":[{\"title\":\"\"}]}]}]}。标题保持简短。"
}

func questionInstruction(version string, limit int) string {
	return "模板 " + version + "。只根据节点标题输出 JSON {\"questions\":[{\"stem\":\"\"}]}。题干数量不超过 " + strconv.Itoa(limit) + "。不要输出材料正文。"
}

type chatPayload struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content any    `json:"content"`
}

type imagePart struct {
	Type     string   `json:"type"`
	ImageURL imageURL `json:"image_url"`
}

type imageURL struct {
	URL string `json:"url"`
}

type mmPayload struct {
	Model string  `json:"model"`
	Input mmInput `json:"input"`
}

type mmInput struct {
	Messages []mmMessage `json:"messages"`
}

type mmMessage struct {
	Role    string   `json:"role"`
	Content []mmPart `json:"content"`
}

type mmPart struct {
	Audio string `json:"audio,omitempty"`
}

func responseUnavailable(status int, raw []byte) bool {
	if status == http.StatusNotFound {
		return true
	}
	var probe struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Error   *struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(raw, &probe); err != nil {
		return false
	}
	code := probe.Code
	message := probe.Message
	if probe.Error != nil {
		if probe.Error.Code != "" {
			code = probe.Error.Code
		}
		if probe.Error.Message != "" {
			message = probe.Error.Message
		}
	}
	return modelUnavailable(code, message)
}

func modelUnavailable(code, message string) bool {
	switch strings.ToLower(strings.TrimSpace(code)) {
	case "model_not_found", "modelnotfound", "model.notfound", "invalid_model", "invalidmodel", "model_not_exist", "modelnotexist":
		return true
	}
	text := strings.ToLower(message)
	if strings.Contains(text, "model not found") || strings.Contains(text, "model does not exist") || strings.Contains(text, "model not exist") || strings.Contains(text, "no such model") || strings.Contains(text, "unsupported model") {
		return true
	}
	return strings.Contains(message, "模型不存在") || strings.Contains(message, "模型未开通") || strings.Contains(message, "模型不可用")
}
