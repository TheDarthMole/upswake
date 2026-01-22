package api

import (
	"context"
	"time"
	"log/slog"

	_ "github.com/TheDarthMole/UPSWake/internal/api/docs" // swaggo docs
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/spf13/afero"
)

type Server struct {
	ctx    context.Context
	cancel context.CancelFunc
	echo   *echo.Echo
	logger *slog.Logger
}

type CustomValidator struct {
	validator *validator.Validate
	ctx       context.Context
}

func NewCustomValidator(ctx context.Context) *CustomValidator {
	return &CustomValidator{
		validator: validator.New(),
		ctx:       ctx,
	}
}

func (cv *CustomValidator) Validate(i any) error {
	if err := cv.validator.StructCtx(cv.ctx, i); err != nil {
		return err
	}
	return nil
}

func NewServer(ctx context.Context, logger *slog.Logger) *Server {
	newCtx, cancel := context.WithCancel(ctx)
	app := echo.New()
	// TODO: Set logger to the app here
	//app.Logger = logger

	app.Validator = NewCustomValidator(newCtx)
	app.Pre(middleware.RemoveTrailingSlash())
	app.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus: true,
		LogURI:    true,
		LogValuesFunc: func(c *echo.Context, v middleware.RequestLoggerValues) error {
			if v.Error == nil {
				logger.Info(
					"REQUEST",
					slog.String("remote_ip", c.RealIP()),
					slog.String("host", c.Request().Host),
					slog.String("method", c.Request().Method),
					slog.String("uri", v.URI),
					slog.String("user_agent", c.Request().UserAgent()),
					slog.Int("status", v.Status),
				)
			} else {
				logger.Error(
					"REQUEST_ERROR",
					slog.String("remote_ip", c.RealIP()),
					slog.String("host", c.Request().Host),
					slog.String("method", c.Request().Method),
					slog.String("uri", v.URI),
					slog.String("user_agent", c.Request().UserAgent()),
					slog.Int("status", v.Status),
					slog.Any("error", v.Error),
				)
			}
			return nil
		},
	}))

	return &Server{
		ctx:    newCtx,
		cancel: cancel,
		echo:   app,
		logger: logger,
	}
}

func (s *Server) Start(fs afero.Fs, address string, useSSL bool, certFile, keyFile string) error {
	fsFileSystem := afero.NewIOFS(fs)
	start := echo.StartConfig{
		Address:         address,
		HideBanner:      true,
		HidePort:        false,
		CertFilesystem:  fsFileSystem,
		GracefulTimeout: 5 * time.Second,
	}
	if useSSL {
		s.echo.Pre(middleware.HTTPSRedirect())
		return start.StartTLS(s.ctx, s.echo, certFile, keyFile)
	}

	return start.Start(s.ctx, s.echo)
}

func (s *Server) Stop() error {
	s.cancel()
	return nil
}

func (s *Server) Root() *echo.Group {
	return s.echo.Group("")
}

func (s *Server) API() *echo.Group {
	return s.echo.Group("/api")
}
