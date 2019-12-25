package main

import (
	"fmt"
	"os"
	"bufio"
	"sync"
	"flag"
	"net/http"
)


const NUM_WORKERS = 3

// Result Struct to hold URL and status code for processing
// You could just pass the response off to the processor, but
// then I wouldn't get to learn about structs :)
type Result struct {
	URL	string `json:"url"`
	StatusCode int `json:"statusCode"`
}

func loadWords(path string) <-chan string {
	words := make(chan string)

	f, _ := os.Open(path)
	
	go func() {
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			words <- scanner.Text()	
		}
		close(words)
	}()

	return words
}

func processWords(words <-chan string, host string, wg *sync.WaitGroup) <-chan Result {
	results := make(chan Result)

	go func() {
		for i := 0; i < NUM_WORKERS; i++ {
			wg.Add(1)
			go webRequest(words, host, results, wg)
		}

		wg.Wait()
		close(results)
	}()

	return results
}

func webRequest(words <-chan string, host string, results chan<- Result, wg *sync.WaitGroup) {
	defer wg.Done()

	for word := range words {
		url := host + word
		resp, err := http.Get(url)
		if err != nil {
			fmt.Println("An error has occured: ", err)
		}
		results <- Result{URL: url, StatusCode: resp.StatusCode}
	}
}

func parseResults(results <-chan Result, wg *sync.WaitGroup) {
	sum := 0
	for result := range results {
		if result.StatusCode == 200 || result.StatusCode == 403 {
			sum++
			fmt.Printf("%v\n", result.URL)
		}
	}

	fmt.Printf("Found %v hidden directories\n", sum)
}

func main() {
	var wg sync.WaitGroup

	hostPtr := flag.String("host", "https://sans.org/", "The hostname to scan")
	wordlistPtr := flag.String("wordlist", "./wordlist.txt", "Path to the wordlist file (newline delimited)")

	flag.Parse()

	fmt.Println("Scanning " + *hostPtr)

	words := loadWords(*wordlistPtr)
	results := processWords(words, *hostPtr, &wg)

	parseResults(results, &wg)
}
