package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"runtime"
	"strconv"
	"sync"
	"testing"
	"time"
)

var fetched = struct {
	m map[int]error
	sync.Mutex
}{m: make(map[int]error)}

type result struct {
	url     int
	content string
}

func fetch(url int) (string, error) {
	if url == -1 {
		return "", nil
	}
	var urlstr = "http://core.local.xdn.com/1/stats/uptime_list?company_id=1&start_time=1464636372&end_time=1464722772&hc_id=" + strconv.Itoa(url)
	log.Printf("==== %s", urlstr)
	res, err := http.Get(urlstr)
	if err != nil {
		log.Println(err)
		return "", err
	}
	defer res.Body.Close()
	doc, err := ioutil.ReadAll(res.Body)
	return string(doc), nil
}

func parseFollowing(url int, doc string, urls chan int) <-chan string {
	content := make(chan string)

	go func() {
		content <- doc
		chk := false
		val := -1
		fetched.Lock()
		for n := range all {
			if _, ok := fetched.m[all[n]]; !ok {
				chk = true
				val = all[n]
				break
			}
		}
		if chk == false {

		}
		urls <- val
		fetched.Unlock()
	}()
	return content
}

func crawl(url int, urls chan int, c chan<- result) {
	fetched.Lock()
	if _, ok := fetched.m[url]; ok {
		fetched.Unlock()
		return
	}
	fetched.Unlock()

	doc, err := fetch(url)
	if err != nil {
		go func(u int) {
			urls <- u
		}(url)
		return
	}

	fetched.Lock()
	fetched.m[url] = err
	fetched.Unlock()

	doc = <-parseFollowing(url, doc, urls)
	//	log.Printf("---- %d %s", url, doc)
	c <- result{url, doc}
}

func worker(done <-chan struct{}, urls chan int, c chan<- result) {
	for url := range urls {
		select {
		case <-done:
			return
		default:
			crawl(url, urls, c)
		}
	}
}

var all []int

func TestMap(t *testing.T) {
	start := time.Now()
	r := new(big.Int)
	fmt.Println(r.Binomial(1000, 10))

	numCPUs := runtime.NumCPU()
	runtime.GOMAXPROCS(numCPUs)

	all = append(all, 1418, 1419, 2502, 2694, 2932, 2933, 2695)
	urls := make(chan int)
	done := make(chan struct{})
	c := make(chan result)

	var wg sync.WaitGroup
	const numWorkers = 10
	wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go func() {
			worker(done, urls, c)
			wg.Done()
		}()
	}

	go func() {
		wg.Wait()
		close(c)
	}()

	urls <- -1

	count := 0
	for r := range c {
		log.Printf("==== %d %s", r.url, r.content)

		count++
		if count > len(all) {
			close(done)
			break
		}
	}

	elapsed := time.Since(start)
	log.Printf("It took %s", elapsed)
}
