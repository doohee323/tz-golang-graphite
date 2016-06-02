package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/gorilla/pat"
	logging "github.com/op/go-logging"
	"io/ioutil"
	"math/big"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

//this is log file
var (
	HCIDS       []int
	STYPE       string
	LOGFILE     *os.File
	CORE_URL                  = "http://core.local.xdn.com/1/stats/uptime_list?company_id=1&start_time=1464636372&end_time=1464722772&hc_id="
	GRAPHTE_URL               = "http://173.243.129.129/render?target=summarize(averageSeries(dhc.stats.hcid.0.0.1.4.1.8.*.metrics.state.*),%221hour%22,%22avg%22)&target=summarize(averageSeries(dhc.stats.hcid.0.0.1.4.1.8.*.judge.state),%221hour%22,%22avg%22)&from=-1440min&until=-0min&format=json"
	LOGPATH                   = "/Users/dhong/tmp/test.log"
	LOGFMT                    = "%{color}%{time:15:04:05.000000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}"
	LOGFORMAT                 = logging.MustStringFormatter(LOGFMT)
	LOG                       = logging.MustGetLogger("logfile")
	GLOGLEVEL   logging.Level = logging.DEBUG
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
	if STYPE == "core" {
		urlstr = CORE_URL + strconv.Itoa(url)
	} else {
		var hcid = "0.0.1.4.1.8"
		var hcid2 = formatHcid(strconv.Itoa(url))
		var baseUrl = GRAPHTE_URL
		urlstr = strings.Replace(baseUrl, hcid, hcid2, -1)
	}

	LOG.Debug("==== %s", urlstr)
	res, err := http.Get(urlstr)
	if err != nil {
		LOG.Debug(err)
		return "", err
	}
	defer res.Body.Close()
	doc, err := ioutil.ReadAll(res.Body)
	return string(doc), nil
}

func parseContent(url int, doc string, urls chan int) <-chan string {
	content := make(chan string)
	go func() {
		content <- doc
		chk := false
		val := -1
		fetched.Lock()
		for n := range HCIDS {
			if _, ok := fetched.m[HCIDS[n]]; !ok {
				chk = true
				val = HCIDS[n]
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

	doc = <-parseContent(url, doc, urls)
	//	LOG.Debug("---- %d %s", url, doc)
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

	pad := maxlen - currlen
	for i := 1; i <= pad; i++ {
		out = fmt.Sprintf("%s.%s", "0", out)
	}
	LOG.Debug("shard: %s", out)
	return out
}

func mainExec() int {
	start := time.Now()
	r := new(big.Int)
	fmt.Println(r.Binomial(1000, 10))

	numCPUs := runtime.NumCPU()
	runtime.GOMAXPROCS(numCPUs)

	urls := make(chan int)
	done := make(chan struct{})
	c := make(chan result)

	var wg sync.WaitGroup
	const numWorkers = 100
	wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go func() {
			worker(done, urls, c)
			wg.Done()
		}()
	}

	go func() {
		wg.Wait()
		close(urls)
		close(c)
	}()

	urls <- -1

	count := 0
	for a := range c {
		LOG.Debug("==== %d %s", a.url, a.content)
		count++
		if count > len(HCIDS) {
			close(done)
			break
		}
	}

	elapsed := time.Since(start)
	LOG.Debug("It took %s", elapsed)
	return count - 1
}

func readLine(path string) string {
	var content = ""
	inFile, err := os.Open(path)
	if err != nil {
		fmt.Println(err.Error() + `: ` + path)
		return ""
	} else {
		defer inFile.Close()
	}
	scanner := bufio.NewScanner(inFile)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		fmt.Println(scanner.Text()) // the line
		content += scanner.Text() + "\n"
	}
	return content
}

// http://localhost:8080/main/core/1418,1419,2502,2694,2932,2933,2695
// http://localhost:8080/main/graphite/1418,1419,2502,2694,2932,2933,2695
func mainHandle(w http.ResponseWriter, r *http.Request) {
	STYPE = r.URL.Query().Get(":type")
	hcidStr := r.URL.Query().Get(":hcids")
	fmt.Println(STYPE, hcidStr)

	arry := strings.Split(hcidStr, ",")
	for i := range arry {
		n, _ := strconv.Atoi(arry[i])
		HCIDS = append(HCIDS, n)
	}

	res := make(map[string]string)
	result := mainExec()

	res["status"] = "OK"
	res["ts"] = time.Now().String()
	res["count"] = strconv.Itoa(result)
	res["result"] = readLine(LOGPATH)

	b, err := json.Marshal(res)
	if err != nil {
		LOG.Errorf("error: %s", err)
	}
	w.Write(b)
	return
}

func webserver() {
	//kill channel to programatically
	killch := make(chan os.Signal, 1)
	signal.Notify(killch, os.Interrupt)
	signal.Notify(killch, syscall.SIGTERM)
	signal.Notify(killch, syscall.SIGINT)
	signal.Notify(killch, syscall.SIGQUIT)
	go func() {
		<-killch
		LOG.Fatalf("Interrupt %s", time.Now().String())
	}()

	httphost := "localhost"
	httpport := "8080"

	//we need to start 2 servers, http for status and zmq
	wg := &sync.WaitGroup{}
	wg.Add(1)
	//first start http interface for self stats
	go func() {
		r := pat.New()
		r.Get("/main/{type}/{hcids}", http.HandlerFunc(mainHandle))
		http.Handle("/", r)

		LOG.Debug("Listening %s : %s", httphost, httpport)
		err := http.ListenAndServe(httphost+":"+httpport, nil)
		if err != nil {
			LOG.Fatalf("ListenAndServe: %s", err)
		}
		wg.Done()
	}()

	wg.Wait()
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
// 2way of run
// - 1st: cmd web
// 		call from brower: http://localhost:8080/main/core/1418,1419,2502,2694,2932,2933,2695
// - 2nd: cmd core/graphite 1418,1419,2502,2694,2932,2933,2695
//////////////////////////////////////////////////////////////////////////////////////////////////////
func main() {
	LOGFILE, err := os.OpenFile(LOGPATH, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		LOG.Fatalf("Log file error: %s %s", LOGPATH, err)
	}
	defer func() {
		LOGFILE.WriteString(fmt.Sprintf("closing %s", time.UnixDate))
		LOGFILE.Close()
	}()

	logback := logging.NewLogBackend(LOGFILE, "", 0)
	logformatted := logging.NewBackendFormatter(logback, LOGFORMAT)
	loglevel := "DEBUG"
	GLOGLEVEL, err := logging.LogLevel(loglevel)
	if err != nil {
		GLOGLEVEL = logging.DEBUG
	}
	logging.SetBackend(logformatted)
	logging.SetLevel(GLOGLEVEL, "")

	programName := os.Args[0:1]
	if len(os.Args) < 2 {
		HCIDS = append(HCIDS, 1418, 1419, 2502, 2694, 2932, 2933, 2695)
		mainExec()
	} else {
		typeStr := os.Args[1:2]
		if len(os.Args) >= 3 {
			hcidStr := os.Args[2:3]
			allArgs := os.Args[1:]
			fmt.Println(programName, typeStr, hcidStr, allArgs)
			arry := strings.Split(hcidStr[0], ",")
			for i := range arry {
				n, _ := strconv.Atoi(arry[i])
				HCIDS = append(HCIDS, n)
			}
		} else {
			allArgs := os.Args[1:]
			fmt.Println(programName, typeStr, allArgs)
		}
		if typeStr[0] == "web" {
			webserver()
		} else {
			STYPE = typeStr[0]
			mainExec()
		}
	}
	fmt.Println(HCIDS)
}
