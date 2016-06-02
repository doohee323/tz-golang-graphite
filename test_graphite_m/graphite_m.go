package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gorilla/pat"
	logging "github.com/op/go-logging"
	"github.com/vaughan0/go-ini"
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

var (
	HCIDS       []int
	STYPE       string
	LOGFILE     *os.File
	DOMAIN      string
	CORE_URL                  = "CORE_URL/1/stats/uptime_list?company_id=1&start_time=1464636372&end_time=1464722772&hc_id="
	GRAPHTE_URL               = "http://173.243.129.12/render?target=summarize(averageSeries(dhc.stats.hcid.HCID.*.metrics.state.*),%221hour%22,%22avg%22)&target=summarize(averageSeries(dhc.stats.hcid.HCID.*.judge.state),%221hour%22,%22avg%22)&from=-1440min&until=-0min&format=json"
	LOGPATH                   = "/var/log/dashboard/graphite_m.log"
	LOGFMT                    = "%{color}%{time:15:04:05.000000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}"
	LOGFORMAT                 = logging.MustStringFormatter(LOGFMT)
	LOG                       = logging.MustGetLogger("logfile")
	GLOGLEVEL   logging.Level = logging.DEBUG
	CONFIGFILE                = "/etc/dashboard/dashboard.ini"
	configfile  *string       = flag.String("config", CONFIGFILE, "Config file location default: "+CONFIGFILE)
	cfg         ini.File
)

type FetchedResult struct {
	url     int
	content string
}

type FetchedUrl struct {
	m map[int]error
	sync.Mutex
}

type Crawler interface {
	Crawl()
}

type GraphiteFetch struct {
	fetchedUrl *FetchedUrl
	p          *Pipeline
	result     chan FetchedResult
	url        int
}

func fetch(url int) (string, error) {
	if url == -1 {
		return "", nil
	}
	var urlstr = ""
	if STYPE == "graphite" {
		var hcid = formatHcid(strconv.Itoa(url))
		var baseUrl = GRAPHTE_URL
		urlstr = strings.Replace(baseUrl, "HCID", hcid, -1)
	} else {
		core_url := strings.Replace(CORE_URL, "CORE_URL", DOMAIN, -1)
		urlstr = core_url + strconv.Itoa(url)
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

func (g *GraphiteFetch) Request(url int) {
	g.p.request <- &GraphiteFetch{
		fetchedUrl: g.fetchedUrl,
		p:          g.p,
		result:     g.result,
		url:        url,
	}
}

func (g *GraphiteFetch) parseContent(url int, doc string) <-chan string {
	content := make(chan string)
	go func() {
		content <- doc
		chk := false
		val := -1
		g.fetchedUrl.Lock()
		for n := range HCIDS {
			if _, ok := g.fetchedUrl.m[HCIDS[n]]; !ok {
				chk = true
				val = HCIDS[n]
				g.Request(val)
				break
			}
		}
		if chk == false {

		}
		g.fetchedUrl.Unlock()
	}()
	return content
}

func (g *GraphiteFetch) Crawl() {
	g.fetchedUrl.Lock()
	if _, ok := g.fetchedUrl.m[g.url]; ok {
		g.fetchedUrl.Unlock()
		return
	}
	g.fetchedUrl.Unlock()

	doc, err := fetch(g.url)
	if err != nil {
		go func(u int) {
			g.Request(u)
		}(g.url)
		return
	}

	g.fetchedUrl.Lock()
	g.fetchedUrl.m[g.url] = err
	g.fetchedUrl.Unlock()

	content := <-g.parseContent(g.url, doc)
	g.result <- FetchedResult{g.url, content}
}

type Pipeline struct {
	request chan Crawler
	done    chan struct{}
	wg      *sync.WaitGroup
}

func NewPipeline() *Pipeline {
	return &Pipeline{
		request: make(chan Crawler),
		done:    make(chan struct{}),
		wg:      new(sync.WaitGroup),
	}
}

func (p *Pipeline) Worker() {
	for r := range p.request {
		select {
		case <-p.done:
			return
		default:
			r.Crawl()
		}
	}
}

func (p *Pipeline) Run() {
	const numWorkers = 100
	p.wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go func() {
			p.Worker()
			p.wg.Done()
		}()
	}

	go func() {
		p.wg.Wait()
	}()
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

func mainExec() map[string]string {
	start := time.Now()
	r := new(big.Int)
	fmt.Println(r.Binomial(1000, 10))

	numCPUs := runtime.NumCPU()
	runtime.GOMAXPROCS(numCPUs)

	p := NewPipeline()
	p.Run()

	following := &GraphiteFetch{
		fetchedUrl: &FetchedUrl{m: make(map[int]error)},
		p:          p,
		result:     make(chan FetchedResult),
		url:        -1,
	}
	p.request <- following

	result := make(map[string]string)
	count := 0
	LOG.Debug("=len(HCIDS)=== %d", len(HCIDS))
	for a := range following.result {
		LOG.Debug("==== %d %s", a.url, a.content)
		count++
		countStr := strconv.Itoa(count)
		result[countStr] = a.content
		if count > len(HCIDS) {
			close(p.done)
			break
		}
	}

	elapsed := time.Since(start)
	LOG.Debug("It took %s", elapsed)
	return result
}

func readValConf(path string, key string) string {
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
		str := scanner.Text()
		if strings.Index(str, key) != -1 {
			str = str[strings.Index(str, "=")+1 : len(str)]
			return strings.Replace(strings.TrimSpace(str), "\"", "", -1)
		}
	}
	return ""
}

// http://localhost:8080/main/core/1418,1419,2502,2694,2932,2933,2695
// http://localhost:8080/main/graphite/1418,1419,2502,2694,2932,2933,2695
func mainHandle(w http.ResponseWriter, r *http.Request) {
	STYPE = r.URL.Query().Get(":type")
	hcidStr := r.URL.Query().Get(":hcids")
	fmt.Println(STYPE, hcidStr)

	HCIDS = []int{}
	arry := strings.Split(hcidStr, ",")
	for i := range arry {
		n, _ := strconv.Atoi(arry[i])
		HCIDS = append(HCIDS, n)
	}

	res := make(map[string]string)
	result := mainExec()

	res["status"] = "OK"
	res["ts"] = time.Now().String()
	str, err := json.Marshal(result)
	fmt.Println(err)
	fmt.Println(str)
	res["count"] = strconv.Itoa(len(result) - 1)
	res["result"] = string(str)

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
// - 1st: graphite_m web
// 		call from brower: http://localhost:8080/main/core/1418,1419,2502,2694,2932,2933,2695
// - 2nd: graphite_m core/graphite 1418,1419,2502,2694,2932,2933,2695
//////////////////////////////////////////////////////////////////////////////////////////////////////
func main() {

	var logfile = LOGPATH
	if _, err := os.Stat(LOGPATH); err != nil {
		LOGPATH, _ := os.Getwd()
		logfile = LOGPATH + "/graphite_m.log"
	}

	LOGFILE, err := os.OpenFile(logfile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		LOG.Fatalf("Log file error: %s %s", logfile, err)
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

	//	cfg, err := ini.LoadFile(*configfile)
	//	if err != nil {
	//		LOG.Fatalf("parse config "+*configfile+" file error: %s", err)
	//	}
	//
	//	logfile, ok := cfg.Get("core_api_url", "logfile")
	//	if !ok {
	//		LOG.Fatalf("'logfile' missing from 'system' section")
	//	}
	DOMAIN = readValConf(CONFIGFILE, "core_api_url")
	if DOMAIN == "" {
		DOMAIN = "http://core.local.xdn.com"
	}

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
