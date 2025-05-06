package devlog

import (
	"context"
	"fmt"
	"io"
	"log"
	"math/rand"
	"time"

	"connectrpc.com/connect"
)

type InterceptorConfig struct {
	Log              io.Writer
	HiddenHeaders    []string
	RedactValue      func(string) string
	MaxBufferSize    int
	ContextExtractor func(context.Context) map[string]string
}

type loggingInterceptor struct {
	config *InterceptorConfig
	logger *log.Logger
}

func NewInterceptor(config *InterceptorConfig) connect.Interceptor {
	if config == nil {
		config = &InterceptorConfig{}
	}
	if config.Log == nil {
		config.Log = log.Writer()
	}
	if config.RedactValue == nil {
		config.RedactValue = func(s string) string { return "[REDACTED]" }
	}
	if config.HiddenHeaders == nil {
		config.HiddenHeaders = []string{"authorization", "token", "cookie"}
	}
	if config.ContextExtractor == nil {
		config.ContextExtractor = func(ctx context.Context) map[string]string { return nil }
	}

	return &loggingInterceptor{
		config: config,
		logger: log.New(config.Log, "", log.LstdFlags),
	}
}

func (i *loggingInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		reqID := generateRequestID()
		start := time.Now()

		// Логирование запроса
		lb := getBuilder(i.config)
		i.logRequest(ctx, lb, reqID, req)
		lb.writeTo(i.logger.Writer())
		putBuilder(lb)

		resp, err := next(ctx, req)
		duration := time.Since(start)

		// Логирование ответа
		lb = getBuilder(i.config)
		i.logResponse(ctx, lb, reqID, req, resp, err, duration)
		lb.writeTo(i.logger.Writer())
		putBuilder(lb)

		return resp, err
	}
}

func (i *loggingInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		reqID := generateRequestID()

		// Логирование начала стрима
		lb := getBuilder(i.config)
		i.logStreamStart(ctx, lb, reqID, conn)
		lb.writeTo(i.logger.Writer())
		putBuilder(lb)

		wrapped := &loggingHandlerStreamConn{
			StreamingHandlerConn: conn,
			interceptor:          i,
			ctx:                  ctx,
			start:                time.Now(),
			reqID:                reqID,
		}

		err := next(ctx, wrapped)

		// Логирование завершения стрима
		lb = getBuilder(i.config)
		i.logStreamEnd(ctx, lb, wrapped, err)
		lb.writeTo(i.logger.Writer())
		putBuilder(lb)

		return err
	}
}

func (i *loggingInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return func(ctx context.Context, spec connect.Spec) connect.StreamingClientConn {
		return next(ctx, spec)
	}
}

func generateRequestID() string {
	return fmt.Sprintf("%d-%04x", time.Now().Unix(), rand.Intn(0x10000))
}
