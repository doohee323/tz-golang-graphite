package main

import (
	"bufio"
	"encoding/json"
	logging "github.com/op/go-logging"
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
	"tz"
)

func TestXxx(t *testing.T) {
	var log = logging.MustGetLogger("graphite")
	var test_types = []string{"simplesend", "sendmetric", "sendmetrics"}
	pwd, _ := os.Getwd()
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
				// https://localhost/render?target=statsd.test.graphite_loaded1&from=-10min&rawData=true&format=json
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
				// https://localhost/render?target=statsd.test.graphite_loaded2&from=-10min&rawData=true&format=json
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
				// https://localhost/render?target=statsd.test.graphite_loaded3&from=-10min&rawData=true&format=json
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
	assert.NotEqual(t, result, 0, "at least test should pass")
}
