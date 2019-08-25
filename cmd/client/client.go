package main

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"google.golang.org/grpc"

	bs "github.com/bytecamp2019d/bustsurvivor/api/bustsurvivor"
)

const (
	connNum    = 30
	rps        = 100
	loopNum    = 500
	requestNum = 10 // 测试总请求次数
)

type calcPkg struct {
	err   bool
	dur   time.Duration
	index int
}

type connectionPool struct {
	connections [connNum] *grpc.ClientConn
	lastUsed    int
}

var IPs = [...] string{
	"127.0.0.1:8080",
	"127.0.0.1:8081",
	//"127.0.0.1:8082",
	//"127.0.0.1:8083",
}

const serverNum = len(IPs)

var hasChanged = false

var weights [len(IPs)] float64

var connectionPools [serverNum]connectionPool

var rwLock sync.RWMutex

func initConnections() {
	for i := 0; i < serverNum; i++ {
		for j := 0; j < connNum; j++ {
			conn, err := grpc.Dial(IPs[i], grpc.WithInsecure())
			if err != nil {
				panic(err)
			}
			connectionPools[i].connections[j] = conn
		}
		connectionPools[i].lastUsed = -1
	}
}

func getConnection() (*grpc.ClientConn, int) {
	if !hasChanged {
		rand.Seed(time.Now().Unix()) // initialize global pseudo random generator
		index := rand.Intn(serverNum)
		connIndex := connectionPools[index].lastUsed
		connIndex = (connIndex + 1) % connNum
		connectionPools[index].lastUsed = connIndex
		return connectionPools[index].connections[connIndex], index
	}

	rwLock.RLock()
	bestIndex := 0
	bestWeight := weights[bestIndex]

	for i, weight := range weights {
		if weight < bestWeight {
			bestWeight = weight
			bestIndex = i
		}
	}
	rwLock.RUnlock()
	connIndex := connectionPools[bestIndex].lastUsed
	connIndex = (connIndex + 1) % connNum
	connectionPools[bestIndex].lastUsed = connIndex
	return connectionPools[bestIndex].connections[connIndex], bestIndex
}

func sendRequest(durChans [serverNum]chan time.Duration, errChans [serverNum]chan error, calcChan chan calcPkg) {
	conn, index := getConnection()
	client := bs.NewSurvivalServiceClient(conn)
	req := &bs.BustSurvivalRequest{
		CardsToPick:   10,
		BustThreshold: 80,
	}
	ctx := context.Background()
	begin := time.Now()
	resp, err := client.BustSurvival(ctx, req)
	dur := time.Since(begin)
	durChans[index] <- dur
	errChans[index] <- err
	calcChan <- calcPkg{
		err:   err != nil,
		dur:   dur,
		index: index,
	}
	_ = resp // ignore resp
}

func calculateWeight(pkg calcPkg) {
	// calculate new weights with calcPkg
	// sync lock?
	rwLock.Lock()
	// todo: modify weights
	/* mock start */
	weights[pkg.index] = pkg.dur.Seconds()
	hasChanged = true
	/* mock end */
	rwLock.Unlock()
}

// call .
func call() {
	initConnections()
	defer func() {
		for _, connPool := range connectionPools {
			for _, conn := range connPool.connections {
				if conn != nil {
					conn.Close()
				}
			}

		}
	}()
	var totalDur [serverNum]time.Duration
	var hints [serverNum]int
	var durChans [serverNum]chan time.Duration
	var errChans [serverNum]chan error
	for i := 0; i < serverNum; i += 1 {
		durChans[i] = make(chan time.Duration, connNum*loopNum)
		errChans[i] = make(chan error, connNum*loopNum)
	}
	calcChan := make(chan calcPkg, connNum*loopNum)
	var errCnt [serverNum]int
	for r := 0; r < requestNum; r++ {
		requestInterval := time.Second * time.Duration(rand.Intn(20))
		fmt.Printf("Next request lists will come %v latter.\n", requestInterval)
		time.Sleep(requestInterval)
		fmt.Println("Come a list of requests!")
		for i := 0; i < connNum; i++ {
			go func() {
				for j := 0; j < loopNum; j++ {
					go sendRequest(durChans, errChans, calcChan)
					time.Sleep(time.Second / rps)
				}
			}()
			time.Sleep(time.Second / rps / connNum)
		}

		for i := 0; i < connNum*loopNum; i++ {
			calcPkg := <-calcChan
			totalDur[calcPkg.index] += <-durChans[calcPkg.index]
			err := <-errChans[calcPkg.index]
			if err != nil {
				errCnt[calcPkg.index]++
			}
			hints[calcPkg.index] ++
			calculateWeight(calcPkg)
		}
	}

	for i := 0; i < serverNum; i += 1 {
		fmt.Printf("Hint %d server %d times\n", i, hints[i])
		fmt.Printf("Average letency is %v, errRate is %.1f%%\n", totalDur[i]/time.Duration(hints[i]), float64(100*errCnt[i])/float64(hints[i]))
	}
}

func main() {
	call()
	fmt.Printf("Finally weights: %v \n", weights)
}
