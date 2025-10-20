package api

import (
	"context"

	_ "github.com/TheDarthMole/UPSWake/internal/api/docs" // swaggo docs
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Server struct {
	ctx   context.Context
	echo  *echo.Echo
	sugar *zap.SugaredLogger
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
	app := echo.New()
	app.Validator = NewCustomValidator(ctx)
	app.Pre(middleware.RemoveTrailingSlash())
	app.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus:   true,
		LogURI:      true,
		LogError:    true,
		HandleError: true, // forwards error to the global error handler, so it can decide appropriate status code
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
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
		ctx:   ctx,
		echo:  app,
		sugar: s,
	}
}

func (s *Server) Start(address string, useSSL bool, certFile, keyFile string) error {
	if useSSL {
		s.echo.Pre(middleware.HTTPSRedirect())
		return s.echo.StartTLS(address, certFile, keyFile)
	}
	return s.echo.Start(address)
}

func (s *Server) Stop() error {
	return s.echo.Shutdown(s.ctx)
}

func (s *Server) Root() *echo.Group {
	return s.echo.Group("")
}

func (s *Server) API() *echo.Group {
	return s.echo.Group("/api")
}
