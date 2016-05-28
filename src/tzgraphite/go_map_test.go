package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	_ "github.com/marpaia/graphite-golang"
	logging "github.com/op/go-logging"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"runtime"
	"strconv"
	"testing"
)

func TestMap(t *testing.T) {
	var log = logging.MustGetLogger("TestMap")

	runtime.GOMAXPROCS(1)
//	runtime.GOMAXPROCS(runtime.NumCPU())
	fmt.Println(runtime.GOMAXPROCS(0))
	fmt.Println("CPU Count : ", runtime.NumCPU())

	log.Debugf("start~~")
	//var Apiurl = "http://dhcapi.local.xdn.com/dhc/5/cloud/jobs?site=nuq2"

	var Apiurl = "http://localhost:8080/rule/viktor-r11264-test-01.fortidirector.http.r1cd.com"
	encoded := GetHttp(Apiurl)
	if encoded == "" {
		encoded = `
		    {"CNAME::viktor-r11264-test-01.fortidirector.http.r1cd.com":{"rules":[{"rule_id":252,"is_default":1,"user_id":1,"action_type":"round_robin","action_meta":{"resources":[{"type":"network","id":1073,"base_url":"http://google.com/","latitude":37.4192,"longitude":-122.0574}]}}],"ruleset_id":270}}
			`
	}
	var rslt = CaseJrl(encoded)
	if assert.NotNil(t, rslt) {
		if !assert.Equal(t, rslt, "ok", "they should be equal") {
			t.Errorf("they should be equal!")
			t.Fatal()
		}
	}

	encoded = `{"Name":"Wednesday","Age":6,"Parents":["Gomez","Morticia"]}`
	CaseMap1(encoded)

	encoded = `
	    {"rule": {
			"person" : {
			        "name": "john",
			        "age": 23
				}
			}
	    }`
	CaseMap2(encoded)

	encoded = `
	    {"CNAME::viktor-r11264-test-01.fortidirector.http.r1cd.com": {
			"rules" : {
			        "name": "john",
			        "age": 23
				}
			}
	    }`
	CaseMap2(encoded)
}

func CaseJrl(encoded string) string {
	type Rules struct {
		Rules []map[string]interface{} `json:"rules"`
	}
	type CName struct {
		R1 Rules `json:"CNAME::viktor-r11264-test-01.fortidirector.http.r1cd.com"`
	}
	var data CName

	err = json.Unmarshal([]byte(encoded), &data)
	if err != nil {
		log.Errorf("HC Obj Unpacking error %s", err)
		return "fail"
	}
	fmt.Println(data.R1.Rules)
	fmt.Println(reflect.TypeOf(data.R1.Rules))
	fmt.Println(data.R1.Rules[0]["rule_id"])

	log.Warningf("status %v", data.R1.Rules[0]["rule_id"])
	if el, ok := data.R1.Rules[0]["action_type"]; ok {
		if els, ok := el.(string); ok {
			log.Warningf("action_type %s", els)
		}
	}

	var m1 map[string]interface{}
	if x, found := data.R1.Rules[0]["action_meta"]; found {
		if m1, found = x.(map[string]interface{}); found {
			fmt.Println(m1)
			var m3 = m1["resources"].([]interface{})
			fmt.Println("-------", m3)
			var m2 = m3[0].(map[string]interface{})
			fmt.Println("-------", m2["base_url"])

			var m4 = m1["resources"].([]interface{})[0].(map[string]interface{})
			fmt.Println("-------", m4["base_url"])

			var m5 = m1["resources"].([]interface{})
			for _, e := range m5 {
				if cmds, ok := e.(map[string]interface{}); ok {
					for k, v := range cmds {
						//						atype := reflect.TypeOf(v)
						switch v.(type) {
						case int:
							log.Warningf("v %v %v", v, k)
							break
						case float32:
							log.Warningf("v %v %v", v, k)
							break
						case float64:
							log.Warningf("v %v %v", v, k)
							break
						case string:
							log.Warningf("v %v %v", v, k)
							break
						case bool:
							log.Warningf("v %v %v", v, k)
							break
						default:
							log.Warningf("v %v %v", v, k)
							break
						}
					}
				}
			}
		} else {
			fmt.Println(x)
		}
	} else {
		fmt.Println(data)
	}
	return "ok"
}

func CaseMap1(encoded string) {
	// map, json parsing
	var f interface{}
	json.Unmarshal([]byte(encoded), &f)
	m1 := f.(map[string]interface{})
	fmt.Println("-------", m1["Name"])
	var testn = 5
	f = map[string]interface{}{
		"Name": "Wednesday",
		"Age":  testn,
		"Parents": map[string]interface{}{
			"Name": "Gomez",
			"Age":  10,
		},
	}
	m1 = f.(map[string]interface{})
	fmt.Println("m1", m1)
	fmt.Println("-------", m1["Age"])
	var m2 = m1["Parents"].(map[string]interface{})
	fmt.Println("-------", m1["Parents"])
	fmt.Println("-------", m2["Age"])
}

func CaseMap2(encoded string) {
	var m map[string]interface{}
	err := json.Unmarshal([]byte(encoded), &m)
	if err != nil {
		errorMsg(err)
	}
	fmt.Println(m)

	var m3 = m["rule"].(map[string]interface{})
	fmt.Println("-------", m3)
	var m4 = m3["person"].(map[string]interface{})
	fmt.Println("-------", m4)
}

func CaseMap3(encoded string) {
	var m map[string]interface{}
	err := json.Unmarshal([]byte(encoded), &m)
	if err != nil {
		errorMsg(err)
	}
	var key = "CNAME::viktor-r11264-test-01.fortidirector.http.r1cd.com"
	fmt.Println(m[key])

	var CNAME = m[key].(map[string]interface{})
	fmt.Println(CNAME)
	var rules = CNAME["rules"].(map[string]interface{})
	fmt.Println(rules)
	fmt.Println(rules["name"])

	var rules2 = m[key].(map[string]interface{})["rules"].(map[string]interface{})
	fmt.Println(rules2)
}

func GetHttp(Apiurl string) string {
	//var Apiurl = "http://localhost:8080/rule/viktor-r11264-test-01.fortidirector.http.r1cd.com"
	//	var Apiuser = "xdn"
	//	var Apipassword = "3cftw2010"
	client := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}

	//get url content first
	log.Debugf("Querying %s", Apiurl)

	u, _ := url.Parse(Apiurl)
	q := u.Query()
	q.Set("version", Build)
	Apiurl = u.String()

	req, err := http.NewRequest("GET", Apiurl, nil)
	//	req.SetBasicAuth(Apiuser, Apipassword)

	response, err := client.Do(req)
	if err != nil {
		errorMsg(err)
		return ""
	} else { //good to go
		defer response.Body.Close()
		contents, err := ioutil.ReadAll(response.Body)
		log.Errorf("contents: %s", contents)
		if err != nil {
			errorMsg(err)
		}
		if response.StatusCode != 200 {
			errorMsg("response code: " + strconv.Itoa(response.StatusCode))
		}
		return string(contents)
	}
}

func errorMsg(err interface{}) {
	defer func() {
		s := recover()
		log.Errorf("API error %v", s)
	}()
	panic(err)
}
