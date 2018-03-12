package main

import (
	"fmt"
	"sync"
)

type Fetcher interface {
	// Fetch returns the body of URL and
	// a slice of URLs found on that page.
	Fetch(url string) (body string, urls []string, err error)
}

type UrlCache struct {
	sync.Mutex
	urls map[string]bool
}

func (uc *UrlCache) Add(url string) {
	uc.Lock()
	defer uc.Unlock()
	if _, exists := uc.urls[url]; !exists {
		uc.urls[url] = true
	}
}

func (uc *UrlCache) IsAdded(url string) bool {
	uc.Lock()
	defer uc.Unlock()
	if _, exists := uc.urls[url]; exists {
		return true
	}
	return false
}

func NewCache() *UrlCache {
	return &UrlCache{
		sync.Mutex{},
		map[string]bool{},
	}
}

// Crawl uses fetcher to recursively crawl
// pages starting with url, to a maximum of depth.
func Crawl(url string, depth int, fetcher Fetcher, urlc *UrlCache, wg sync.WaitGroup) {
	// TODO: Fetch URLs in parallel.
	// TODO: Don't fetch the same URL twice.
	// This implementation doesn't do either:
	if depth <= 0 {
		wg.Done()
		return
	}
	if !urlc.IsAdded(url) {
		// if 0 != 1 {
		body, urls, err := fetcher.Fetch(url)
		if err != nil {
			fmt.Println(err)
			return
		}
		// urlc.Add(url)
		fmt.Printf("found: %s %q\n", url, body)
		fmt.Println(urls)
		wg.Add(len(urls))
		for _, u := range urls {
			func(_u string) {
				go Crawl(_u, depth-1, fetcher, urlc, wg)
			}(u)
		}
	}
	wg.Done()
	return
}

func main() {
	c := NewCache()
	wg := sync.WaitGroup{}
	wg.Add(1)
	Crawl("https://golang.org/", 4, fetcher, c, wg)
	wg.Wait()
}

// fakeFetcher is Fetcher that returns canned results.
type fakeFetcher map[string]*fakeResult

type fakeResult struct {
	body string
	urls []string
}

func (f fakeFetcher) Fetch(url string) (string, []string, error) {
	if res, ok := f[url]; ok {
		return res.body, res.urls, nil
	}
	return "", nil, fmt.Errorf("not found: %s", url)
}

// fetcher is a populated fakeFetcher.
var fetcher = fakeFetcher{
	"https://golang.org/": &fakeResult{
		"The Go Programming Language",
		[]string{
			"https://golang.org/pkg/",
			"https://golang.org/cmd/",
		},
	},
	"https://golang.org/pkg/": &fakeResult{
		"Packages",
		[]string{
			"https://golang.org/",
			"https://golang.org/cmd/",
			"https://golang.org/pkg/fmt/",
			"https://golang.org/pkg/os/",
		},
	},
	"https://golang.org/pkg/fmt/": &fakeResult{
		"Package fmt",
		[]string{
			"https://golang.org/",
			"https://golang.org/pkg/",
		},
	},
	"https://golang.org/pkg/os/": &fakeResult{
		"Package os",
		[]string{
			"https://golang.org/",
			"https://golang.org/pkg/",
		},
	},
}
