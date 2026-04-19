package telegrambot

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"

	systemapp "go-content-bot/internal/system/application"
)

type Client struct {
	baseURL    string
	botToken   string
	httpClient *http.Client
}

type Update struct {
	UpdateID    int64    `json:"update_id"`
	Message     *Message `json:"message,omitempty"`
	ChannelPost *Message `json:"channel_post,omitempty"`
}

type Message struct {
	MessageID       int64  `json:"message_id"`
	MessageThreadID *int64 `json:"message_thread_id,omitempty"`
	Text            string `json:"text"`
	Caption         string `json:"caption"`
	Chat            Chat   `json:"chat"`
	From            *User  `json:"from,omitempty"`
}

type Chat struct {
	ID       int64  `json:"id"`
	Type     string `json:"type"`
	Title    string `json:"title"`
	Username string `json:"username"`
}

type User struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	Username  string `json:"username"`
}

type getMeResponse struct {
	OK          bool   `json:"ok"`
	Description string `json:"description"`
	Result      struct {
		ID        int64  `json:"id"`
		FirstName string `json:"first_name"`
		Username  string `json:"username"`
	} `json:"result"`
}

type sendMessageResponse struct {
	OK          bool   `json:"ok"`
	Description string `json:"description"`
	Result      struct {
		MessageID int64 `json:"message_id"`
	} `json:"result"`
}

type getUpdatesResponse struct {
	OK          bool     `json:"ok"`
	Description string   `json:"description"`
	Result      []Update `json:"result"`
}

type getChatResponse struct {
	OK          bool   `json:"ok"`
	Description string `json:"description"`
	Result      Chat   `json:"result"`
}

func New(botToken string) *Client {
	return &Client{
		baseURL:  "https://api.telegram.org",
		botToken: strings.TrimSpace(botToken),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) Check(ctx context.Context) (systemapp.TelegramCheck, error) {
	if c.botToken == "" {
		return systemapp.TelegramCheck{
			Status:  systemapp.StatusSkipped,
			Message: "telegram bot token is not configured",
		}, nil
	}

	endpoint := fmt.Sprintf("%s/bot%s/getMe", c.baseURL, url.PathEscape(c.botToken))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return systemapp.TelegramCheck{}, fmt.Errorf("build getMe request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return systemapp.TelegramCheck{}, fmt.Errorf("call getMe: %w", err)
	}
	defer resp.Body.Close()

	var payload getMeResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return systemapp.TelegramCheck{}, fmt.Errorf("decode getMe response: %w", err)
	}

	if resp.StatusCode != http.StatusOK || !payload.OK {
		description := payload.Description
		if description == "" {
			description = resp.Status
		}
		return systemapp.TelegramCheck{}, errors.New(description)
	}

	return systemapp.TelegramCheck{
		Status:    systemapp.StatusOK,
		Message:   "telegram bot authentication succeeded",
		BotID:     payload.Result.ID,
		Username:  payload.Result.Username,
		FirstName: payload.Result.FirstName,
	}, nil
}

func (c *Client) GetUpdates(ctx context.Context, offset int64, limit int) ([]Update, error) {
	if c.botToken == "" {
		return nil, fmt.Errorf("telegram bot token is not configured")
	}

	endpoint := fmt.Sprintf("%s/bot%s/getUpdates", c.baseURL, url.PathEscape(c.botToken))
	values := url.Values{}
	if offset > 0 {
		values.Set("offset", fmt.Sprintf("%d", offset))
	}
	if limit > 0 {
		values.Set("limit", fmt.Sprintf("%d", limit))
	}
	values.Set("timeout", "0")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"?"+values.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("build getUpdates request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call getUpdates: %w", err)
	}
	defer resp.Body.Close()

	var payload getUpdatesResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode getUpdates response: %w", err)
	}

	if resp.StatusCode != http.StatusOK || !payload.OK {
		description := payload.Description
		if description == "" {
			description = resp.Status
		}
		return nil, errors.New(description)
	}

	return payload.Result, nil
}

func (c *Client) GetChat(ctx context.Context, chatID string) (Chat, error) {
	if c.botToken == "" {
		return Chat{}, fmt.Errorf("telegram bot token is not configured")
	}

	endpoint := fmt.Sprintf("%s/bot%s/getChat", c.baseURL, url.PathEscape(c.botToken))
	values := url.Values{}
	values.Set("chat_id", strings.TrimSpace(chatID))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"?"+values.Encode(), nil)
	if err != nil {
		return Chat{}, fmt.Errorf("build getChat request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return Chat{}, fmt.Errorf("call getChat: %w", err)
	}
	defer resp.Body.Close()

	var payload getChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return Chat{}, fmt.Errorf("decode getChat response: %w", err)
	}

	if resp.StatusCode != http.StatusOK || !payload.OK {
		description := payload.Description
		if description == "" {
			description = resp.Status
		}
		return Chat{}, errors.New(description)
	}

	return payload.Result, nil
}

func (c *Client) SendMessage(ctx context.Context, chatID string, threadID *int64, text string) (string, error) {
	if c.botToken == "" {
		return "", fmt.Errorf("telegram bot token is not configured")
	}

	endpoint := fmt.Sprintf("%s/bot%s/sendMessage", c.baseURL, url.PathEscape(c.botToken))
	values := url.Values{}
	values.Set("chat_id", strings.TrimSpace(chatID))
	values.Set("text", sanitizeTelegramText(text))
	values.Set("disable_web_page_preview", "true")
	if threadID != nil {
		values.Set("message_thread_id", fmt.Sprintf("%d", *threadID))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(values.Encode()))
	if err != nil {
		return "", fmt.Errorf("build sendMessage request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("call sendMessage: %w", err)
	}
	defer resp.Body.Close()

	var payload sendMessageResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", fmt.Errorf("decode sendMessage response: %w", err)
	}

	if resp.StatusCode != http.StatusOK || !payload.OK {
		description := payload.Description
		if description == "" {
			description = resp.Status
		}
		return "", errors.New(description)
	}

	return fmt.Sprintf("%d", payload.Result.MessageID), nil
}

func sanitizeTelegramText(text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}
	if utf8.ValidString(text) {
		return text
	}
	return strings.ToValidUTF8(text, "")
}
