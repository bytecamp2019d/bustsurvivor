package main

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"

	bs "github.com/bytecamp2019d/bustsurvivor/api/bustsurvivor"
)

const (
	connNum = 3
	rps     = 100
	loopNum = 500
)

// call .
func call() {
	target := "127.0.0.1:8080"
	var totalDur time.Duration
	durChan := make(chan time.Duration, connNum*loopNum)
	errChan := make(chan error, connNum*loopNum)
	for i := 0; i < connNum; i++ {
		go func() {
			conn, err := grpc.Dial(target, grpc.WithInsecure())
			if err != nil {
				panic(err)
			}
			defer conn.Close()
			client := bs.NewSurvivalServiceClient(conn)
			ctx := context.Background()
			for j := 0; j < loopNum; j++ {
				go func() {
					req := &bs.BustSurvivalRequest{
						CardsToPick:   10,
						BustThreshold: 80,
					}
					begin := time.Now()
					resp, err := client.BustSurvival(ctx, req)
					durChan <- time.Since(begin)
					errChan <- err
					_ = resp // ignore resp
				}()
				time.Sleep(time.Second / rps)
			}
		}()
		time.Sleep(time.Second / rps / connNum)
	}
	errCnt := 0
	for i := 0; i < connNum*loopNum; i++ {
		totalDur += <-durChan
		err := <-errChan
		if err != nil {
			errCnt++
		}
	}
	fmt.Printf("Average letency is %v, errRate is %.1f%%\n", totalDur/connNum/loopNum, float64(100*errCnt)/float64(connNum*loopNum))
}

func main() {
	call()
}
