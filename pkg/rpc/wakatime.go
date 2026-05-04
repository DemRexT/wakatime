package rpc

import (
	"context"
	"github.com/vmkteam/zenrpc/v2"
)

type WakaTimeService struct {
	zenrpc.Service
}

func NewWakaTimeService() *WakaTimeService {
	return &WakaTimeService{}
}

type User struct {
	ID        int
	Username  string
	SecretKey string
	Status    string
}

type RegisterResult struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Status   string `json:"status"`
}

type TopResult struct {
	Period       string    `json:"period"`
	PeriodStart  string    `json:"periodStart"`
	PeriodEnd    string    `json:"periodEnd"`
	TotalSeconds int       `json:"totalSeconds"`
	Items        []TopItem `json:"items"`
	FetchedAt    string    `json:"fetchedAt"`
}

type TopItem struct {
	Rank     int    `json:"rank"`
	Username string `json:"username"`
}

func (s *WakaTimeService) Register(ctx context.Context, username, secretKey string) (*RegisterResult, error) {
	if username == "" || secretKey[:5] == "waka_" {
		return nil, newInternalError(ErrValidation)
	}

	return &RegisterResult{
		ID:       42,
		Username: "Иван Иванов",
		Status:   "enabled",
	}, nil
}

func (s *WakaTimeService) GetTop(ctx context.Context, period string) (*TopResult, error) {
	return &TopResult{
		Period:       "week",
		PeriodStart:  "2026-04-25",
		PeriodEnd:    "2026-05-01",
		TotalSeconds: 449460,
		Items: []TopItem{
			{
				Rank:     1,
				Username: "Иван Иванов",
			},
		},
		FetchedAt: "2026-05-01T12:35:00Z",
	}, nil
}
