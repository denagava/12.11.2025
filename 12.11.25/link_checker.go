package main

import (
	"context"
	"net/http"
	"sync"
	"time"
)

type LinkChecker struct {
	client *http.Client
}

func NewLinkChecker() *LinkChecker {
	return &LinkChecker{
		client: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
			},
		},
	}
}

func (lc *LinkChecker) CheckLink(ctx context.Context, url string) string {
	fullURL := url
	if len(url) > 0 && url[0] != '#' {
		if len(url) < 8 || (url[:7] != "http://" && url[:8] != "https://") {
			fullURL = "http://" + url
		}
	}
	req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if err != nil {
		return "not available"
	}
	resp, err := lc.client.Do(req)
	if err != nil {
		return "not available"
	}
	defer resp.Body.Close()
	if resp.StatusCode < 400 {
		return "available"
	}
	return "not available"
}

func (lc *LinkChecker) CheckLinks(ctx context.Context, urls []string) map[string]string {
	var wg sync.WaitGroup
	results := make(map[string]string)
	var mu sync.Mutex
	for _, url := range urls {
		wg.Add(1)
		go func(u string) {
			defer wg.Done()
			select {
			case <-ctx.Done():
				mu.Lock()
				results[u] = "not available"
				mu.Unlock()
				return
			default:
				status := lc.CheckLink(ctx, u)
				mu.Lock()
				results[u] = status
				mu.Unlock()
			}
		}(url)
	}
	wg.Wait()
	return results
}
