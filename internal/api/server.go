package api

import (
	"context"

	_ "github.com/TheDarthMole/UPSWake/internal/api/docs" // swaggo docs
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/spf13/afero"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Server struct {
	ctx    context.Context
	cancel context.CancelFunc
	echo   *echo.Echo
	sugar  *zap.SugaredLogger
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

func NewServer(ctx context.Context, s *zap.SugaredLogger) *Server {
	newCtx, cancel := context.WithCancel(ctx)
	app := echo.New()
	app.Validator = NewCustomValidator(ctx)
	app.Pre(middleware.RemoveTrailingSlash())
	app.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus: true,
		LogURI:    true,
		LogValuesFunc: func(c *echo.Context, v middleware.RequestLoggerValues) error {
			if v.Error == nil {
				s.Logw(zapcore.InfoLevel,
					"REQUEST",
					"remote_ip", c.RealIP(),
					"host", c.Request().Host,
					"method", c.Request().Method,
					"uri", v.URI,
					"user_agent", c.Request().UserAgent(),
					"status", v.Status,
				)
			} else {
				s.Logw(zapcore.ErrorLevel,
					"REQUEST_ERROR",
					"remote_ip", c.RealIP(),
					"host", c.Request().Host,
					"method", c.Request().Method,
					"uri", v.URI,
					"user_agent", c.Request().UserAgent(),
					"status", v.Status,
					"error", v.Error.Error(),
				)
			}
			return nil
		},
	}))

	return &Server{
		ctx:    newCtx,
		cancel: cancel,
		echo:   app,
		sugar:  s,
	}
}

func (s *Server) Start(fs afero.Fs, address string, useSSL bool, certFile, keyFile string) error {
	fsFileSystem := afero.NewIOFS(fs)
	start := echo.StartConfig{
		Address:         address,
		HideBanner:      true,
		HidePort:        false,
		CertFilesystem:  fsFileSystem,
		GracefulTimeout: 5,
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
