package utils

import (
	"context"
	"errors"
	"fmt"
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
	// Use exactly defaultTagFetchConcurrency UIDs so each worker handles one UID.
	// This ensures a single barrier cycle with no multi-batch re-send to ready.
	var inflight atomic.Int64
	var maxSeen atomic.Int64

	// ready is signalled once per inflight fetch so the test can wait for full saturation.
	ready := make(chan struct{}, defaultTagFetchConcurrency)
	// release is closed to unblock all workers at once.
	release := make(chan struct{})

	uids := make([]string, defaultTagFetchConcurrency)
	for i := range uids {
		uids[i] = fmt.Sprintf("uid-%d", i)
	}

	done := make(chan struct{})
	go func() {
		FetchTagsConcurrently(context.Background(), uids, func(_ context.Context, _ string) (map[string]string, error) {
			cur := inflight.Add(1)
			for {
				m := maxSeen.Load()
				if cur <= m || maxSeen.CompareAndSwap(m, cur) {
					break
				}
			}
			ready <- struct{}{} // signal: this fetch is inflight
			<-release           // block until test observes full saturation
			inflight.Add(-1)
			return nil, nil
		})
		close(done)
	}()

	// Wait for all workers to be simultaneously inflight.
	for i := 0; i < defaultTagFetchConcurrency; i++ {
		<-ready
	}
	// All workers are blocked: maxSeen reflects true simultaneous inflight count.
	assert.Equal(t, int64(defaultTagFetchConcurrency), maxSeen.Load())

	close(release) // unblock workers so the function can complete
	<-done
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
