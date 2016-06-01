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
	logFile   *os.File
	logPath                 = "/Users/dhong/tmp/test.log"
	LOGFMT                  = "%{color}%{time:15:04:05.000000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}"
	logFormat               = logging.MustStringFormatter(LOGFMT)
	log                     = logging.MustGetLogger("logfile")
	Gloglevel logging.Level = logging.DEBUG
)

var fetched = struct {
	m map[int]error
	sync.Mutex
}{m: make(map[int]error)}

type result struct {
	url     int
	content string
}

const (
	CORE_URL    = "http://core.local.xdn.com/1/stats/uptime_list?company_id=1&start_time=1464636372&end_time=1464722772&hc_id="
	GRAPHTE_URL = "http://173.243.129.129/render?target=summarize(averageSeries(dhc.stats.hcid.0.0.1.4.1.8.*.metrics.state.*),%221hour%22,%22avg%22)&target=summarize(averageSeries(dhc.stats.hcid.0.0.1.4.1.8.*.judge.state),%221hour%22,%22avg%22)&from=-1440min&until=-0min&format=json"
)

var _hcids []int
var _type string

func fetch(url int) (string, error) {
	if url == -1 {
		return "", nil
	}
	var urlstr = ""
	if _type == "core" {
		urlstr = CORE_URL + strconv.Itoa(url)
	} else {
		var hcid = "0.0.1.4.1.8"
		var hcid2 = formatHcid(strconv.Itoa(url))
		var baseUrl = GRAPHTE_URL
		urlstr = strings.Replace(baseUrl, hcid, hcid2, -1)
	}

	log.Debug("==== %s", urlstr)
	res, err := http.Get(urlstr)
	if err != nil {
		log.Debug(err)
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
		for n := range _hcids {
			if _, ok := fetched.m[_hcids[n]]; !ok {
				chk = true
				val = _hcids[n]
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
	//	log.Debug("---- %d %s", url, doc)
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
	log.Debug("shard: %s", out)
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
	for r := range c {
		log.Debug("==== %d %s", r.url, r.content)
		count++
		if count > len(_hcids) {
			close(done)
			break
		}
	}

	elapsed := time.Since(start)
	log.Debug("It took %s", elapsed)
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
	_type = r.URL.Query().Get(":type")
	hcidStr := r.URL.Query().Get(":hcids")
	fmt.Println(_type, hcidStr)

	arry := strings.Split(hcidStr, ",")
	for i := range arry {
		n, _ := strconv.Atoi(arry[i])
		_hcids = append(_hcids, n)
	}

	res := make(map[string]string)
	result := _main()

	res["status"] = "OK"
	res["ts"] = time.Now().String()
	res["count"] = strconv.Itoa(result)
	res["result"] = readLine(logPath)

	b, err := json.Marshal(res)
	if err != nil {
		log.Errorf("error: %s", err)
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
		log.Fatalf("Interrupt %s", time.Now().String())
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

		log.Debug("Listening %s : %s", httphost, httpport)
		err := http.ListenAndServe(httphost+":"+httpport, nil)
		if err != nil {
			log.Fatalf("ListenAndServe: %s", err)
		}
		wg.Done()
	}()

	wg.Wait()
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
// 2way of run
// - 1st: go_cmd web
// 		call from brower: http://localhost:8080/main/core/1418,1419,2502,2694,2932,2933,2695
// - 2nd: go_cmd core/graphite 1418,1419,2502,2694,2932,2933,2695
//////////////////////////////////////////////////////////////////////////////////////////////////////
func main() {
	logFile, err := os.OpenFile(logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Log file error: %s %s", logPath, err)
	}
	defer func() {
		logFile.WriteString(fmt.Sprintf("closing %s", time.UnixDate))
		logFile.Close()
	}()

	logback := logging.NewLogBackend(logFile, "", 0)
	logformatted := logging.NewBackendFormatter(logback, logFormat)
	loglevel := "DEBUG"
	Gloglevel, err := logging.LogLevel(loglevel)
	if err != nil {
		Gloglevel = logging.DEBUG
	}
	logging.SetBackend(logformatted)
	logging.SetLevel(Gloglevel, "")

	programName := os.Args[0:1]
	if len(os.Args) < 2 {
		_hcids = append(_hcids, 1418, 1419, 2502, 2694, 2932, 2933, 2695)
		_main()
	} else {
		typeStr := os.Args[1:2]
		if len(os.Args) >= 3 {
			hcidStr := os.Args[2:3]
			allArgs := os.Args[1:]
			fmt.Println(programName, typeStr, hcidStr, allArgs)
			arry := strings.Split(hcidStr[0], ",")
			for i := range arry {
				n, _ := strconv.Atoi(arry[i])
				_hcids = append(_hcids, n)
			}
		} else {
			allArgs := os.Args[1:]
			fmt.Println(programName, typeStr, allArgs)
		}
		if typeStr[0] == "web" {
			webserver()
		} else {
			_type = typeStr[0]
			_main()
		}
	}
	fmt.Println(_hcids)
}
