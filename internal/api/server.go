package api

import (
	"context"
	_ "github.com/TheDarthMole/UPSWake/internal/api/docs"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
	"go.uber.org/zap"
)

type Server struct {
	ctx   context.Context
	echo  *echo.Echo
	sugar *zap.SugaredLogger
}

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		return echo.NewHTTPError(400, err.Error())
	}
	return nil
}

func NewServer(ctx context.Context, s *zap.SugaredLogger) *Server {
	app := echo.New()
	app.Validator = &CustomValidator{validator: validator.New()}
	app.Pre(middleware.RemoveTrailingSlash())
	app.Use(middleware.Logger())

	app.GET("/swagger/*", echoSwagger.WrapHandler)

	return &Server{
		ctx:   ctx,
		echo:  app,
		sugar: s,
	}
}

func (s *Server) Start(address string) error {
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

func (s *Server) PrintRoutes() {
	for _, route := range s.echo.Routes() {
		s.sugar.Infof("%s %s", route.Method, route.Path)
	}
}
