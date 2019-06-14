/**
* graphite Module by Dewey
**/

package tz

import (
	"encoding/json"
	"flag"
	"fmt"
	graphite "github.com/marpaia/graphite-golang"
	logging "github.com/op/go-logging"
	"github.com/vaughan0/go-ini"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

//this is log file
var logFile *os.File
var logFormat = logging.MustStringFormatter("%{color}%{time:15:04:05.000000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}")
var Gloglevel logging.Level = logging.DEBUG

//this where errors go to die
var err error

//this is version
var Build string

//the following are flags passed from commandline
var Configfile *string = flag.String("config", "/Users/dhong/Documents/workspace/go/src/tz.com/tz-golang-graphite/etc/graphite.cfg", "Config file location")
var help *bool = flag.Bool("help", false, "Show these options")
var version *string = flag.String("version", Build, "Build version")
var cfg ini.File
var Graphite *graphite.Graphite
var Metric graphite.Metric
var Metrics []graphite.Metric

var log = logging.MustGetLogger("logfile")

func NewGraphite(method string, key string, args []byte) error {

	runtime.GOMAXPROCS(runtime.NumCPU())

	//parse config
	log.Debug("%s\n", *Configfile)
	cfg, err = ini.LoadFile(*Configfile)

	if err != nil {
		log.Fatalf("parse config "+*Configfile+" file error: %s", err)
	}

	logfile, ok := cfg.Get("system", "logfile")
	if !ok {
		log.Fatal("'logfile' missing from 'system' section")
	}

	logDir := logfile[0:strings.LastIndex(logfile, "/")]
	err = os.MkdirAll(logDir, 0777)
	if err != nil {
		log.Fatalf("MkdirAll error: %s", err)
	}

	//open log file
	logFile, err := os.OpenFile(logfile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("%s, %s ", logfile, err)
	}

	defer func() {
		logFile.WriteString(fmt.Sprintf("closing %s", time.UnixDate))
		logFile.Close()
	}()

	logback := logging.NewLogBackend(logFile, "", 0)
	logformatted := logging.NewBackendFormatter(logback, logFormat)

	loglevel, ok := cfg.Get("system", "loglevel")
	if !ok {
		log.Fatal("'loglevel' missing from 'system' section")
	}

	loglevel = strings.ToUpper(loglevel)

	Gloglevel, err = logging.LogLevel(loglevel)
	if err != nil {
		Gloglevel = logging.DEBUG
	}

	logging.SetBackend(logformatted)

	//see what we have here
	log.Debug("BUILD: %s\n", Build)
	for name, section := range cfg {
		log.Debug("Section: %v\n", name)
		for k, v := range section {
			log.Debug("%v: %v\n", k, v)
		}
	}

	host, ok := cfg.Get("server", "host")
	if !ok {
		log.Fatal("'server' missing from 'host' section")
	}
	var port int
	port_str, ok := cfg.Get("server", "port")
	if !ok {
		log.Fatal("'server' missing from 'port' section")
	} else {
		v, err := strconv.Atoi(port_str)
		if err != nil {
			log.Warning("'server' missing from 'port' section")
		} else {
			port = v
		}
	}

	Graphite, err = graphite.NewGraphite(host, port)
	log.Debug("Loaded Graphite connection: %#v", Graphite)
	if method == "simplesend" {
		var meta map[string]interface{}
		json.Unmarshal(args, &meta)
		err = Graphite.SimpleSend(key, getArg(meta, "Value"))
		log.Debug("Loaded Graphite SimpleSend: %#v", err)
	} else if method == "sendmetric" {
		var meta map[string]interface{}
		json.Unmarshal(args, &meta)
		Metric.Name = key
		Metric.Value = getArg(meta, "Value")
		Metric.Timestamp = time.Now().Unix()
		err = Graphite.SendMetric(Metric)
		log.Debug("Loaded Graphite SendMetric: %#v", err)
	} else if method == "sendmetrics" {
		metas := make([]map[string]interface{}, 0)
		json.Unmarshal(args, &metas)
		log.Debug("metas: %#v", metas)
		for _, meta := range metas {
			Metric.Name = key
			Metric.Value = getArg(meta, "Value")
			Metric.Timestamp, _ = strconv.ParseInt(getArg(meta, "Date"), 10, 64)
			Metrics = append(Metrics, Metric)
		}
		log.Debug("Metrics: %#v", Metrics)
		err = Graphite.SendMetrics(Metrics)
		log.Debug("Loaded Graphite SendMetric: %#v", err)
	}
	Graphite.Disconnect()
	//	logging.SetLevel(Gloglevel, "")

	if err != nil {
		return err
	}
	return nil
}

func getArg(meta map[string]interface{}, key string) string {
	if el, ok := meta[key]; ok {
		if _, ok := el.(string); ok {
			v := el.(string)
			log.Debug("key: %v\n", v)
			return v
		} else {
			log.Error("method: Invalid " + key + " value")
			return ""
		}
	} else {
		log.Error("method: Invalid " + key + " value")
		return ""
	}
}
