package deepseek

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
	Data []struct {
		ID string `json:"id"`
	} `json:"data"`
}

type chatCompletionsRequest struct {
	Model          string          `json:"model"`
	Messages       []chatMessage   `json:"messages"`
	Stream         bool            `json:"stream"`
	ResponseFormat *responseFormat `json:"response_format,omitempty"`
}

type responseFormat struct {
	Type string `json:"type"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatCompletionsResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func New(apiKey, model string) *Client {
	return &Client{
		baseURL: "https://api.deepseek.com",
		apiKey:  strings.TrimSpace(apiKey),
		model:   strings.TrimSpace(model),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) Check(ctx context.Context) (systemapp.DeepSeekCheck, error) {
	if c.apiKey == "" {
		return systemapp.DeepSeekCheck{
			Status:  systemapp.StatusSkipped,
			Message: "deepseek api key is not configured",
		}, nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/models", nil)
	if err != nil {
		return systemapp.DeepSeekCheck{}, fmt.Errorf("build deepseek models request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return systemapp.DeepSeekCheck{}, fmt.Errorf("call deepseek models endpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return systemapp.DeepSeekCheck{}, fmt.Errorf("unexpected deepseek status: %s", resp.Status)
	}

	var payload listModelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return systemapp.DeepSeekCheck{}, fmt.Errorf("decode deepseek response: %w", err)
	}

	sample := []string{}
	for _, model := range payload.Data {
		sample = append(sample, model.ID)
		if len(sample) == 3 {
			break
		}
	}

	return systemapp.DeepSeekCheck{
		Status:       systemapp.StatusOK,
		Message:      "deepseek api authentication succeeded",
		ModelCount:   len(payload.Data),
		SampleModels: sample,
	}, nil
}

func (c *Client) RewriteText(ctx context.Context, originalText string) (contentapp.RewriteResult, error) {
	if c.apiKey == "" {
		return contentapp.RewriteResult{}, fmt.Errorf("deepseek api key is not configured")
	}

	body, err := json.Marshal(chatCompletionsRequest{
		Model: defaultModel(c.model),
		Messages: []chatMessage{
			{
				Role:    "system",
				Content: "Bạn là biên tập viên tin tức tài chính - kinh tế - chính trị quốc tế. Chỉ trả về JSON hợp lệ đúng schema được yêu cầu.",
			},
			{
				Role:    "user",
				Content: structuredRewritePrompt(originalText),
			},
		},
		Stream:         false,
		ResponseFormat: &responseFormat{Type: "json_object"},
	})
	if err != nil {
		return contentapp.RewriteResult{}, fmt.Errorf("marshal deepseek chat completion request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", strings.NewReader(string(body)))
	if err != nil {
		return contentapp.RewriteResult{}, fmt.Errorf("build deepseek chat completion request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return contentapp.RewriteResult{}, fmt.Errorf("call deepseek chat completions endpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return contentapp.RewriteResult{}, fmt.Errorf("unexpected deepseek status: %s", resp.Status)
	}

	var payload chatCompletionsResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return contentapp.RewriteResult{}, fmt.Errorf("decode deepseek chat completion response: %w", err)
	}
	if len(payload.Choices) == 0 {
		return contentapp.RewriteResult{}, fmt.Errorf("deepseek returned no choices")
	}

	text := strings.TrimSpace(payload.Choices[0].Message.Content)
	if text == "" {
		return contentapp.RewriteResult{}, fmt.Errorf("deepseek returned empty rewritten content")
	}
	result, err := structuredrewrite.Parse(text)
	if err != nil {
		return contentapp.RewriteResult{}, fmt.Errorf("parse deepseek structured rewrite: %w", err)
	}
	return result, nil
}

func defaultModel(model string) string {
	if strings.TrimSpace(model) == "" {
		return "deepseek-chat"
	}
	return model
}

func structuredRewritePrompt(originalText string) string {
	return `Nhiệm vụ:
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
