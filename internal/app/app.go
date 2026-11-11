package app

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"

	"poster.app/studio/internal/platform/config"
	"poster.app/studio/internal/platform/logging"
)

func Run() {
	if config.ShowVersion(os.Args[1:]) {
		fmt.Printf("poster %s\ncommit %s\nbuilt %s\n", Version, Commit, BuiltAt)
		return
	}
	cfg, err := config.Load(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	logger := logging.New(cfg.Logging.Level, cfg.Logging.Format)
	slog.SetDefault(logger)
	logger.Info("starting", cfg.SafeLines()...)
	application := fx.New(
		fx.Supply(cfg, logger),
		Module,
		fx.StartTimeout(60*time.Second),
		fx.StopTimeout(cfg.Runtime.ShutdownTimeout),
		fx.WithLogger(func() fxevent.Logger { return newFXLogger(logger) }),
	)
	application.Run()
}
