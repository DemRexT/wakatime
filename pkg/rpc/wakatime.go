package rpc

import (
	"context"
	"strings"
	"time"

	"github.com/vmkteam/zenrpc/v2"
)

type WakaTimeService struct {
	zenrpc.Service
}

func NewWakaTimeService() *WakaTimeService {
	return &WakaTimeService{}
}

type RegisterResult struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Status   string `json:"status"`
}

type TopResult struct {
	Period       string    `json:"period"`
	PeriodStart  time.Time `json:"periodStart"`
	PeriodEnd    time.Time `json:"periodEnd"`
	TotalSeconds int       `json:"totalSeconds"`
	Items        []TopItem `json:"items"`
	FetchedAt    time.Time `json:"fetchedAt"`
}

type TopItem struct {
	Rank     int    `json:"rank"`
	Username string `json:"username"`
}

func (s WakaTimeService) Register(ctx context.Context, username, secretAPIKey string) (*RegisterResult, error) {
	if username == "" || !strings.HasPrefix(secretAPIKey, "waka_") {
		return nil, newValidationError(ErrValidation)
	}

	return &RegisterResult{
		ID:       42,
		Username: username,
		Status:   "enabled",
	}, nil
}

func (s WakaTimeService) GetTop(ctx context.Context, period string) (*TopResult, error) {
	switch period {
	case "week", "month", "all":
	default:
		return nil, newValidationError(ErrValidation)
	}

	return &TopResult{
		Period:       period,
		PeriodStart:  time.Date(2026, 4, 25, 0, 0, 0, 0, time.UTC),
		PeriodEnd:    time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
		TotalSeconds: 449460,
		Items: []TopItem{
			{
				Rank:     1,
				Username: "Иван Иванов",
			},
		},
		FetchedAt: time.Date(2026, 5, 1, 12, 35, 0, 0, time.UTC),
	}, nil
}
