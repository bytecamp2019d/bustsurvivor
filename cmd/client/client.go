package main

import (
	"fmt"
	bs "github.com/bytecamp2019d/bustsurvivor/api/bustsurvivor"
	"github.com/bytecamp2019d/bustsurvivor/model/balancer"
	"math/rand"
	"time"
)

const (
	connNum    = 30
	rps        = 100
	loopNum    = 500
	requestNum = 10 // 测试总请求次数
)

// call .
func call() {
	balancer.InitBalancer(connNum)

	totalDur := time.Duration(0)
	durChan := make(chan time.Duration)
	errChan := make(chan error)
	errCnt := 0
	for r := 0; r < requestNum; r++ {
		requestInterval := time.Second * time.Duration(rand.Intn(20))
		fmt.Printf("Next request lists will come %v latter.\n", requestInterval)
		time.Sleep(requestInterval)
		fmt.Println("Come a list of requests!")
		for i := 0; i < connNum; i++ {
			go func() {
				for j := 0; j < loopNum; j++ {
					req := &bs.BustSurvivalRequest{
						CardsToPick:   10,
						BustThreshold: 80,
					}
					go balancer.SendRequest(req, durChan, errChan)
					time.Sleep(time.Second / rps)
				}
			}()
			time.Sleep(time.Second / rps / connNum)
		}

		for i := 0; i < connNum*loopNum; i++ {
			totalDur += <-durChan
			err := <-errChan
			if err != nil {
				errCnt++
			}
		}
	}

	fmt.Printf("Average letency is %v, errRate is %.1f%%\n",
		totalDur/(connNum*loopNum*requestNum),
		float64(100*errCnt)/float64(connNum*loopNum*requestNum),
	)
	balancer.GetReporter()

}

func main() {
	call()
}
