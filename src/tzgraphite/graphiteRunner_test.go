/**
* Graphite Module by Dewey
* Usage: tzgraphite.test -config=/Users/dhong/Documents/workspace/go/src/tz.com/tz_golang_graphite/etc/graphite.cfg
**/

package main

import (
	"bufio"
	"encoding/json"
	"github.com/op/go-logging"
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
	tz "tz.com/tz_golang_graphite/src/tz"
)

func TestXxx(t *testing.T) {
	var log = logging.MustGetLogger("graphite_test")

	backend1 := logging.NewLogBackend(os.Stderr, "", 0)
	backend2 := logging.NewLogBackend(os.Stderr, "", 0)
	var format = logging.MustStringFormatter(
		`%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`,
	)
	
	backend2Formatter := logging.NewBackendFormatter(backend2, format)
	backend1Leveled := logging.AddModuleLevel(backend1)
    backend1Leveled.SetLevel(logging.ERROR, "")

    logging.SetBackend(backend1Leveled, backend2Formatter)
	
	var test_types = []string{"simplesend", "sendmetric", "sendmetrics"}
	//	pwd, _ := os.Getwd()
	pwd := "/Users/dhong/Documents/workspace/go/src/tz.com/tz_golang_graphite"
	result := 0
	for _, tType := range test_types {
		test_file := pwd + "/etc/" + tType + ".txt"
		log.Debug("test_file: %s\n", test_file)
		f, err := os.Open(test_file)
		if err != nil {
			log.Error("error opening file ", err)
			t.Errorf("error: %s", err)
			os.Exit(1)
		}
		defer f.Close()
		r := bufio.NewReader(f)
		for {
			str, err := r.ReadString(10)
			if err == io.EOF {
				break
			} else if err != nil {
				log.Error("%s\n", err)
				t.Errorf("error: %s", err)
				result++
			}
			log.Debug("str ", str)
			arry := strings.Split(str, ",")
			if tType == "simplesend" {
				// https://172.30.168.10/render?target=deploys.test1.count1&from=-10min&rawData=true&format=json
				type SimpleMessage struct {
					Value string
				}
				m := SimpleMessage{strings.Trim(arry[1], "\n")}
				data, err := json.Marshal(m)
				log.Debug("%s\n", data)
				err = tz.NewGraphite(tType, arry[0], data)
				if err != nil {
					log.Error("%s\n", err)
					t.Errorf("error: %s for %+v", err, data)
					result++
				}
			} else if tType == "sendmetric" {
				// https://172.30.168.10/render?target=deploys.test1.count2&from=-10min&rawData=true&format=json
				type Metric2 struct {
					Value string
				}
				m1 := Metric2{strings.Trim(arry[1], "\n")}
				data, err := json.Marshal(m1)
				log.Debug("%s\n", data)
				err = tz.NewGraphite(tType, arry[0], data)
				if err != nil {
					log.Error("%s\n", err)
					t.Errorf("error: %s for %+v", err, data)
					result++
				}
			} else if tType == "sendmetrics" {
				// https://172.30.168.10/render?target=deploys.test1.count3&from=-10min&rawData=true&format=json
				type Metrics2 struct {
					Date  string
					Value string
				}
				var mArray []Metrics2
				mArray = append(mArray, Metrics2{strconv.FormatInt(time.Now().Unix(), 10), strings.Trim(arry[1], "\n")})
				data, err := json.Marshal(mArray)
				err = tz.NewGraphite(tType, arry[0], data)
				if err != nil {
					log.Error("%s\n", err)
					t.Errorf("error: %s for %+v", err, data)
					result++
				}
			}
			time.Sleep(1000 * time.Millisecond)
		}
	}
	assert.Equal(t, result, 0, "at least test should pass")
}
