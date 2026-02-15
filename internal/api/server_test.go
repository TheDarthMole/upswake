package api

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	cryptRand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"math/rand/v2"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	config "github.com/TheDarthMole/UPSWake/internal/domain/entity"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestLoggerWithBuffer() (*slog.Logger, *bytes.Buffer) {
	logBuf := new(bytes.Buffer)
	handler := slog.NewJSONHandler(logBuf, nil)
	logger := slog.New(handler)
	return logger, logBuf
}

func newTestLogger() *slog.Logger {
	logger, _ := newTestLoggerWithBuffer()
	return logger
}

func pingHandler(c *echo.Context) error {
	return c.String(http.StatusOK, "pong")
}

func TestCustomValidator_Validate(t *testing.T) {
	type fields struct {
		validator *validator.Validate
	}
	type args struct {
		i any
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		error  bool
	}{
		{
			name:   "Test with nil input",
			fields: fields{validator: validator.New()},
			args:   args{i: nil},
			error:  true,
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
			error: false,
		},
		// TODO: This test case makes me think that the validate function isn't working as expected.
		// It uses the `validate` struct tag for validation, but we are using .Validate() methods
		// {
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
		//	error: true,
		// },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cv := &CustomValidator{
				validator: tt.fields.validator,
			}
			err := cv.Validate(tt.args.i)
			assert.Equal(t, tt.error, err != nil)
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
		ctx    context.Context
		logger *slog.Logger
	}
	ctx := context.Background()
	logger := newTestLogger()
	tests := []struct {
		name string
		args args
		want *Server
	}{
		{
			name: "New Server with background context and slog logger",
			args: args{
				ctx:    ctx,
				logger: logger,
			},
			want: &Server{
				ctx:    ctx,
				echo:   echo.New(),
				logger: logger,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewServer(tt.args.ctx, tt.args.logger)

			assert.NotNil(t, got)
			assert.Equal(t, tt.want.logger, got.logger)

			// echo instance and validator
			assert.NotNil(t, got.echo)
			assert.IsType(t, &CustomValidator{}, got.echo.Validator)
			assert.NotNil(t, got.echo.Validator)

			req := httptest.NewRequest(http.MethodGet, "/ping", http.NoBody)
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
	logger := newTestLogger()

	e := NewServer(t.Context(), logger)
	expected := e.echo.Group("/api")

	assert.Equal(t, expected, e.API())
}

func TestServer_Root(t *testing.T) {
	logger := newTestLogger()

	e := NewServer(context.Background(), logger)
	expected := e.echo.Group("")

	assert.Equal(t, expected, e.Root())
}

func certificateTemplate(t *testing.T) *x509.Certificate {
	notBefore := time.Now()
	notAfter := notBefore.Add(365 * 24 * time.Hour)

	serialNumber, err := cryptRand.Int(cryptRand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	require.NoError(t, err)

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"My Organization"}, //nolint: misspell
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}
	return &template
}

func generateEncodedRSAKeys(t *testing.T) ([]byte, []byte) {
	privateKey, err := rsa.GenerateKey(cryptRand.Reader, 2048)
	require.NoError(t, err)

	template := certificateTemplate(t)

	derBytes, err := x509.CreateCertificate(cryptRand.Reader, template, template, &privateKey.PublicKey, privateKey)
	require.NoError(t, err)

	var pemKey bytes.Buffer
	var pemCert bytes.Buffer

	err = pem.Encode(&pemCert, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	require.NoError(t, err)

	b := x509.MarshalPKCS1PrivateKey(privateKey)

	err = pem.Encode(&pemKey, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: b})
	require.NoError(t, err)

	return pemKey.Bytes(), pemCert.Bytes()
}

func generateEncodedECCKeys(t *testing.T) ([]byte, []byte) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), cryptRand.Reader)
	require.NoError(t, err)

	template := certificateTemplate(t)

	derBytes, err := x509.CreateCertificate(cryptRand.Reader, template, template, &privateKey.PublicKey, privateKey)
	require.NoError(t, err)

	var pemKey bytes.Buffer
	var pemCert bytes.Buffer

	err = pem.Encode(&pemCert, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	require.NoError(t, err)

	b, err := x509.MarshalECPrivateKey(privateKey)
	require.NoError(t, err)

	err = pem.Encode(&pemKey, &pem.Block{Type: "EC PRIVATE KEY", Bytes: b})
	require.NoError(t, err)

	return pemKey.Bytes(), pemCert.Bytes()
}

func TestServer_Start_Stop(t *testing.T) {
	type fields struct {
		ctx    context.Context
		logger *slog.Logger
	}

	certFs := afero.NewMemMapFs()

	privateRSAKey, publicRSAKey := generateEncodedRSAKeys(t)

	require.NoError(t, afero.WriteFile(certFs, "rsa.key", privateRSAKey, os.ModePerm))
	require.NoError(t, afero.WriteFile(certFs, "rsa.cert", publicRSAKey, os.ModePerm))

	privateECCKey, publicECCKey := generateEncodedECCKeys(t)

	require.NoError(t, afero.WriteFile(certFs, "ecc.key", privateECCKey, os.ModePerm))
	require.NoError(t, afero.WriteFile(certFs, "ecc.cert", publicECCKey, os.ModePerm))

	type args struct {
		address  string
		useSSL   bool
		certFile string
		keyFile  string
	}
	tests := []struct {
		name         string
		fields       fields
		args         args
		wantStartErr bool
		wantStopErr  bool
	}{
		{
			name: "Start server without SSL",
			fields: fields{
				ctx:    context.Background(),
				logger: newTestLogger(),
			},
			args: args{
				address:  "127.0.0.1:0",
				useSSL:   false,
				certFile: "",
				keyFile:  "",
			},
			wantStartErr: false,
			wantStopErr:  false,
		},
		{
			name: "Start server with SSL using RSA certs",
			fields: fields{
				ctx:    context.Background(),
				logger: newTestLogger(),
			},
			args: args{
				address:  "127.0.0.1:0",
				useSSL:   true,
				certFile: "rsa.cert",
				keyFile:  "rsa.key",
			},
			wantStartErr: false,
			wantStopErr:  false,
		},
		{
			name: "Start server with SSL using ECC certs",
			fields: fields{
				ctx:    context.Background(),
				logger: newTestLogger(),
			},
			args: args{
				address:  "127.0.0.1:0",
				useSSL:   true,
				certFile: "ecc.cert",
				keyFile:  "ecc.key",
			},
			wantStartErr: false,
			wantStopErr:  false,
		},
		{
			name: "Start server with SSL without certs",
			fields: fields{
				ctx:    context.Background(),
				logger: newTestLogger(),
			},
			args: args{
				address:  "127.0.0.1:0",
				useSSL:   true,
				certFile: "",
				keyFile:  "",
			},
			wantStartErr: true,
			wantStopErr:  false,
		},
		{
			name: "Start server with no port",
			fields: fields{
				ctx:    context.Background(),
				logger: newTestLogger(),
			},
			args: args{
				address:  "127.0.0.1",
				useSSL:   false,
				certFile: "",
				keyFile:  "",
			},
			wantStartErr: true,
			wantStopErr:  false,
		},
		{
			name: "Start server with no address",
			fields: fields{
				ctx:    context.Background(),
				logger: newTestLogger(),
			},
			args: args{
				address:  fmt.Sprintf(":%d", rand.IntN(65535-49152)+49152),
				useSSL:   false,
				certFile: "",
				keyFile:  "",
			},
			wantStartErr: false,
			wantStopErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			srv := NewServer(tt.fields.ctx, tt.fields.logger)

			go func() {
				time.Sleep(500 * time.Millisecond)
				err := srv.Stop()
				if (err != nil) != tt.wantStopErr {
					t.Errorf("Stop() error = %v, error %v", err, tt.wantStopErr)
				}
			}()

			err := srv.Start(certFs, tt.args.address, tt.args.useSSL, tt.args.certFile, tt.args.keyFile)
			// http.ErrServerClosed is returned when the server is shut down normally
			if (err != nil && !errors.Is(err, http.ErrServerClosed)) != tt.wantStartErr {
				t.Errorf("Start() error = %v, error %v", err, tt.wantStartErr)
			}
		})
	}
}
