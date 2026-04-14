package utils

import (
	"context"
	"sync"
)

// ApplyTagFilter filters resources by tag, choosing the optimal fetch strategy:
//   - limit > 0: sequential fetch with early stopping (avoids unnecessary API calls
//     once enough matches are found)
//   - limit == 0: parallel fetch via FetchTagsConcurrently (minimises wall-clock time
//     when all results are needed)
//
// Resources whose tags could not be fetched are excluded from the results; their UIDs
// and errors are returned in the second map so the caller can report them after any
// spinner has been stopped (avoiding interleaved terminal output).
func ApplyTagFilter[T any](
	ctx context.Context,
	resources []T,
	uidFunc func(T) string,
	fetch func(context.Context, string) (map[string]string, error),
	tagFilters []string,
	limit int,
) ([]T, map[string]error) {
	if limit > 0 {
		result := make([]T, 0, limit)
		var fetchErrs map[string]error
		for _, r := range resources {
			uid := uidFunc(r)
			tags, err := fetch(ctx, uid)
			if err != nil {
				if fetchErrs == nil {
					fetchErrs = make(map[string]error)
				}
				fetchErrs[uid] = err
				continue
			}
			if MatchesTagFilters(tags, tagFilters) {
				result = append(result, r)
				if len(result) >= limit {
					break
				}
			}
		}
		return result, fetchErrs
	}

	// Unlimited: fetch all tags in parallel for minimal wall-clock time.
	uids := make([]string, len(resources))
	for i, r := range resources {
		uids[i] = uidFunc(r)
	}
	allTags, fetchErrs := FetchTagsConcurrently(ctx, uids, fetch)
	result := make([]T, 0, len(resources))
	for _, r := range resources {
		uid := uidFunc(r)
		if _, failed := fetchErrs[uid]; failed {
			continue
		}
		if MatchesTagFilters(allTags[uid], tagFilters) {
			result = append(result, r)
		}
	}
	return result, fetchErrs
}

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
