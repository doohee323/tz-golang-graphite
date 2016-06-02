package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	_ "github.com/marpaia/graphite-golang"
	logging "github.com/op/go-logging"
	_ "github.com/stretchr/testify/assert"
	"log"
	"os/exec"
	"runtime"
	"strings"
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

	ExampleLookPath()
	//ExampleCommand()
	ExampleCmd_Output()
	ExampleCmd_Start()
	ExampleCmd_StdoutPipe()
}

func ExampleLookPath() {
	path, err := exec.LookPath("ping")
	if err != nil {
		log.Fatal("installing fortune is in your ping")
	}
	fmt.Printf("fortune is available at %s\n", path)
}

func ExampleCommand() {
	cmd := exec.Command("ping", "a")
	cmd.Stdin = strings.NewReader("some input")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("in all caps: %q\n", out.String())
}

func ExampleCmd_Output() {
	out, err := exec.Command("date").Output()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("The date is %s\n", out)
}

func ExampleCmd_Start() {
	cmd := exec.Command("sleep", "5")
	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Waiting for command to finish...")
	err = cmd.Wait()
	log.Printf("Command finished with error: %v", err)
}

func ExampleCmd_StdoutPipe() {
	cmd := exec.Command("echo", "-n", `{"Name": "Bob", "Age": 32}`)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}
	var person struct {
		Name string
		Age  int
	}
	if err := json.NewDecoder(stdout).Decode(&person); err != nil {
		log.Fatal(err)
	}
	if err := cmd.Wait(); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s is %d years old\n", person.Name, person.Age)
}
