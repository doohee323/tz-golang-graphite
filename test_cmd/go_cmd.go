package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
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
	var urlstr = ""
	if 1 == 1 {
		urlstr = "http://core.local.xdn.com/1/stats/uptime_list?company_id=1&start_time=1464636372&end_time=1464722772&hc_id=" + strconv.Itoa(url)
	} else {
		var hcid = "0.0.1.4.1.8"
		var hcid2 = formatHcid(strconv.Itoa(url))
		var baseUrl = "http://173.243.129.12/render?target=summarize(averageSeries(dhc.stats.hcid.0.0.1.4.1.8.*.metrics.state.*),%221hour%22,%22avg%22)&target=summarize(averageSeries(dhc.stats.hcid.0.0.1.4.1.8.*.judge.state),%221hour%22,%22avg%22)&from=-1440min&until=-0min&format=json"
		urlstr = strings.Replace(baseUrl, hcid, hcid2, -1)
	}

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

func formatHcid(hcid string) string {
	maxlen := 6
	out := ""
	currlen := 0

	for _, r := range hcid {
		char := string(r)
		out = fmt.Sprintf("%s.%s", out, char)
		currlen++
	}

	out = strings.Trim(out, ".")

	//now lets see how much we need to pad it
	pad := maxlen - currlen
	for i := 1; i <= pad; i++ {
		out = fmt.Sprintf("%s.%s", "0", out)
	}
	log.Printf("shard: %s", out)
	return out
}

func _main() int {
	start := time.Now()
	r := new(big.Int)
	fmt.Println(r.Binomial(1000, 10))

	numCPUs := runtime.NumCPU()
	runtime.GOMAXPROCS(numCPUs)

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

	return count - 1
}

var all []int

func main() {

	programName := os.Args[0:1]

	if len(os.Args) < 2 {
		all = append(all, 1418, 1419, 2502, 2694, 2932, 2933, 2695)
	} else {
		firstArg := os.Args[1:2]
		//	secondArg := os.Args[2:3]
		allArgs := os.Args[1:]
		fmt.Println(programName, firstArg, allArgs)

		arry := strings.Split(firstArg[0], ",")
		for i := range arry {
			n, _ := strconv.Atoi(arry[i])
			all = append(all, n)
		}
	}
	fmt.Println(all)

	_main()
}
