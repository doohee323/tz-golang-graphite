/**
* Graphite Module by Dewey
* Usage: tzgraphite -method=simplesend -key=deploys.test1.count2 -value=50 (-config=/Users/dhong/Documents/workspace/go/src/tz.com/tz_golang_graphite/etc/graphite.cfg)
**/

package main

import (
	"encoding/json"
	"flag"
	graphite "github.com/marpaia/graphite-golang"
	logging "github.com/op/go-logging"
//	"runtime"
	"strconv"
	"time"
	tz "tz.com/tz_golang_graphite/src/tz"
)

var log = logging.MustGetLogger("graphite")
var err error
var Build string
var Metric graphite.Metric

var method = flag.String("method", "", "method")
var key = flag.String("key", "", "key")
var value = flag.String("value", "", "value")

func init() {
}

func main() {
	flag.Parse()

//	runtime.GOMAXPROCS(runtime.NumCPU())

	log.Debug("%s\n", *method)
	log.Debug("%s\n", *key)
	log.Debug("%s\n", *value)

	if *method == "simplesend" {
		// https://172.30.168.10/render?target=deploys.test1.count2&from=-10min&rawData=true&format=json
		if *key == "" {
			*key = "deploys.test1.count2"
		}
		if *value == "" {
			*value = "11"
		}

		type SimpleMessage struct {
			Value string
		}
		m := SimpleMessage{*value}
		data, err := json.Marshal(m)
		log.Debug("%s\n", data)

		err = tz.NewGraphite(*method, *key, data)
		if err != nil {
			log.Debug("%s\n", err)
		}
	} else if *method == "sendmetric" {
		// https://172.30.168.10/render?target=deploys.test1.count3&from=-10min&rawData=true&format=json
		if *key == "" {
			*key = "deploys.test1.count3"
		}
		if *value == "" {
			*value = "40"
		}

		type Metric2 struct {
			Value string
		}

		m1 := Metric2{*value}
		data1, err := json.Marshal(m1)
		log.Debug("%s\n", data1)
		err = tz.NewGraphite(*method, *key, data1)
		if err != nil {
			log.Debug("%s\n", err)
		}
	} else if *method == "sendmetrics" {
		// https://172.30.168.10/render?target=deploys.test1.count4&from=-10min&rawData=true&format=json
		if *key == "" {
			*key = "deploys.test1.count4"
		}
		if *value == "" {
			*value = "522"
		}

		type Metrics2 struct {
			Date  string
			Value string
		}

		var mArray []Metrics2
		mArray = append(mArray, Metrics2{strconv.FormatInt(time.Now().Unix(), 10), *value})
		data2, err := json.Marshal(mArray)
		err = tz.NewGraphite(*method, *key, data2)
		if err != nil {
			log.Debug("%s\n", err)
		}
	}

}
