package application

import "context"

type TelegramChecker interface {
	Check(ctx context.Context) (TelegramCheck, error)
}

type DeepSeekChecker interface {
	Check(ctx context.Context) (DeepSeekCheck, error)
}

type GeminiChecker interface {
	Check(ctx context.Context) (GeminiCheck, error)
}
