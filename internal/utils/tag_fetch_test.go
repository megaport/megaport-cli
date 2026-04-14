package utils

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFetchTagsConcurrently_Empty(t *testing.T) {
	called := false
	tagMap, errMap := FetchTagsConcurrently(context.Background(), nil, func(_ context.Context, _ string) (map[string]string, error) {
		called = true
		return nil, nil
	})
	assert.Nil(t, tagMap)
	assert.Nil(t, errMap)
	assert.False(t, called)
}

func TestFetchTagsConcurrently_AllSuccess(t *testing.T) {
	data := map[string]map[string]string{
		"uid-1": {"env": "prod"},
		"uid-2": {"env": "staging"},
		"uid-3": {},
	}
	uids := []string{"uid-1", "uid-2", "uid-3"}

	tagMap, errMap := FetchTagsConcurrently(context.Background(), uids, func(_ context.Context, uid string) (map[string]string, error) {
		return data[uid], nil
	})

	require.Empty(t, errMap)
	assert.Equal(t, data["uid-1"], tagMap["uid-1"])
	assert.Equal(t, data["uid-2"], tagMap["uid-2"])
	assert.Equal(t, data["uid-3"], tagMap["uid-3"])
}

func TestFetchTagsConcurrently_PartialErrors(t *testing.T) {
	fetchErr := errors.New("not found")
	uids := []string{"uid-ok", "uid-err"}

	tagMap, errMap := FetchTagsConcurrently(context.Background(), uids, func(_ context.Context, uid string) (map[string]string, error) {
		if uid == "uid-err" {
			return nil, fetchErr
		}
		return map[string]string{"k": "v"}, nil
	})

	assert.Equal(t, map[string]string{"k": "v"}, tagMap["uid-ok"])
	assert.Equal(t, fetchErr, errMap["uid-err"])
	assert.NotContains(t, tagMap, "uid-err")
}

func TestFetchTagsConcurrently_BoundedConcurrency(t *testing.T) {
	// Verify that no more than defaultTagFetchConcurrency goroutines fetch at once.
	var inflight atomic.Int64
	var maxSeen atomic.Int64

	uids := make([]string, defaultTagFetchConcurrency*3)
	for i := range uids {
		uids[i] = "uid"
	}

	FetchTagsConcurrently(context.Background(), uids, func(_ context.Context, _ string) (map[string]string, error) {
		cur := inflight.Add(1)
		for {
			m := maxSeen.Load()
			if cur <= m || maxSeen.CompareAndSwap(m, cur) {
				break
			}
		}
		inflight.Add(-1)
		return nil, nil
	})

	assert.LessOrEqual(t, maxSeen.Load(), int64(defaultTagFetchConcurrency))
}

func TestFetchTagsConcurrently_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // pre-cancel

	uids := []string{"uid-1", "uid-2"}
	_, errMap := FetchTagsConcurrently(ctx, uids, func(ctx context.Context, uid string) (map[string]string, error) {
		return nil, ctx.Err()
	})

	assert.Len(t, errMap, 2)
}
