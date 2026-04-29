package directups

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/TheDarthMole/UPSWake/internal/domain/entity"
	"github.com/TheDarthMole/UPSWake/internal/domain/repository"
	nut "github.com/robbiet480/go.nut"
)

// DirectRepository connects to the NUT server on every call.
// Satisfies repository.UPSRepository.
type DirectRepository struct{}

func NewDirectRepository() repository.UPSRepository {
	return &DirectRepository{}
}

var (
	ErrAuthenticationFailed  = errors.New("failed to authenticate to NUT server")
	ErrFailureAuthenticating = errors.New("an error occurred during authentication to NUT server")
	ErrConnectionFailed      = errors.New("could not connect to NUT server")
)

func connect(host string, port int, username, password string) (*nut.Client, error) {
	client, err := nut.Connect(host, port)
	if err != nil {
		return &nut.Client{}, fmt.Errorf("%w: %w", ErrConnectionFailed, err)
	}

	authenticate, err := client.Authenticate(username, password)
	if err != nil {
		return &nut.Client{}, fmt.Errorf("%w: %w", ErrFailureAuthenticating, err)
	}
	if !authenticate {
		disconnect(&client, host)
		return &nut.Client{}, fmt.Errorf("%w: could not authenticate to NUT server at %s:%d", ErrAuthenticationFailed, host, port)
	}
	return &client, nil
}

func disconnect(client *nut.Client, host string) {
	_, err := client.Disconnect()
	if err != nil {
		slog.Warn("Error disconnecting from NUT server",
			slog.String("host", host),
			slog.Any("error", err))
	}
}

func (*DirectRepository) GetJSON(server *entity.NutServer) (string, error) {
	client, err := connect(server.Host, server.Port, server.Username, server.Password)
	if err != nil {
		return "", err
	}
	defer disconnect(client, server.Host)

	ups, err := client.GetUPSList()
	if err != nil {
		return "", err
	}

	jsonData, err := json.Marshal(ups)
	return string(jsonData), err
}
