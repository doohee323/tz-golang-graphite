package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	logging "github.com/op/go-logging"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"testing"
)

func TestBlah(t *testing.T) {
	var log = logging.MustGetLogger("graphite_test")
	log.Debugf("start~~")
	//var Apiurl = "http://dhcapi.local.xdn.com/dhc/5/cloud/jobs?site=nuq2"

	//	var id map[string]interface{}
	//	var ok bool
	//	if x, found := m[key]; found {
	//		if id, ok = x.(map[string]interface{}); !ok {
	//			fmt.Println(id)
	//			//			err = json.Unmarshal([]byte(id), &m)
	//			//			fmt.Println(m["rules"])
	//		} else {
	//			fmt.Println(m)
	//		}
	//	} else {
	//		fmt.Println(m)
	//	}

	var Apiurl = "http://localhost:8080/rule/viktor-r11264-test-01.fortidirector.http.r1cd.com"
	//	encoded := `
	//			    {"CNAME::viktor-r11264-test-01.fortidirector.http.r1cd.com":{"rules":[{"rule_id":252,"is_default":1,"user_id":1,"action_type":"round_robin","action_meta":{"resources":[{"type":"network","id":1073,"base_url":"http://google.com/","latitude":37.4192,"longitude":-122.0574}]}}],"ruleset_id":270}}
	//				`
	encoded := GetHttp(Apiurl)
	CaseJrl(encoded)

	//		if els, ok := data.Control["commands"]; ok {
	//			if el, ok := els.([]interface{}); ok {
	//				for _, e := range el {
	//					if cmds, ok := e.(map[string]interface{}); ok {
	//						for k, v := range cmds {
	//							if reset, ok := v.(string); ok {
	//								log.Warningf("reset %s %v", reset, k)
	//							}
	//						}
	//					}
	//				}
	//			}
	//		}
	//
	//		//hack to prevent deletion
	//		log.Warningf("status %s", data.Control["status"])
	//		if el, ok := data.Control["status"]; !ok {
	//			if els, ok := el.(string); ok {
	//				log.Warningf("reset %s", els)
	//			}
	//		}

	//t.Errorf("blah blah")
	assert.Equal(t, 123, 123, "they should be equal")
	//t.Fatal()
}

func CaseJrl(encoded string) {
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
		return
	}
	fmt.Println(data.R1.Rules)
	fmt.Println(reflect.TypeOf(data.R1.Rules))
	fmt.Println(data.R1.Rules[0]["rule_id"])
}

// case 1:
//	// map, json parsing
//	b := []byte(`{"Name":"Wednesday","Age":6,"Parents":["Gomez","Morticia"]}`)
//	var f interface{}
//	json.Unmarshal(b, &f)
//	m1 := f.(map[string]interface{})
//	fmt.Println("is string", m1["Name"])
//	var testn = 5
//	f = map[string]interface{}{
//		"Name": "Wednesday",
//		"Age":  testn,
//		"Parents": map[string]interface{}{
//			"Name": "Gomez",
//			"Age": 10,
//		},
//	}
//	m1 = f.(map[string]interface{})
//	fmt.Println("is string", m1["Age"])
//	var m2 = m1["Parents"].(map[string]interface{})
//	fmt.Println("is string", m1["Parents"])
//	fmt.Println("is string", m2["Age"])
//
// case 2:
//	encoded := `
//    {"rule": {
//		"person" : {
//		        "name": "john",
//		        "age": 23
//			}
//		}
//    }`
//
//	var m map[string]interface{}
//	err := json.Unmarshal([]byte(encoded), &m)
//	if err != nil {
//		panic(err)
//	}
//	fmt.Println(m)
//
//	var m3 = m["rule"].(map[string]interface{})
//	fmt.Println("is string", m3)
//	var m4 = m3["person"].(map[string]interface{})
//	fmt.Println("is string", m4)
//
// case 3:
//	encoded := `
//	    {"CNAME::viktor-r11264-test-01.fortidirector.http.r1cd.com": {
//			"rules" : {
//			        "name": "john",
//			        "age": 23
//				}
//			}
//	    }`
//	var m map[string]interface{}
//	err := json.Unmarshal([]byte(encoded), &m)
//	if err != nil {
//		panic(err)
//	}
//	var key = "CNAME::viktor-r11264-test-01.fortidirector.http.r1cd.com"
//	fmt.Println(m[key])
//
//	var CNAME = m[key].(map[string]interface{})
//	fmt.Println(CNAME)
//	var rules = CNAME["rules"].(map[string]interface{})
//	fmt.Println(rules)
//	fmt.Println(rules["name"])
//
//	var rules2 = m[key].(map[string]interface{})["rules"].(map[string]interface{})
//	fmt.Println(rules2)

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
		log.Errorf("HC API error %s", err)
		return ""

	} else { //good to go
		defer response.Body.Close()
		contents, err := ioutil.ReadAll(response.Body)
		log.Errorf("contents: %s", contents)

		encoded := contents

		log.Errorf("contents: %s", encoded)
		if err != nil {
			log.Errorf("HC API error: %s", err)
			return ""
		}

		if response.StatusCode != 200 {
			log.Errorf("HC API response code: %d", response.StatusCode)
			return ""
		}

		return string(contents)
	}
}
