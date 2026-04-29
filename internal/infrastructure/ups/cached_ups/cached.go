package cachedups

import (
	"fmt"
	"sync"

	"github.com/TheDarthMole/UPSWake/internal/domain/entity"
	"github.com/TheDarthMole/UPSWake/internal/domain/repository"
)

// CachedRepository wraps another UPSRepository and caches results
// keyed by host:port. Call Reset() between evaluation cycles to
// force fresh data on the next tick.
type CachedRepository struct {
	inner repository.UPSRepository
	cache map[string]cachedResult
	mu    sync.Mutex
}

type cachedResult struct {
	err  error
	json string
}

func NewCachedRepository(inner repository.UPSRepository) *CachedRepository {
	return &CachedRepository{
		inner: inner,
		cache: make(map[string]cachedResult),
	}
}

func (r *CachedRepository) GetJSON(server *entity.NutServer) (string, error) {
	key := fmt.Sprintf("%s:%d", server.Host, server.Port)

	r.mu.Lock()
	defer r.mu.Unlock()

	if result, ok := r.cache[key]; ok {
		return result.json, result.err
	}

	json, err := r.inner.GetJSON(server)
	r.cache[key] = cachedResult{json: json, err: err}
	return json, err
}

// Reset clears the cache so the next GetJSON call fetches fresh data.
func (r *CachedRepository) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	clear(r.cache)
}
