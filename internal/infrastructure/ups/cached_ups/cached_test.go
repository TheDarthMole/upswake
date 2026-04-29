package cachedups

import (
	"errors"
	"sync/atomic"
	"testing"

	"github.com/TheDarthMole/UPSWake/internal/domain/entity"
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

func TestCachedRepository_DeduplicatesByHostPort(t *testing.T) {
	inner := &countingRepo{json: `[{"Name":"ups1"}]`}
	cached := NewCachedRepository(inner)

	server := &entity.NutServer{Host: "192.168.1.10", Port: 3493}

	// Call twice with the same server
	json1, err1 := cached.GetJSON(server)
	require.NoError(t, err1)

	json2, err2 := cached.GetJSON(server)
	require.NoError(t, err2)

	assert.Equal(t, json1, json2)
	assert.Equal(t, int32(1), inner.calls.Load(), "inner repo should only be called once")
}

func TestCachedRepository_DifferentServersCallInner(t *testing.T) {
	inner := &countingRepo{json: `[{"Name":"ups1"}]`}
	cached := NewCachedRepository(inner)

	server1 := &entity.NutServer{Host: "192.168.1.10", Port: 3493}
	server2 := &entity.NutServer{Host: "192.168.1.11", Port: 3493}

	_, _ = cached.GetJSON(server1)
	_, _ = cached.GetJSON(server2)

	assert.Equal(t, int32(2), inner.calls.Load(), "different servers should each call inner")
}

func TestCachedRepository_CachesErrors(t *testing.T) {
	expectedErr := errors.New("connection refused")
	inner := &countingRepo{err: expectedErr}
	cached := NewCachedRepository(inner)

	server := &entity.NutServer{Host: "192.168.1.10", Port: 3493}

	_, err1 := cached.GetJSON(server)
	_, err2 := cached.GetJSON(server)

	assert.ErrorIs(t, err1, expectedErr)
	assert.ErrorIs(t, err2, expectedErr)
	assert.Equal(t, int32(1), inner.calls.Load(), "error should be cached too")
}

func TestCachedRepository_ResetClearsCache(t *testing.T) {
	inner := &countingRepo{json: `[{"Name":"ups1"}]`}
	cached := NewCachedRepository(inner)

	server := &entity.NutServer{Host: "192.168.1.10", Port: 3493}

	_, _ = cached.GetJSON(server)
	cached.Reset()
	_, _ = cached.GetJSON(server)

	assert.Equal(t, int32(2), inner.calls.Load(), "reset should force a fresh call")
}
