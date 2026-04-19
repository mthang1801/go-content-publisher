package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	contentapp "go-content-bot/internal/content/application"
	systemapp "go-content-bot/internal/system/application"
	"go-content-bot/internal/system/infrastructure/clients/structuredrewrite"
)

type Client struct {
	baseURL    string
	apiKey     string
	model      string
	httpClient *http.Client
}

type listModelsResponse struct {
	Models []struct {
		Name string `json:"name"`
	} `json:"models"`
}

type generateContentRequest struct {
	Contents         []content        `json:"contents"`
	GenerationConfig generationConfig `json:"generationConfig,omitempty"`
}

type generationConfig struct {
	ResponseMimeType string `json:"responseMimeType,omitempty"`
}

type content struct {
	Role  string `json:"role,omitempty"`
	Parts []part `json:"parts"`
}

type part struct {
	Text string `json:"text"`
}

type generateContentResponse struct {
	Candidates []struct {
		Content struct {
			Parts []part `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

func New(apiKey, model string) *Client {
	return &Client{
		baseURL: "https://generativelanguage.googleapis.com/v1beta",
		apiKey:  strings.TrimSpace(apiKey),
		model:   strings.TrimSpace(model),
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (c *Client) Check(ctx context.Context) (systemapp.GeminiCheck, error) {
	if c.apiKey == "" {
		return systemapp.GeminiCheck{
			Status:  systemapp.StatusSkipped,
			Message: "gemini api key is not configured",
		}, nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/models", nil)
	if err != nil {
		return systemapp.GeminiCheck{}, fmt.Errorf("build gemini models request: %w", err)
	}
	req.Header.Set("x-goog-api-key", c.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return systemapp.GeminiCheck{}, fmt.Errorf("call gemini models endpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return systemapp.GeminiCheck{}, fmt.Errorf("unexpected gemini status: %s", resp.Status)
	}

	var payload listModelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return systemapp.GeminiCheck{}, fmt.Errorf("decode gemini response: %w", err)
	}

	sample := make([]string, 0, 3)
	for _, model := range payload.Models {
		sample = append(sample, model.Name)
		if len(sample) == 3 {
			break
		}
	}

	return systemapp.GeminiCheck{
		Status:       systemapp.StatusOK,
		Message:      "gemini api authentication succeeded",
		ModelCount:   len(payload.Models),
		SampleModels: sample,
	}, nil
}

func (c *Client) RewriteText(ctx context.Context, originalText string) (contentapp.RewriteResult, error) {
	if c.apiKey == "" {
		return contentapp.RewriteResult{}, fmt.Errorf("gemini api key is not configured")
	}

	body, err := json.Marshal(generateContentRequest{
		Contents: []content{
			{
				Role: "user",
				Parts: []part{
					{
						Text: structuredRewritePrompt(originalText),
					},
				},
			},
		},
		GenerationConfig: generationConfig{ResponseMimeType: "application/json"},
	})
	if err != nil {
		return contentapp.RewriteResult{}, fmt.Errorf("marshal gemini generate content request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/models/"+defaultModel(c.model)+":generateContent", strings.NewReader(string(body)))
	if err != nil {
		return contentapp.RewriteResult{}, fmt.Errorf("build gemini generate content request: %w", err)
	}
	req.Header.Set("x-goog-api-key", c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return contentapp.RewriteResult{}, fmt.Errorf("call gemini generate content endpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return contentapp.RewriteResult{}, fmt.Errorf("unexpected gemini status: %s", resp.Status)
	}

	var payload generateContentResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return contentapp.RewriteResult{}, fmt.Errorf("decode gemini generate content response: %w", err)
	}
	if len(payload.Candidates) == 0 || len(payload.Candidates[0].Content.Parts) == 0 {
		return contentapp.RewriteResult{}, fmt.Errorf("gemini returned no candidates")
	}

	text := strings.TrimSpace(payload.Candidates[0].Content.Parts[0].Text)
	if text == "" {
		return contentapp.RewriteResult{}, fmt.Errorf("gemini returned empty rewritten content")
	}
	result, err := structuredrewrite.Parse(text)
	if err != nil {
		return contentapp.RewriteResult{}, fmt.Errorf("parse gemini structured rewrite: %w", err)
	}
	return result, nil
}

func defaultModel(model string) string {
	if strings.TrimSpace(model) == "" {
		return "gemini-2.5-flash"
	}
	return model
}

func structuredRewritePrompt(originalText string) string {
	return `Bạn là một biên tập viên tin tức tài chính - kinh tế - chính trị quốc tế chuyên nghiệp.

Nhiệm vụ:
1. Viết lại nội dung gốc thành bài tin tức ngắn gọn, khách quan, chuyên nghiệp bằng tiếng Việt.
2. Tạo bản tiếng Anh tương đương.
3. Tạo tweet ngắn tiếng Việt và tiếng Anh.
4. Kiểm tra spam/quảng cáo/link referral/shill/airdrop. Nếu không nên đăng, đặt shouldPublish=false và nêu reason.

Quy tắc:
- Không copy nguyên văn nội dung gốc.
- Giữ đúng tên riêng, số liệu, thuật ngữ tài chính phổ biến.
- Loại bỏ URL, @username, tên nguồn và lời kêu gọi follow/join.
- Không phóng đại hoặc bịa thêm dữ kiện.
- Ưu tiên thông tin có tác động đến thị trường, kinh tế vĩ mô, địa chính trị, hàng hóa, chứng khoán, forex hoặc crypto.

Chỉ trả về JSON hợp lệ theo schema:
{
  "rewrittenText": "Bài viết tiếng Việt, để trống nếu shouldPublish=false",
  "rewrittenTextEn": "Bản tiếng Anh tương đương, để trống nếu shouldPublish=false",
  "tweetVI": "Tweet tiếng Việt tối đa 250 ký tự, không Markdown",
  "tweetEN": "Tweet tiếng Anh tối đa 250 ký tự, không Markdown",
  "factCheckNote": "Ghi chú kiểm chứng ngắn bằng tiếng Việt",
  "shouldPublish": true,
  "reason": "Lý do nếu shouldPublish=false"
}

Nội dung gốc:
` + originalText
}
