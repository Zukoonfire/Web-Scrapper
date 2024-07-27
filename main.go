package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Job struct {
	URL string
}
type Result struct {
	URL  string
	Data string
}

var (
	jobs    = make(chan Job, 10)
	results = make(chan Result, 10)
)

func main() {
	urls := []string{
		"https://go.dev/doc/effective_go#concurrency",
		"https://golangbot.com/buffered-channels-worker-pools/",
		"https://chatgpt.com/c/33681582-e751-42e5-a5f9-98a6f5cdd7f0",
	}
	startTime := time.Now()
	noOfworker := 2
	
	go allocateJobs(urls)
	done := make(chan bool)
	go resultsContent(done)
	createWorkerPool(noOfworker)
	<-done
	endTime := time.Now()
	diff := endTime.Sub(startTime)
	fmt.Println("Total Time taken is  ", diff.Seconds(), "Seconds")
}
func createWorkerPool(noOfWorkers int) {
	var wg sync.WaitGroup
	for i := 0; i < noOfWorkers; i++ {
		wg.Add(1)
		go worker(&wg)
	}
	wg.Wait()
	close(results)
}
func allocateJobs(urls []string) {
	for _, url := range urls {
		jobs <- Job{
			URL: url,
		}
	}
	close(jobs)
}
func worker(wg *sync.WaitGroup) {
	defer wg.Done()
	for job := range jobs {
		var res Result
		data, err := fetchData(job.URL)
		if err != nil {
			log.Printf("Error Fetching Data from %s:%v", job.URL, err)
			continue
		}
		res = Result{
			Data: data,
			URL:  job.URL,
		}
		results <- res
	}

}
func fetchData(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", err
	}

	//Extracting the title
	title := doc.Find("title").Text()

	//Extracting the first paragraph
	var paragraphs []string
	doc.Find("p").Each(func(i int, s *goquery.Selection) {
		paragraphs = append(paragraphs, s.Text())
	})
	//Combining all paragraphs into a single thing
	allParagraphs := strings.Join(paragraphs, "\n\n")

	//format the extracted data
	data := fmt.Sprintf("Title: %s\n\nFirst Paragraph:%s\n", title, allParagraphs)
	return data, nil
}
func resultsContent(done chan bool) {
	for resValue := range results {
		fmt.Printf("URL: %s\nContent:\n%s\n\n", resValue.URL, resValue.Data)
	}
	done <- true
}
