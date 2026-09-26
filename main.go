package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type result struct {
	URL       string `json:"url"`
	Status    int    `json:"status"`
	LatencyMS int64  `json:"latency_ms"`
	Error     string `json:"error,omitempty"`
}

func readURLs(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var urls []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		urls = append(urls, line)
	}
	return urls, scanner.Err()
}

func check(client *http.Client, url string) result {
	started := time.Now()
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return result{URL: url, Error: err.Error()}
	}
	req.Header.Set("User-Agent", "pulsecheck/1.0")

	resp, err := client.Do(req)
	elapsed := time.Since(started).Milliseconds()
	if err != nil {
		return result{URL: url, LatencyMS: elapsed, Error: err.Error()}
	}
	defer resp.Body.Close()

	return result{URL: url, Status: resp.StatusCode, LatencyMS: elapsed}
}

func main() {
	workers := flag.Int("workers", 8, "cantidad de comprobaciones simultaneas")
	timeout := flag.Duration("timeout", 5*time.Second, "timeout por request")
	asJSON := flag.Bool("json", false, "mostrar salida JSON")
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Println("uso: pulsecheck [opciones] urls.txt")
		os.Exit(1)
	}
	if *workers < 1 {
		*workers = 1
	}

	urls, err := readURLs(flag.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, "no se pudo leer el archivo:", err)
		os.Exit(1)
	}
	if len(urls) == 0 {
		fmt.Println("no hay URLs para comprobar")
		return
	}

	client := &http.Client{Timeout: *timeout}
	jobs := make(chan string)
	out := make(chan result)
	var wg sync.WaitGroup

	for i := 0; i < *workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for url := range jobs {
				out <- check(client, url)
			}
		}()
	}

	go func() {
		for _, url := range urls {
			jobs <- url
		}
		close(jobs)
		wg.Wait()
		close(out)
	}()

	results := make([]result, 0, len(urls))
	for item := range out {
		results = append(results, item)
	}

	if *asJSON {
		data, _ := json.MarshalIndent(results, "", "  ")
		fmt.Println(string(data))
		return
	}

	ok := 0
	for _, item := range results {
		if item.Error != "" {
			fmt.Printf("%-45s ERROR  %dms  %s\n", item.URL, item.LatencyMS, item.Error)
			continue
		}
		if item.Status >= 200 && item.Status < 400 {
			ok++
		}
		fmt.Printf("%-45s %3d    %dms\n", item.URL, item.Status, item.LatencyMS)
	}
	fmt.Printf("\n%d/%d disponibles\n", ok, len(results))
}
