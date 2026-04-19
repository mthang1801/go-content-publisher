package twitter

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/dghubble/oauth1"
)

type Client struct {
	baseURL     string
	bearerToken string
	httpClient  *http.Client
	viClient    *http.Client
	enClient    *http.Client
}

type User struct {
	ID       string
	Username string
	Name     string
}

type Tweet struct {
	ID        string
	Text      string
	CreatedAt *time.Time
}

type userLookupResponse struct {
	Data *struct {
		ID       string `json:"id"`
		Username string `json:"username"`
		Name     string `json:"name"`
	} `json:"data"`
}

type userTimelineResponse struct {
	Data []struct {
		ID        string `json:"id"`
		Text      string `json:"text"`
		CreatedAt string `json:"created_at"`
	} `json:"data"`
}

type createTweetRequest struct {
	Text string `json:"text"`
}

type createTweetResponse struct {
	Data *struct {
		ID   string `json:"id"`
		Text string `json:"text"`
	} `json:"data"`
}

func New(bearerToken string, viCreds OAuthCredentials, enCreds OAuthCredentials) *Client {
	return &Client{
		baseURL:     "https://api.x.com/2",
		bearerToken: strings.TrimSpace(bearerToken),
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
		viClient: newOAuthHTTPClient(viCreds),
		enClient: newOAuthHTTPClient(enCreds),
	}
}

type OAuthCredentials struct {
	APIKey       string
	APISecret    string
	AccessToken  string
	AccessSecret string
}

func (c OAuthCredentials) Valid() bool {
	return strings.TrimSpace(c.APIKey) != "" &&
		strings.TrimSpace(c.APISecret) != "" &&
		strings.TrimSpace(c.AccessToken) != "" &&
		strings.TrimSpace(c.AccessSecret) != ""
}

func newOAuthHTTPClient(creds OAuthCredentials) *http.Client {
	if !creds.Valid() {
		return nil
	}
	config := oauth1.NewConfig(strings.TrimSpace(creds.APIKey), strings.TrimSpace(creds.APISecret))
	token := oauth1.NewToken(strings.TrimSpace(creds.AccessToken), strings.TrimSpace(creds.AccessSecret))
	return config.Client(context.Background(), token)
}

func (c *Client) CanCrawl() bool {
	return strings.TrimSpace(c.bearerToken) != ""
}

func (c *Client) CanPublishVI() bool {
	return c.viClient != nil
}

func (c *Client) CanPublishEN() bool {
	return c.enClient != nil
}

func (c *Client) LookupUserByUsername(ctx context.Context, username string) (User, error) {
	if !c.CanCrawl() {
		return User{}, fmt.Errorf("twitter bearer token is not configured")
	}

	username = strings.TrimPrefix(strings.TrimSpace(username), "@")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/users/by/username/"+url.PathEscape(username), nil)
	if err != nil {
		return User{}, fmt.Errorf("build twitter user lookup request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.bearerToken)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return User{}, fmt.Errorf("call twitter user lookup endpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return User{}, fmt.Errorf("unexpected twitter user lookup status: %s", resp.Status)
	}

	var payload userLookupResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return User{}, fmt.Errorf("decode twitter user lookup response: %w", err)
	}
	if payload.Data == nil || strings.TrimSpace(payload.Data.ID) == "" {
		return User{}, fmt.Errorf("twitter user lookup returned no user")
	}

	return User{
		ID:       payload.Data.ID,
		Username: payload.Data.Username,
		Name:     payload.Data.Name,
	}, nil
}

func (c *Client) GetUserTweets(ctx context.Context, userID string, sinceID string, maxResults int) ([]Tweet, error) {
	if !c.CanCrawl() {
		return nil, fmt.Errorf("twitter bearer token is not configured")
	}
	if maxResults <= 0 {
		maxResults = 10
	}
	if maxResults < 5 {
		maxResults = 5
	}
	if maxResults > 100 {
		maxResults = 100
	}

	query := url.Values{}
	query.Set("max_results", strconv.Itoa(maxResults))
	query.Set("tweet.fields", "created_at,text")
	query.Set("exclude", "retweets,replies")
	if strings.TrimSpace(sinceID) != "" {
		query.Set("since_id", strings.TrimSpace(sinceID))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/users/"+url.PathEscape(strings.TrimSpace(userID))+"/tweets?"+query.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("build twitter timeline request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.bearerToken)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call twitter timeline endpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected twitter timeline status: %s", resp.Status)
	}

	var payload userTimelineResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode twitter timeline response: %w", err)
	}

	tweets := make([]Tweet, 0, len(payload.Data))
	for _, item := range payload.Data {
		tweet := Tweet{
			ID:   item.ID,
			Text: item.Text,
		}
		if parsed, err := time.Parse(time.RFC3339, item.CreatedAt); err == nil {
			tweet.CreatedAt = &parsed
		}
		tweets = append(tweets, tweet)
	}
	return tweets, nil
}

func (c *Client) PublishTweetVI(ctx context.Context, text string) (string, error) {
	return c.publishTweet(ctx, c.viClient, text, "vi")
}

func (c *Client) PublishTweetEN(ctx context.Context, text string) (string, error) {
	return c.publishTweet(ctx, c.enClient, text, "en")
}

func (c *Client) publishTweet(ctx context.Context, client *http.Client, text string, account string) (string, error) {
	if client == nil {
		return "", fmt.Errorf("twitter %s publish credentials are not configured", account)
	}

	body, err := json.Marshal(createTweetRequest{Text: text})
	if err != nil {
		return "", fmt.Errorf("marshal twitter create tweet request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/tweets", strings.NewReader(string(body)))
	if err != nil {
		return "", fmt.Errorf("build twitter create tweet request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("call twitter create tweet endpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected twitter create tweet status: %s", resp.Status)
	}

	var payload createTweetResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", fmt.Errorf("decode twitter create tweet response: %w", err)
	}
	if payload.Data == nil || strings.TrimSpace(payload.Data.ID) == "" {
		return "", fmt.Errorf("twitter create tweet returned no tweet id")
	}
	return payload.Data.ID, nil
}
