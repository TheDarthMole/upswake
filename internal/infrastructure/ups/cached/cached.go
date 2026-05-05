package cachedups

import (
	"fmt"
	"sync"
	"time"

	"github.com/TheDarthMole/UPSWake/internal/domain/entity"
	"github.com/TheDarthMole/UPSWake/internal/domain/repository"
)

// CachedRepository wraps another UPSRepository and caches results
// keyed by host:port. Call Reset() between evaluation cycles to
// force fresh data on the next tick.
type CachedRepository struct {
	inner repository.UPSRepository
	cache map[string]cachedResult
	ttl   time.Duration
	mu    sync.Mutex
}

type cachedResult struct {
	cachedAt time.Time
	err      error
	json     string
}

// NewCachedRepository creates a CachedRepository that wraps the provided UPSRepository and caches GetJSON results keyed by host:port.
// It initialises an empty cache and uses ttl as the maximum age for cached entries.
func NewCachedRepository(inner repository.UPSRepository, ttl time.Duration) *CachedRepository {
	return &CachedRepository{
		inner: inner,
		cache: make(map[string]cachedResult),
		ttl:   ttl,
	}
}

func (r *CachedRepository) GetJSON(server *entity.NutServer) (string, error) {
	key := fmt.Sprintf("%s:%d", server.Host, server.Port)

	r.mu.Lock()
	defer r.mu.Unlock()

	if result, ok := r.cache[key]; ok && time.Since(result.cachedAt) < r.ttl {
		return result.json, result.err
	}

	json, err := r.inner.GetJSON(server)
	r.cache[key] = cachedResult{
		json:     json,
		err:      err,
		cachedAt: time.Now(),
	}
	return json, err
}

// Reset clears the cache so the next GetJSON call fetches fresh data.
func (r *CachedRepository) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	clear(r.cache)
}
