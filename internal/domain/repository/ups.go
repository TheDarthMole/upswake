package repository

import "github.com/TheDarthMole/UPSWake/internal/domain/entity"

// UPSRepository provides access to UPS status data from NUT servers.
// Implementations can cache results per evaluation cycle to avoid
// redundant connections when multiple targets share a NUT server.
type UPSRepository interface {
	// GetJSON returns the JSON status for a NUT server.
	GetJSON(server *entity.NutServer) (string, error)
}
