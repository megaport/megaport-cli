package utils

import (
	"context"
	"sync"
)

// defaultTagFetchConcurrency is the size of the fixed worker pool used when fetching tags.
const defaultTagFetchConcurrency = 20

type tagFetchResult struct {
	uid  string
	tags map[string]string
	err  error
}

// FetchTagsConcurrently fetches resource tags for multiple UIDs using a fixed worker pool
// capped at defaultTagFetchConcurrency, so neither goroutine count nor channel buffer
// grows with the number of UIDs.
// It returns two maps: uid→tags for successful fetches and uid→error for failures.
func FetchTagsConcurrently(
	ctx context.Context,
	uids []string,
	fetch func(ctx context.Context, uid string) (map[string]string, error),
) (map[string]map[string]string, map[string]error) {
	if len(uids) == 0 {
		return nil, nil
	}

	workerCount := defaultTagFetchConcurrency
	if len(uids) < workerCount {
		workerCount = len(uids)
	}

	jobs := make(chan string, workerCount)
	results := make(chan tagFetchResult, workerCount)

	var wg sync.WaitGroup
	for range workerCount {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for uid := range jobs {
				tags, err := fetch(ctx, uid)
				results <- tagFetchResult{uid: uid, tags: tags, err: err}
			}
		}()
	}

	go func() {
		for _, uid := range uids {
			jobs <- uid
		}
		close(jobs)
	}()

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
