package main

import (
	"encoding/json"
	"fmt"
	bs "github.com/bytecamp2019d/bustsurvivor/api/bustsurvivor"
	"github.com/bytecamp2019d/bustsurvivor/model/balancer"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

type TestCase struct {
	ConnNum      int `json:"connNum"`
	Rps          int `json:"rps"`
	LoopNum      int `json:"loopNum"`
	NextInterval int `json:"nextInterval"`
}

type TestCases struct {
	Cases []TestCase `json:"cases"`
}

// call .
func call() {
	configPath, _ := filepath.Abs("test_config.json")
	jsonFile, err := os.Open(configPath)

	if err != nil {
		panic(err)
	}
	defer jsonFile.Close()
	var tcs TestCases
	byteValue, _ := ioutil.ReadAll(jsonFile)
	_ = json.Unmarshal(byteValue, &tcs)
	for caseIndex, tc := range tcs.Cases {
		fmt.Printf("============= No.%d test case ===============\n", caseIndex+1)
		connNum := tc.ConnNum
		rps := tc.Rps
		loopNum := tc.LoopNum
		nextInterval := tc.NextInterval

		balancer.InitBalancer(connNum, caseIndex == 0)

		totalDur := time.Duration(0)
		durChan := make(chan time.Duration)
		errChan := make(chan error)
		errCnt := 0
		for i := 0; i < connNum; i++ {
			go func() {
				for j := 0; j < loopNum; j++ {
					req := &bs.BustSurvivalRequest{
						CardsToPick:   10,
						BustThreshold: 80,
					}
					go balancer.SendRequest(req, durChan, errChan)
					time.Sleep(time.Second / time.Duration(rps))
				}
			}()
			time.Sleep(time.Second / time.Duration(rps*connNum))
		}

		for i := 0; i < connNum*loopNum; i++ {
			totalDur += <-durChan
			err := <-errChan
			if err != nil {
				errCnt++
			}
		}

		fmt.Println("### Client side ###")
		fmt.Printf("Average letency is %v, errRate is %.1f%%\n",
			totalDur/time.Duration(connNum*loopNum),
			float64(100*errCnt)/float64(connNum*loopNum),
		)
		fmt.Println("### Balancer side ###")
		balancer.GetReport()
		fmt.Printf("Next request lists will come %vs latter.\n", nextInterval)
		time.Sleep(time.Duration(nextInterval) * time.Second)
	}
}

func main() {
	call()
}
