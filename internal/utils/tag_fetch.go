package utils

import (
	"context"
	"sync"
)

// defaultTagFetchConcurrency is the maximum number of parallel tag-fetch API calls.
const defaultTagFetchConcurrency = 20

type tagFetchResult struct {
	uid  string
	tags map[string]string
	err  error
}

// FetchTagsConcurrently fetches resource tags for multiple UIDs in parallel,
// bounding concurrency to defaultTagFetchConcurrency outstanding requests.
// It returns two maps: uid→tags for successful fetches and uid→error for failures.
func FetchTagsConcurrently(
	ctx context.Context,
	uids []string,
	fetch func(ctx context.Context, uid string) (map[string]string, error),
) (map[string]map[string]string, map[string]error) {
	if len(uids) == 0 {
		return nil, nil
	}

	results := make(chan tagFetchResult, len(uids))
	sem := make(chan struct{}, defaultTagFetchConcurrency)

	var wg sync.WaitGroup
	for _, uid := range uids {
		uid := uid
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			tags, err := fetch(ctx, uid)
			results <- tagFetchResult{uid: uid, tags: tags, err: err}
		}()
	}
	go func() {
		wg.Wait()
		close(results)
	}()

	tagMap := make(map[string]map[string]string, len(uids))
	errMap := make(map[string]error)
	for r := range results {
		if r.err != nil {
			errMap[r.uid] = r.err
		} else {
			tagMap[r.uid] = r.tags
		}
	}
	return tagMap, errMap
}
