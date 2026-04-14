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

// --- ApplyTagFilter tests ---

type testResource struct {
	uid  string
	tags map[string]string
}

func makeTagFetch(resources []testResource) func(context.Context, string) (map[string]string, error) {
	m := make(map[string]map[string]string, len(resources))
	for _, r := range resources {
		m[r.uid] = r.tags
	}
	return func(_ context.Context, uid string) (map[string]string, error) {
		return m[uid], nil
	}
}

func TestApplyTagFilter_NoLimit_AllMatch(t *testing.T) {
	resources := []testResource{
		{uid: "a", tags: map[string]string{"env": "prod"}},
		{uid: "b", tags: map[string]string{"env": "prod"}},
	}
	got, errs := ApplyTagFilter(context.Background(), resources,
		func(r testResource) string { return r.uid },
		makeTagFetch(resources),
		[]string{"env=prod"}, 0,
	)
	assert.Len(t, got, 2)
	assert.Empty(t, errs)
}

func TestApplyTagFilter_NoLimit_SomeMatch(t *testing.T) {
	resources := []testResource{
		{uid: "a", tags: map[string]string{"env": "prod"}},
		{uid: "b", tags: map[string]string{"env": "staging"}},
		{uid: "c", tags: map[string]string{"env": "prod"}},
	}
	got, errs := ApplyTagFilter(context.Background(), resources,
		func(r testResource) string { return r.uid },
		makeTagFetch(resources),
		[]string{"env=prod"}, 0,
	)
	require.Len(t, got, 2)
	assert.Empty(t, errs)
	assert.Equal(t, "a", got[0].uid)
	assert.Equal(t, "c", got[1].uid)
}

func TestApplyTagFilter_WithLimit_StopsEarly(t *testing.T) {
	var callCount int
	resources := []testResource{
		{uid: "a", tags: map[string]string{"env": "prod"}},
		{uid: "b", tags: map[string]string{"env": "prod"}},
		{uid: "c", tags: map[string]string{"env": "prod"}},
	}
	fetch := func(_ context.Context, uid string) (map[string]string, error) {
		callCount++
		for _, r := range resources {
			if r.uid == uid {
				return r.tags, nil
			}
		}
		return nil, nil
	}
	got, errs := ApplyTagFilter(context.Background(), resources,
		func(r testResource) string { return r.uid },
		fetch,
		[]string{"env=prod"}, 2,
	)
	assert.Len(t, got, 2)
	assert.Empty(t, errs)
	// With limit=2, should stop after 2 matches — only 2 API calls needed.
	assert.Equal(t, 2, callCount)
}

func TestApplyTagFilter_FetchError_Excluded(t *testing.T) {
	resources := []testResource{{uid: "ok"}, {uid: "err"}}
	fetch := func(_ context.Context, uid string) (map[string]string, error) {
		if uid == "err" {
			return nil, errors.New("fetch failed")
		}
		return map[string]string{"env": "prod"}, nil
	}
	got, errs := ApplyTagFilter(context.Background(), resources,
		func(r testResource) string { return r.uid },
		fetch,
		[]string{"env=prod"}, 0,
	)
	require.Len(t, got, 1)
	assert.Equal(t, "ok", got[0].uid)
	assert.Contains(t, errs, "err")
	assert.EqualError(t, errs["err"], "fetch failed")
}
