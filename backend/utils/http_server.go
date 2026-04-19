package utils

import (
	"context"
	"errors"
	"fst/backend/pkg/config"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func NewHTTPServer(addr string, handler http.Handler) *http.Server {
	readHeaderTimeout := 5 * time.Second
	readTimeout := 5 * time.Second
	writeTimeout := 10 * time.Second
	idleTimeout := 120 * time.Second
	maxHeaderBytes := 1 << 20

	if cfg := config.GlobalConfig; cfg != nil {
		if cfg.HTTPReadHeaderTimeoutSeconds > 0 {
			readHeaderTimeout = time.Duration(cfg.HTTPReadHeaderTimeoutSeconds) * time.Second
		}
		if cfg.HTTPReadTimeoutSeconds > 0 {
			readTimeout = time.Duration(cfg.HTTPReadTimeoutSeconds) * time.Second
		}
		if cfg.HTTPWriteTimeoutSeconds > 0 {
			writeTimeout = time.Duration(cfg.HTTPWriteTimeoutSeconds) * time.Second
		}
		if cfg.HTTPIdleTimeoutSeconds > 0 {
			idleTimeout = time.Duration(cfg.HTTPIdleTimeoutSeconds) * time.Second
		}
		if cfg.HTTPMaxHeaderBytes > 0 {
			maxHeaderBytes = cfg.HTTPMaxHeaderBytes
		}
	}

	return &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
		MaxHeaderBytes:    maxHeaderBytes,
	}
}

func ServeHTTPServer(srv *http.Server, shutdownHook func() error) error {
	errCh := make(chan error, 1)
	go func() {
		err := srv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(quit)

	select {
	case err := <-errCh:
		return err
	case sig := <-quit:
		log.Printf("[Server] 正在关闭，收到信号: %s", sig.String())

		shutdownErr := shutdownHTTPServer(srv)
		hookErr := runShutdownHook(shutdownHook)
		if shutdownErr != nil || hookErr != nil {
			return errors.Join(shutdownErr, hookErr)
		}

		log.Println("[Server] 已关闭")
		return nil
	}
}

func shutdownHTTPServer(srv *http.Server) error {
	shutdownTimeout := 10 * time.Second
	if cfg := config.GlobalConfig; cfg != nil && cfg.HTTPShutdownTimeoutSeconds > 0 {
		shutdownTimeout = time.Duration(cfg.HTTPShutdownTimeoutSeconds) * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		_ = srv.Close()
		return err
	}
	return nil
}

func runShutdownHook(shutdownHook func() error) error {
	if shutdownHook == nil {
		return nil
	}
	return shutdownHook()
}
