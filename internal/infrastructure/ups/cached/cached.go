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
	mu    sync.RWMutex
}

type cachedResult struct {
	cachedAt time.Time
	err      error
	json     string
}

func NewCachedRepository(inner repository.UPSRepository, ttl time.Duration) *CachedRepository {
	return &CachedRepository{
		inner: inner,
		cache: make(map[string]cachedResult),
		ttl:   ttl,
	}
}

func (r *CachedRepository) GetJSON(server *entity.NutServer) (string, error) {
	key := fmt.Sprintf("%s:%d", server.Host, server.Port)

	r.mu.RLock()
	if result, ok := r.cache[key]; ok && time.Since(result.cachedAt) < r.ttl {
		r.mu.RUnlock()
		return result.json, result.err
	}
	r.mu.RUnlock()

	json, err := r.inner.GetJSON(server)
	if err == nil { // Save to cache if there was no error
		r.mu.Lock()
		r.cache[key] = cachedResult{
			json:     json,
			cachedAt: time.Now(),
		}
		r.mu.Unlock()
	}

	return json, err
}

// Reset clears the cache so the next GetJSON call fetches fresh data.
func (r *CachedRepository) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	clear(r.cache)
}
