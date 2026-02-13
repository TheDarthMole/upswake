package entity

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func genEcdsaKey() (*ecdsa.PrivateKey, error) {
	return ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
}

func genRsaKey() (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, 2048)
}

func encodeEcdsa(t *testing.T, privateKey *ecdsa.PrivateKey, certificate []byte) ([]byte, []byte) {
	x509Encoded, err := x509.MarshalECPrivateKey(privateKey)
	assert.NoError(t, err)
	pemEncoded := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: x509Encoded})

	pemEncodedCert := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certificate})

	return pemEncoded, pemEncodedCert
}

func encodeRSA(_ *testing.T, privateKey *rsa.PrivateKey, certificate []byte) ([]byte, []byte) {
	x509Encoded := x509.MarshalPKCS1PrivateKey(privateKey)
	pemEncoded := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: x509Encoded})

	pemEncodedCert := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certificate})

	return pemEncoded, pemEncodedCert
}

func generateTestCert(t *testing.T, priv, public any) ([]byte, error) {
	t.Helper()

	notBefore := time.Now()
	notAfter := notBefore.Add(1 * time.Hour)
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	assert.NoError(t, err)

	template := x509.Certificate{
		SerialNumber:          serialNumber,
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		Subject: pkix.Name{
			Organization: []string{"Test Organization"}, //nolint:misspell
		},
	}

	return x509.CreateCertificate(rand.Reader, &template, &template, public, priv)
}

func genRSACertAndKey(t *testing.T) ([]byte, []byte) {
	rsaKey, err := genRsaKey()
	assert.NoError(t, err)
	rsaCert, err := generateTestCert(t, rsaKey, &rsaKey.PublicKey)
	assert.NoError(t, err)
	rsaKeyPEM, rsaCertPEM := encodeRSA(t, rsaKey, rsaCert)
	return rsaKeyPEM, rsaCertPEM
}

func genEcdsaCertAndKey(t *testing.T) ([]byte, []byte) {
	ecdsaKey, err := genEcdsaKey()
	assert.NoError(t, err)
	ecdsaCert, err := generateTestCert(t, ecdsaKey, &ecdsaKey.PublicKey)
	assert.NoError(t, err)
	ecdsaKeyPEM, ecdsaCertPEM := encodeEcdsa(t, ecdsaKey, ecdsaCert)
	return ecdsaKeyPEM, ecdsaCertPEM
}

func TestCLIArgs_URL(t *testing.T) {
	type fields struct {
		ConfigFile string
		UseSSL     bool
		CertFile   string
		KeyFile    string
		Host       net.IP
		Port       string
		TLSConfig  *tls.Config
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "HTTP Port 8080 on 127.0.0.1",
			fields: fields{
				Host:   net.ParseIP("127.0.0.1"),
				Port:   "8080",
				UseSSL: false,
			},
			want: "http://127.0.0.1:8080",
		},
		{
			name: "HTTPS Port 8443 on 127.0.0.1",
			fields: fields{
				Host:   net.ParseIP("127.0.0.1"),
				Port:   "8443",
				UseSSL: true,
			},
			want: "https://127.0.0.1:8443",
		},
		{
			name: "HTTPS Port 8443 on 1.2.3.4",
			fields: fields{
				Host:   net.ParseIP("1.2.3.4"),
				Port:   "8443",
				UseSSL: true,
			},
			want: "https://1.2.3.4:8443",
		},
		{
			name: "HTTP Port 8080 on 1.2.3.4",
			fields: fields{
				Host:   net.ParseIP("1.2.3.4"),
				Port:   "8080",
				UseSSL: false,
			},
			want: "http://1.2.3.4:8080",
		},
		{
			name: "HTTP Port 8080 on 0.0.0.0",
			fields: fields{
				Host:   net.ParseIP("0.0.0.0"),
				Port:   "8080",
				UseSSL: false,
			},
			want: "http://127.0.0.1:8080",
		},
		{
			name: "HTTP Port 8080 on ::",
			fields: fields{
				Host:   net.ParseIP("::"),
				Port:   "8080",
				UseSSL: false,
			},
			want: "http://127.0.0.1:8080",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CLIArgs{
				ConfigFile: tt.fields.ConfigFile,
				UseSSL:     tt.fields.UseSSL,
				CertFile:   tt.fields.CertFile,
				KeyFile:    tt.fields.KeyFile,
				Host:       tt.fields.Host,
				Port:       tt.fields.Port,
				TLSConfig:  tt.fields.TLSConfig,
			}
			if got := c.URL(); got != tt.want {
				t.Errorf("URL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCLIArgs_ListenAddress(t *testing.T) {
	type fields struct {
		ConfigFile string
		UseSSL     bool
		CertFile   string
		KeyFile    string
		Host       net.IP
		Port       string
		TLSConfig  *tls.Config
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "Port 8080 on 127.0.0.1",
			fields: fields{
				Host:   net.ParseIP("127.0.0.1"),
				Port:   "8080",
				UseSSL: false,
			},
			want: "127.0.0.1:8080",
		},
		{
			name: "Port 8443 on 127.0.0.1",
			fields: fields{
				Host:   net.ParseIP("127.0.0.1"),
				Port:   "8443",
				UseSSL: true,
			},
			want: "127.0.0.1:8443",
		},
		{
			name: "Port 8443 on 1.2.3.4",
			fields: fields{
				Host:   net.ParseIP("1.2.3.4"),
				Port:   "8443",
				UseSSL: true,
			},
			want: "1.2.3.4:8443",
		},
		{
			name: "Port 8080 on 1.2.3.4",
			fields: fields{
				Host:   net.ParseIP("1.2.3.4"),
				Port:   "8080",
				UseSSL: false,
			},
			want: "1.2.3.4:8080",
		},
		{
			name: "Port 8080 on 0.0.0.0",
			fields: fields{
				Host:   net.ParseIP("0.0.0.0"),
				Port:   "8080",
				UseSSL: false,
			},
			want: "[::]:8080",
		},
		{
			name: "Port 8080 on ::",
			fields: fields{
				Host:   net.ParseIP("::"),
				Port:   "8080",
				UseSSL: false,
			},
			want: "[::]:8080",
		},
		{
			name: "HTTPS Port 8443 on 2001:db8::1",
			fields: fields{
				Host:   net.ParseIP("2001:db8::1"),
				Port:   "8443",
				UseSSL: true,
			},
			want: "[2001:db8::1]:8443",
		},
		{
			name: "HTTP Port 8080 on 2001:db8::1",
			fields: fields{
				Host:   net.ParseIP("2001:db8::1"),
				Port:   "8080",
				UseSSL: false,
			},
			want: "[2001:db8::1]:8080",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CLIArgs{
				ConfigFile: tt.fields.ConfigFile,
				UseSSL:     tt.fields.UseSSL,
				CertFile:   tt.fields.CertFile,
				KeyFile:    tt.fields.KeyFile,
				Host:       tt.fields.Host,
				Port:       tt.fields.Port,
				TLSConfig:  tt.fields.TLSConfig,
			}
			if got := c.ListenAddress(); got != tt.want {
				t.Errorf("ListenAddress() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCLIArgs_URLPrefix(t *testing.T) {
	type fields struct {
		ConfigFile string
		UseSSL     bool
		CertFile   string
		KeyFile    string
		Host       net.IP
		Port       string
		TLSConfig  *tls.Config
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "Test HTTP URL Prefix",
			fields: fields{
				UseSSL: false,
			},
			want: "http://",
		},
		{
			name: "Test HTTPS URL Prefix",
			fields: fields{
				UseSSL: true,
			},
			want: "https://",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CLIArgs{
				ConfigFile: tt.fields.ConfigFile,
				UseSSL:     tt.fields.UseSSL,
				CertFile:   tt.fields.CertFile,
				KeyFile:    tt.fields.KeyFile,
				Host:       tt.fields.Host,
				Port:       tt.fields.Port,
				TLSConfig:  tt.fields.TLSConfig,
			}
			if got := c.URLPrefix(); got != tt.want {
				t.Errorf("URLPrefix() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCLIArgs_Validate(t *testing.T) {
	type fields struct {
		ConfigFile string
		UseSSL     bool
		CertFile   string
		KeyFile    string
		Host       net.IP
		Port       string
		TLSConfig  *tls.Config
	}
	tests := []struct {
		name   string
		fields fields
		error  error
	}{
		{
			name: "Valid HTTP Config",
			fields: fields{
				ConfigFile: "",
				UseSSL:     false,
				CertFile:   "certs/server.cert",
				KeyFile:    "certs/server.key",
				Host:       net.ParseIP("127.0.0.1"),
				Port:       "8080",
				TLSConfig:  nil,
			},
			error: nil,
		},
		{
			name: "Valid HTTPS Config",
			fields: fields{
				ConfigFile: "",
				UseSSL:     true,
				CertFile:   "certs/server.cert",
				KeyFile:    "certs/server.key",
				Host:       net.ParseIP("127.0.0.1"),
				Port:       "8443",
				TLSConfig:  &tls.Config{},
			},
			error: nil,
		},
		{
			name: "HTTPS Config without Certfile",
			fields: fields{
				ConfigFile: "",
				UseSSL:     true,
				CertFile:   "",
				KeyFile:    "certs/server.key",
				Host:       net.ParseIP("127.0.0.1"),
				Port:       "8443",
				TLSConfig:  &tls.Config{},
			},
			error: ErrCertFilesNotSet,
		},
		{
			name: "HTTPS Config without Keyfile",
			fields: fields{
				ConfigFile: "",
				UseSSL:     true,
				CertFile:   "certs/server.cert",
				KeyFile:    "",
				Host:       net.ParseIP("127.0.0.1"),
				Port:       "8443",
				TLSConfig:  &tls.Config{},
			},
			error: ErrCertFilesNotSet,
		},
		{
			name: "HTTPS Config without TLSConfig",
			fields: fields{
				ConfigFile: "",
				UseSSL:     true,
				CertFile:   "certs/server.cert",
				KeyFile:    "certs/server.key",
				Host:       net.ParseIP("127.0.0.1"),
				Port:       "8443",
				TLSConfig:  nil,
			},
			error: ErrTLSConfigNotSet,
		},
		{
			name: "HTTP Config without valid host",
			fields: fields{
				ConfigFile: "",
				UseSSL:     false,
				CertFile:   "certs/server.cert",
				KeyFile:    "certs/server.key",
				Host:       net.ParseIP("999.999.999.999"),
				Port:       "8080",
				TLSConfig:  nil,
			},
			error: ErrHostRequired,
		},
		{
			name: "HTTP Config with non-integer port",
			fields: fields{
				ConfigFile: "",
				UseSSL:     false,
				CertFile:   "certs/server.cert",
				KeyFile:    "certs/server.key",
				Host:       net.ParseIP("127.0.0.1"),
				Port:       "invalidPort",
				TLSConfig:  nil,
			},
			error: ErrInvalidPort,
		},
		{
			name: "HTTP Config with port too large",
			fields: fields{
				ConfigFile: "",
				UseSSL:     false,
				CertFile:   "certs/server.cert",
				KeyFile:    "certs/server.key",
				Host:       net.ParseIP("127.0.0.1"),
				Port:       "999999999",
				TLSConfig:  nil,
			},
			error: ErrInvalidPort,
		},
		{
			name: "HTTP Config with port too small",
			fields: fields{
				ConfigFile: "",
				UseSSL:     false,
				CertFile:   "certs/server.cert",
				KeyFile:    "certs/server.key",
				Host:       net.ParseIP("127.0.0.1"),
				Port:       "-1",
				TLSConfig:  nil,
			},
			error: ErrInvalidPort,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CLIArgs{
				ConfigFile: tt.fields.ConfigFile,
				UseSSL:     tt.fields.UseSSL,
				CertFile:   tt.fields.CertFile,
				KeyFile:    tt.fields.KeyFile,
				Host:       tt.fields.Host,
				Port:       tt.fields.Port,
				TLSConfig:  tt.fields.TLSConfig,
			}
			err := c.Validate()

			assert.ErrorIs(t, err, tt.error)
		})
	}
}

func TestCLIArgs_x509Cert(t *testing.T) {
	fileSystem := afero.NewMemMapFs()
	rsaPrivPEM, rsaCertPEM := genRSACertAndKey(t)
	ecdsaPrivPEM, ecdsaCertPEM := genEcdsaCertAndKey(t)
	assert.NoError(t, afero.WriteFile(fileSystem, "rsaServer.cert", rsaCertPEM, 0o644))
	assert.NoError(t, afero.WriteFile(fileSystem, "rsaServer.key", rsaPrivPEM, 0o644))
	assert.NoError(t, afero.WriteFile(fileSystem, "ecdsaServer.cert", ecdsaCertPEM, 0o644))
	assert.NoError(t, afero.WriteFile(fileSystem, "ecdsaServer.key", ecdsaPrivPEM, 0o644))
	assert.NoError(t, afero.WriteFile(fileSystem, "invalidServer.cert", []byte("invalid cert"), 0o644))
	assert.NoError(t, afero.WriteFile(fileSystem, "invalidServer.key", []byte("invalid key"), 0o644))

	rsaKey, err := genRsaKey()
	assert.NoError(t, err)
	_, invalidRSACertPEM := encodeRSA(t, rsaKey, []byte("invalid cert"))
	assert.NoError(t, afero.WriteFile(fileSystem, "invalidCert.cert", invalidRSACertPEM, 0o644))
	type fields struct {
		ConfigFile string
		UseSSL     bool
		CertFile   string
		KeyFile    string
		Host       net.IP
		Port       string
		TLSConfig  *tls.Config
	}
	tests := []struct {
		name   string
		fields fields
		error  error
	}{
		{
			name: "Valid RSA Cert",
			fields: fields{
				CertFile: "rsaServer.cert",
			},
			error: nil,
		},
		{
			name: "Valid ecdsa Cert",
			fields: fields{
				CertFile: "ecdsaServer.cert",
			},
			error: nil,
		},
		{
			name: "Invalid Cert format",
			fields: fields{
				CertFile: "invalidServer.cert",
			},
			error: ErrFailedParsePEM,
		},
		{
			name: "Cert file does not exist",
			fields: fields{
				CertFile: "doesNotExist.cert",
			},
			error: ErrFailedReadCertFile,
		},
		{
			name: "Cert file PEM encoded but invalid",
			fields: fields{
				CertFile: "invalidCert.cert",
			},
			error: ErrFailedParsePEM,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CLIArgs{
				ConfigFile: tt.fields.ConfigFile,
				UseSSL:     tt.fields.UseSSL,
				CertFile:   tt.fields.CertFile,
				KeyFile:    tt.fields.KeyFile,
				Host:       tt.fields.Host,
				Port:       tt.fields.Port,
				TLSConfig:  tt.fields.TLSConfig,
			}
			_, err := c.x509Cert(fileSystem)
			assert.ErrorIs(t, err, tt.error)
		})
	}
}

func TestNewCLIArgs(t *testing.T) {
	fileSystem := afero.NewMemMapFs()
	ecdsaPrivPEM, ecdsaCertPEM := genEcdsaCertAndKey(t)
	assert.NoError(t, afero.WriteFile(fileSystem, "ecdsaServer.cert", ecdsaCertPEM, 0o644))
	assert.NoError(t, afero.WriteFile(fileSystem, "ecdsaServer.key", ecdsaPrivPEM, 0o644))
	type args struct {
		configFile string
		useSSL     bool
		certFile   string
		keyFile    string
		host       string
		port       string
		fileSystem afero.Fs
	}
	tests := []struct {
		name  string
		args  args
		want  *CLIArgs
		error error
	}{
		{
			name: "Valid HTTP Config",
			args: args{
				configFile: "",
				useSSL:     false,
				certFile:   "",
				keyFile:    "",
				host:       "127.0.0.1",
				port:       "8080",
				fileSystem: fileSystem,
			},
			want: &CLIArgs{
				ConfigFile: "",
				UseSSL:     false,
				CertFile:   "",
				KeyFile:    "",
				Host:       net.ParseIP("127.0.0.1"),
				Port:       "8080",
				TLSConfig:  nil,
			},
			error: nil,
		},
		{
			name: "Valid HTTPS Config",
			args: args{
				configFile: "",
				useSSL:     true,
				certFile:   "ecdsaServer.cert",
				keyFile:    "ecdsaServer.key",
				host:       "127.0.0.1",
				port:       "8443",
				fileSystem: fileSystem,
			},
			want: &CLIArgs{
				ConfigFile: "",
				UseSSL:     true,
				CertFile:   "ecdsaServer.cert",
				KeyFile:    "ecdsaServer.key",
				Host:       net.ParseIP("127.0.0.1"),
				Port:       "8443",
				TLSConfig:  &tls.Config{},
			},
			error: nil,
		},
		{
			name: "Invalid HTTPS Config",
			args: args{
				configFile: "",
				useSSL:     true,
				certFile:   "ecdsaServer.cert",
				keyFile:    "ecdsaServer.key",
				host:       "999.999.999.999",
				port:       "8080",
				fileSystem: fileSystem,
			},
			want:  nil,
			error: ErrHostRequired,
		},
		{
			name: "HTTPS Certificate Not Found",
			args: args{
				configFile: "",
				useSSL:     true,
				certFile:   "not-found.cert",
				keyFile:    "ecdsaServer.key",
				host:       "127.0.0.1",
				port:       "8443",
				fileSystem: fileSystem,
			},
			want:  nil,
			error: ErrFailedReadCertFile,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewCLIArgs(tt.args.fileSystem, tt.args.configFile, tt.args.useSSL, tt.args.certFile, tt.args.keyFile, tt.args.host, tt.args.port)
			assert.ErrorIs(t, err, tt.error)

			if tt.error != nil {
				return
			}
			assert.NotNil(t, got)
			assert.Equal(t, tt.want.ConfigFile, got.ConfigFile, "ConfigFile")
			assert.Equal(t, tt.want.UseSSL, got.UseSSL, "UseSSL")
			assert.Equal(t, tt.want.CertFile, got.CertFile, "CertFile")
			assert.Equal(t, tt.want.KeyFile, got.KeyFile, "KeyFile")
			assert.Equal(t, tt.want.Host, got.Host, "Host")
			assert.Equal(t, tt.want.Port, got.Port, "Port")
			// comparing tls config is tricky, so just check if it's nil or not
			if tt.want.UseSSL {
				assert.NotNil(t, got.TLSConfig, "TLSConfig should not be nil")
			}
		})
	}
}
