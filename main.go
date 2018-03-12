package main

import (
	"fmt"
	"runtime"
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

func (urlc *UrlCache) IsVisited(url string) bool {
	urlc.Lock()
	defer urlc.Unlock()
	_, exists := urlc.urls[url]
	return exists == true
}

func (urlc *UrlCache) VisitUrl(url string) {
	urlc.Lock()
	defer urlc.Unlock()
	urlc.urls[url] = true
}

// Crawl uses fetcher to recursively crawl
// pages starting with url, to a maximum of depth.
func Crawl(url string, depth int, fetcher Fetcher, urlc *UrlCache, wg *sync.WaitGroup, name int) {
	defer wg.Done()
	if depth <= 0 {
		return
	}
	if !urlc.IsVisited(url) {
		body, urls, err := fetcher.Fetch(url)
		if err != nil {
			fmt.Println(err)
			return
		}
		urlc.VisitUrl(url)
		fmt.Printf("name: %v, urls: %v\n", name, urls)
		fmt.Printf("name: %v, found: %s %q\n", name, url, body)
		fmt.Printf("num routines: %v\n", runtime.NumGoroutine())
		for i, u := range urls {
			wg.Add(1)
			go func(_u string, _i int) {
				Crawl(_u, depth-1, fetcher, urlc, wg, name+i+20)
				// bind parameters to this closure
				// otherwise, Crawl call is bound to the u reference, and will likely be end of loop value
			}(u, i)
		}
	} else {
		fmt.Printf("name: %v ignoring existing url: %v\n", name, url)
	}
}

func main() {
	var wg sync.WaitGroup
	urlc := &UrlCache{
		urls: map[string]bool{},
	}
	wg.Add(1)
	name := 0
	go Crawl("https://golang.org/", 4, fetcher, urlc, &wg, name)
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
