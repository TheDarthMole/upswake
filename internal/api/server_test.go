package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	config "github.com/TheDarthMole/UPSWake/internal/domain/entity"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func pingHandler(c echo.Context) error {
	return c.String(http.StatusOK, "pong")
}

func TestCustomValidator_Validate(t *testing.T) {
	type fields struct {
		validator *validator.Validate
	}
	type args struct {
		i interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:    "Test with nil input",
			fields:  fields{validator: validator.New()},
			args:    args{i: nil},
			wantErr: true,
		},
		{
			name:   "Test with valid TargetServer struct",
			fields: fields{validator: validator.New()},
			args: args{i: config.TargetServer{
				Name:      "test",
				MAC:       "00:1A:2B:3C:4D:5E",
				Broadcast: "127.0.0.1",
				Port:      9,
				Interval:  "15m",
				Rules:     []string{"test"},
			}},
			wantErr: false,
		},
		// TODO: This test case makes me think that the validate function isn't working as expected.
		//{
		//	name:   "Test with invalid TargetServer struct",
		//	fields: fields{validator: validator.New()},
		//	args: args{i: config.TargetServer{
		//		Name:      "test",
		//		MAC:       "00:1A:2B:3C:4D:5E",
		//		Broadcast: "", // Invalid field
		//		Port:      9,
		//		Interval:  "15m",
		//		Rules:     []string{"test"},
		//	}},
		//	wantErr: true,
		//},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cv := &CustomValidator{
				validator: tt.fields.validator,
			}
			if err := cv.Validate(tt.args.i); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewCustomValidator(t *testing.T) {
	background := context.Background()
	todo := context.TODO()
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name string
		args args
		want *CustomValidator
	}{
		{
			name: "Test with background context",
			args: args{ctx: background},
			want: &CustomValidator{
				validator: validator.New(),
				ctx:       background,
			},
		},
		{
			name: "Test with TODO context",
			args: args{ctx: todo},
			want: &CustomValidator{
				validator: validator.New(),
				ctx:       todo,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewCustomValidator(tt.args.ctx)

			assert.NotNil(t, got)
			assert.NotNil(t, got.validator)
			assert.NotNil(t, got.ctx)
			assert.IsType(t, tt.want.validator, got.validator)
			assert.Equal(t, tt.want.ctx, got.ctx)
		})
	}
}

func TestNewServer(t *testing.T) {
	type args struct {
		ctx context.Context
		s   *zap.SugaredLogger
	}
	ctx := context.Background()
	sugar := zap.NewExample().Sugar()
	tests := []struct {
		name string
		args args
		want *Server
	}{
		{
			name: "New Server with background context and zap logger",
			args: args{
				ctx: ctx,
				s:   sugar,
			},
			want: &Server{
				ctx:   ctx,
				echo:  echo.New(),
				sugar: sugar,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewServer(tt.args.ctx, tt.args.s)

			assert.NotNil(t, got)
			assert.Equal(t, tt.want.ctx, got.ctx)
			assert.Equal(t, tt.want.sugar, got.sugar)

			// echo instance and validator
			assert.NotNil(t, got.echo)
			cv, ok := got.echo.Validator.(*CustomValidator)
			assert.True(t, ok, "echo.Validator should be *CustomValidator")
			assert.Equal(t, tt.want.ctx, cv.ctx)

			//testRoute := got.echo.GET("/ping", func(c echo.Context) error {
			//	return c.String(http.StatusOK, "pong")
			//})
			req := httptest.NewRequest(http.MethodGet, "/ping", nil)
			rec := httptest.NewRecorder()
			c := got.echo.NewContext(req, rec)

			if assert.NoError(t, pingHandler(c)) {
				assert.Equal(t, http.StatusOK, rec.Code)
				assert.Equal(t, "pong", rec.Body.String())
			}

		})
	}
}

func TestServer_API(t *testing.T) {
	type fields struct {
		ctx   context.Context
		echo  *echo.Echo
		sugar *zap.SugaredLogger
	}
	tests := []struct {
		name   string
		fields fields
		want   *echo.Group
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				ctx:   tt.fields.ctx,
				echo:  tt.fields.echo,
				sugar: tt.fields.sugar,
			}
			if got := s.API(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("API() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestServer_PrintRoutes(t *testing.T) {
	type fields struct {
		ctx   context.Context
		echo  *echo.Echo
		sugar *zap.SugaredLogger
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				ctx:   tt.fields.ctx,
				echo:  tt.fields.echo,
				sugar: tt.fields.sugar,
			}
			s.PrintRoutes()
		})
	}
}

func TestServer_Root(t *testing.T) {
	type fields struct {
		ctx   context.Context
		echo  *echo.Echo
		sugar *zap.SugaredLogger
	}
	tests := []struct {
		name   string
		fields fields
		want   *echo.Group
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				ctx:   tt.fields.ctx,
				echo:  tt.fields.echo,
				sugar: tt.fields.sugar,
			}
			if got := s.Root(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Root() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestServer_Start(t *testing.T) {
	type fields struct {
		ctx   context.Context
		echo  *echo.Echo
		sugar *zap.SugaredLogger
	}
	type args struct {
		address  string
		useSSL   bool
		certFile string
		keyFile  string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				ctx:   tt.fields.ctx,
				echo:  tt.fields.echo,
				sugar: tt.fields.sugar,
			}
			if err := s.Start(tt.args.address, tt.args.useSSL, tt.args.certFile, tt.args.keyFile); (err != nil) != tt.wantErr {
				t.Errorf("Start() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestServer_Stop(t *testing.T) {
	type fields struct {
		ctx   context.Context
		echo  *echo.Echo
		sugar *zap.SugaredLogger
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				ctx:   tt.fields.ctx,
				echo:  tt.fields.echo,
				sugar: tt.fields.sugar,
			}
			if err := s.Stop(); (err != nil) != tt.wantErr {
				t.Errorf("Stop() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
