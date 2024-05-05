package main

import "fmt"

type WebsiteChecker func(string) bool

func CheckWebsitesSlow(wc WebsiteChecker, urls []string) map[string]bool {
	results := make(map[string]bool)

	// Does not take advantage of concurrency and is linear
	for _, url := range urls {
		results[url] = wc(url)
	}

	return results
}

type result struct {
	string
	bool
}

func CheckWebsites(wc WebsiteChecker, urls []string) map[string]bool {
	results := make(map[string]bool)
	resultChannel := make(chan result)

	// This part is running concurrently
	for _, url := range urls {
		go func(u string) {
			// Send statement
			resultChannel <- result{u, wc(u)}
		}(url)
	}

	fmt.Println("Waiting for results...")

	// This part is running linearly
	for i := 0; i < len(urls); i++ {
		// Receive statement
		r := <-resultChannel
		fmt.Println("Received result for", r.string)
		results[r.string] = r.bool
	}

	return results
}
