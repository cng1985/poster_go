package app

import (
	"log/slog"

	"go.uber.org/fx/fxevent"
)

type fxLogger struct {
	logger *slog.Logger
}

func newFXLogger(logger *slog.Logger) fxevent.Logger {
	return fxLogger{logger: logger}
}

func (f fxLogger) LogEvent(event fxevent.Event) {
	switch ev := event.(type) {
	case *fxevent.Started:
		f.logger.Info("fx started")
	case *fxevent.Stopped:
		if ev.Err != nil {
			f.logger.Error("fx stop", "err", ev.Err)
		}
	case *fxevent.Invoked:
		if ev.Err != nil {
			f.logger.Error("fx invoke", "function", ev.FunctionName, "err", ev.Err)
		}
	case *fxevent.OnStartExecuted:
		if ev.Err != nil {
			f.logger.Error("fx onstart", "function", ev.FunctionName, "err", ev.Err)
		}
	case *fxevent.OnStopExecuted:
		if ev.Err != nil {
			f.logger.Error("fx onstop", "function", ev.FunctionName, "err", ev.Err)
		}
	case *fxevent.RollingBack:
		f.logger.Error("fx rollback", "err", ev.StartErr)
	}
}
