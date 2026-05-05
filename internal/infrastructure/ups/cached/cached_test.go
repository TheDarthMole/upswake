package cachedups

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/TheDarthMole/UPSWake/internal/domain/entity"
	"github.com/TheDarthMole/UPSWake/internal/domain/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type countingRepo struct {
	err   error
	json  string
	calls atomic.Int32
}

func (r *countingRepo) GetJSON(_ *entity.NutServer) (string, error) {
	r.calls.Add(1)
	return r.json, r.err
}

func TestNewCachedRepository(t *testing.T) {
	type args struct {
		inner repository.UPSRepository
		ttl   time.Duration
	}
	tests := []struct {
		want *CachedRepository
		args args
		name string
	}{
		{
			name: "5 second cache",
			args: args{
				inner: &countingRepo{},
				ttl:   5 * time.Second,
			},
			want: &CachedRepository{
				inner: &countingRepo{},
				cache: map[string]cachedResult{},
				ttl:   5 * time.Second,
				mu:    sync.Mutex{},
			},
		},
		{
			name: "1 minute cache",
			args: args{
				inner: &countingRepo{},
				ttl:   1 * time.Minute,
			},
			want: &CachedRepository{
				inner: &countingRepo{},
				cache: map[string]cachedResult{},
				ttl:   1 * time.Minute,
				mu:    sync.Mutex{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, NewCachedRepository(tt.args.inner, tt.args.ttl), "NewCachedRepository(%v, %v)", tt.args.inner, tt.args.ttl)
		})
	}
}

func TestCachedRepository_GetJSON(t *testing.T) {
	t.Run("two calls one cached response", func(t *testing.T) {
		inner := &countingRepo{json: `[{"Name":"ups1"}]`}
		cached := NewCachedRepository(inner, 5*time.Second)

		server := &entity.NutServer{Host: "192.168.1.10", Port: 3493}

		// Call twice with the same server
		json1, err1 := cached.GetJSON(server)
		require.NoError(t, err1)

		json2, err2 := cached.GetJSON(server)
		require.NoError(t, err2)

		assert.Equal(t, json1, json2)
		assert.Equal(t, err1, err2)
		assert.Equal(t, int32(1), inner.calls.Load(), "inner repo should only be called once")
	})

	t.Run("two servers two unique inner calls", func(t *testing.T) {
		inner := &countingRepo{json: `[{"Name":"ups1"}]`}
		cached := NewCachedRepository(inner, 5*time.Second)

		server1 := &entity.NutServer{Host: "192.168.1.10", Port: 3493}
		server2 := &entity.NutServer{Host: "192.168.1.11", Port: 3493}

		_, _ = cached.GetJSON(server1)
		_, _ = cached.GetJSON(server2)

		assert.Equal(t, int32(2), inner.calls.Load(), "different servers should each call inner")
	})

	t.Run("errors skip cache", func(t *testing.T) {
		expectedErr := errors.New("connection refused")
		inner := &countingRepo{err: expectedErr}
		cached := NewCachedRepository(inner, 5*time.Second)

		server := &entity.NutServer{Host: "192.168.1.10", Port: 3493}

		_, err1 := cached.GetJSON(server)
		_, err2 := cached.GetJSON(server)

		assert.ErrorIs(t, err1, expectedErr)
		assert.ErrorIs(t, err2, expectedErr)
		assert.Equal(t, int32(2), inner.calls.Load(), "error should not be saved to cache")
	})

	t.Run("reset clears cache", func(t *testing.T) {
		inner := &countingRepo{json: `[{"Name":"ups1"}]`}
		cached := NewCachedRepository(inner, 5*time.Second)

		server := &entity.NutServer{Host: "192.168.1.10", Port: 3493}

		_, _ = cached.GetJSON(server)
		cached.Reset()
		_, _ = cached.GetJSON(server)

		assert.Equal(t, int32(2), inner.calls.Load(), "reset should force a fresh call")
	})

	t.Run("cache expires after TTL", func(t *testing.T) {
		inner := &countingRepo{json: `[{"Name":"ups1"}]`}
		cached := NewCachedRepository(inner, 50*time.Millisecond)

		server := &entity.NutServer{Host: "192.168.1.10", Port: 3493}

		// Call twice with the same server
		json1, err1 := cached.GetJSON(server)
		require.NoError(t, err1)

		time.Sleep(100 * time.Millisecond) // wait for cache to expire

		json2, err2 := cached.GetJSON(server)
		require.NoError(t, err2)

		assert.Equal(t, json1, json2)
		assert.Equal(t, err1, err2)
		assert.Equal(t, int32(2), inner.calls.Load(), "inner repo should be called twice")
	})
}
