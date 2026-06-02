package rpc

import (
	"net/http"

	"wakatime/pkg/db"

	"github.com/vmkteam/embedlog"
	zm "github.com/vmkteam/zenrpc-middleware"
	"github.com/vmkteam/zenrpc/v2"
)

var (
	//nolint:unused
	ErrInvToken = zenrpc.NewStringError(1001, "invalid_token")
	//nolint:unused
	ErrWakaUnavailable = zenrpc.NewStringError(1002, "wakatime_unavailable")
	ErrValidation      = zenrpc.NewStringError(1003, "validation_error")
)

var allowDebugFn = func() zm.AllowDebugFunc {
	return func(req *http.Request) bool {
		return req != nil && req.FormValue("__level") == "5"
	}
}

//go:generate zenrpc

// New returns new zenrpc Server.
func New(dbo db.DB, logger embedlog.Logger, isDevel bool) *zenrpc.Server {
	rpc := zenrpc.NewServer(zenrpc.Options{
		ExposeSMD: true,
		AllowCORS: true,
	})

	rpc.Use(
		zm.WithDevel(isDevel),
		zm.WithHeaders(),
		zm.WithSentry(zm.DefaultServerName),
		zm.WithNoCancelContext(),
		zm.WithMetrics(zm.DefaultServerName),
		zm.WithTiming(isDevel, allowDebugFn()),
		zm.WithSQLLogger(dbo.DB, isDevel, allowDebugFn(), allowDebugFn()),
	)

	rpc.Use(
		zm.WithSLog(logger.Print, zm.DefaultServerName, nil),
		zm.WithErrorSLog(logger.Print, zm.DefaultServerName, nil),
	)

	// services
	rpc.RegisterAll(map[string]zenrpc.Invoker{
		// "sample": NewSampleService(db, logger),
		"wakatime": NewWakaTimeService(),
	})

	return rpc
}

//nolint:unused
func newInternalError(err error) *zenrpc.Error {
	return zenrpc.NewError(http.StatusInternalServerError, err)
}

func newValidationError(err error) *zenrpc.Error {
	return zenrpc.NewError(http.StatusBadRequest, err)
}
