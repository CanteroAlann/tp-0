package signals

import (
	"context"
	"io"
	"os/signal"
	"syscall"

	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/logger"
)

func SetupSignalHandler(closers ...io.Closer) (context.Context, context.CancelFunc) {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)

	go func() {
		<-ctx.Done()
		logger.Warn("client-signal", logger.InProgress, "msg", "SIGTERM received, closing network resources")
		for _, closer := range closers {
			if closer != nil {
				_ = closer.Close()
			}
		}
	}()

	return ctx, stop
}
