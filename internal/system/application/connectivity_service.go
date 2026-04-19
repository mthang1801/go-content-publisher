package application

import (
	"context"
	"database/sql"
	"fmt"
)

type Status string

const (
	StatusOK      Status = "ok"
	StatusFailed  Status = "failed"
	StatusSkipped Status = "skipped"
)

type DatabaseCheck struct {
	Status  Status `json:"status"`
	Message string `json:"message"`
}

type TelegramCheck struct {
	Status    Status `json:"status"`
	Message   string `json:"message"`
	BotID     int64  `json:"bot_id,omitempty"`
	Username  string `json:"username,omitempty"`
	FirstName string `json:"first_name,omitempty"`
}

type DeepSeekCheck struct {
	Status       Status   `json:"status"`
	Message      string   `json:"message"`
	ModelCount   int      `json:"model_count,omitempty"`
	SampleModels []string `json:"sample_models,omitempty"`
}

type GeminiCheck struct {
	Status       Status   `json:"status"`
	Message      string   `json:"message"`
	ModelCount   int      `json:"model_count,omitempty"`
	SampleModels []string `json:"sample_models,omitempty"`
}

type ConnectivityResults struct {
	Database DatabaseCheck `json:"database"`
	Telegram TelegramCheck `json:"telegram"`
	DeepSeek DeepSeekCheck `json:"deepseek"`
	Gemini   GeminiCheck   `json:"gemini"`
}

type ConnectivityService struct {
	db       *sql.DB
	telegram TelegramChecker
	deepseek DeepSeekChecker
	gemini   GeminiChecker
}

func NewConnectivityService(db *sql.DB, telegram TelegramChecker, deepseek DeepSeekChecker, gemini GeminiChecker) *ConnectivityService {
	return &ConnectivityService{
		db:       db,
		telegram: telegram,
		deepseek: deepseek,
		gemini:   gemini,
	}
}

func (s *ConnectivityService) CheckAll(ctx context.Context) (ConnectivityResults, error) {
	results := ConnectivityResults{
		Database: s.CheckDatabase(ctx),
	}

	telegram, err := s.telegram.Check(ctx)
	if err != nil {
		results.Telegram = TelegramCheck{
			Status:  StatusFailed,
			Message: fmt.Sprintf("telegram check failed: %v", err),
		}
	} else {
		results.Telegram = telegram
	}

	deepseek, err := s.deepseek.Check(ctx)
	if err != nil {
		results.DeepSeek = DeepSeekCheck{
			Status:  StatusFailed,
			Message: fmt.Sprintf("deepseek check failed: %v", err),
		}
	} else {
		results.DeepSeek = deepseek
	}

	gemini, err := s.gemini.Check(ctx)
	if err != nil {
		results.Gemini = GeminiCheck{
			Status:  StatusFailed,
			Message: fmt.Sprintf("gemini check failed: %v", err),
		}
	} else {
		results.Gemini = gemini
	}

	return results, nil
}

func (s *ConnectivityService) CheckDatabase(ctx context.Context) DatabaseCheck {
	if s.db == nil {
		return DatabaseCheck{Status: StatusFailed, Message: "database handle is not initialized"}
	}
	if err := s.db.PingContext(ctx); err != nil {
		return DatabaseCheck{Status: StatusFailed, Message: fmt.Sprintf("database ping failed: %v", err)}
	}
	return DatabaseCheck{Status: StatusOK, Message: "database ping succeeded"}
}
