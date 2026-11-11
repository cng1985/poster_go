package server

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"gorm.io/gorm"

	"poster.app/studio/internal/platform/config"
	apperrors "poster.app/studio/internal/platform/errors"
	"poster.app/studio/internal/platform/middleware"
	"poster.app/studio/internal/platform/response"
	"poster.app/studio/internal/platform/routing"
)

type State struct {
	ready atomic.Bool
}

func NewState() *State { return &State{} }

func (s *State) MarkReady() { s.ready.Store(true) }

func (s *State) MarkNotReady() { s.ready.Store(false) }

func (s *State) Ready() bool { return s.ready.Load() }

type engineIn struct {
	fx.In
	Config *config.Config
	Logger *slog.Logger
	Routes []routing.Registrar `group:"routes"`
	Auth   gin.HandlerFunc     `name:"authMW"`
	State  *State
	DB     *gorm.DB
}

func NewEngineFromFX(in engineIn) (*gin.Engine, error) {
	return NewEngine(in.Config, in.Logger, in.Routes, in.Auth, in.State, in.DB)
}

func NewEngine(cfg *config.Config, logger *slog.Logger, routes []routing.Registrar, auth gin.HandlerFunc, state *State, db *gorm.DB) (*gin.Engine, error) {
	if cfg.Logging.Level == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	engine := gin.New()
	if err := engine.SetTrustedProxies(cfg.Server.TrustedProxies); err != nil {
		return nil, err
	}
	engine.HandleMethodNotAllowed = true
	engine.Use(
		middleware.RequestID(),
		middleware.ClientIP(),
		middleware.SecurityHeaders(),
		middleware.Recovery(logger),
		middleware.AccessLog(logger),
		middleware.RequestTimeout(cfg.Runtime.RequestTimeout),
		middleware.BodyLimit(cfg.Server.MaxBodyBytes, cfg.Upload.MaxBytes),
	)
	engine.GET("/health/live", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "live"})
	})
	engine.GET("/health/ready", func(c *gin.Context) {
		if state == nil || !state.Ready() {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not_ready"})
			return
		}
		sqlDB, err := db.DB()
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not_ready"})
			return
		}
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		if err := sqlDB.PingContext(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not_ready"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})
	api := engine.Group("/api")
	public := api.Group("")
	protected := api.Group("", auth)
	for _, route := range routes {
		route.RegisterPublic(public)
		route.RegisterProtected(protected)
	}
	engine.NoMethod(func(c *gin.Context) {
		if isAPI(c.Request.URL.Path) {
			response.HandleError(c, apperrors.MethodNotAllowed())
			return
		}
		c.Status(http.StatusMethodNotAllowed)
	})
	engine.NoRoute(func(c *gin.Context) {
		if isAPI(c.Request.URL.Path) {
			response.HandleError(c, apperrors.NotFound("接口不存在"))
			return
		}
		c.Status(http.StatusNotFound)
	})
	return engine, nil
}

func isAPI(path string) bool {
	return path == "/api" || strings.HasPrefix(path, "/api/")
}

type httpIn struct {
	fx.In
	Lifecycle  fx.Lifecycle
	Engine     *gin.Engine
	Config     *config.Config
	Logger     *slog.Logger
	State      *State
	Shutdowner fx.Shutdowner
}

func RegisterHTTP(in httpIn) {
	var server *http.Server
	in.Lifecycle.Append(fx.Hook{
		OnStart: func(context.Context) error {
			addr := net.JoinHostPort(in.Config.Server.Host, strconv.Itoa(in.Config.Server.Port))
			listener, err := net.Listen("tcp", addr)
			if err != nil {
				return err
			}
			server = &http.Server{
				Handler:           in.Engine,
				ReadHeaderTimeout: in.Config.Runtime.ReadHeaderTimeout,
				IdleTimeout:       in.Config.Runtime.IdleTimeout,
				MaxHeaderBytes:    in.Config.Server.MaxHeaderBytes,
			}
			go func() {
				err := server.Serve(listener)
				if err != nil && !errors.Is(err, http.ErrServerClosed) {
					in.Logger.Error("http stopped", "err", err)
					_ = in.Shutdowner.Shutdown()
				}
			}()
			in.Logger.Info("http listening", "addr", addr)
			return nil
		},
		OnStop: func(ctx context.Context) error {
			in.State.MarkNotReady()
			if server == nil {
				return nil
			}
			return server.Shutdown(ctx)
		},
	})
}
