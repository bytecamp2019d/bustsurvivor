package balancer

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"math/rand"
	"sync"
	"time"

	bs "github.com/bytecamp2019d/bustsurvivor/api/bustsurvivor"
)

var IPs = [...] string{
	"127.0.0.1:8080",
	"127.0.0.1:8081",
	//"127.0.0.1:8082",
	//"127.0.0.1:8083",
}

const serverNum = len(IPs)

var totalDurs [serverNum]time.Duration
var errCnts [serverNum]int
var hints [serverNum]int

var weights [serverNum] float64

var ctx = context.Background()

type clientPool struct {
	clients  []bs.SurvivalServiceClient
	lastUsed int
}

var connNum = 1

var hasChanged = false
var rwLock sync.RWMutex
var clientPools [serverNum]clientPool

type calcPkg struct {
	err   bool
	dur   time.Duration
	index int
}

var calcChan = make(chan calcPkg)


func InitBalancer(totalConnNum int) {
	connNum = totalConnNum
	for i := 0; i < serverNum; i++ {
		totalDurs[i] = 0
		errCnts[i] = 0
		hints[i] = 0
		clientPools[i].clients = nil
		clientPools[i].lastUsed = -1
	}
	initClients(totalConnNum)
	go checkWeight()
}

func GetReport() {
	for i := 0; i < serverNum; i += 1 {
		fmt.Printf("Hint %s server %d times\n", IPs[i], hints[i])
		if hints[i] > 0 {
			fmt.Printf(
				"Average letency is %v, errRate is %.1f%%\n",
				totalDurs[i]/time.Duration(hints[i]),
				float64(100*errCnts[i])/float64(hints[i]),
			)
		}
	}
	fmt.Printf("Finally weights: %v\n", weights)
}
func checkWeight() {
	for {
		pkg := <-calcChan
		totalDurs[pkg.index] += pkg.dur
		hints[pkg.index] += 1
		if pkg.err {
			errCnts[pkg.index] += 1
		}
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

}

func initClients(totalConnNum int) {
	for i := 0; i < serverNum; i++ {
		for j := 0; j < totalConnNum; j++ {
			conn, err := grpc.Dial(IPs[i], grpc.WithInsecure())
			if err != nil {
				panic(err)
			}
			client := bs.NewSurvivalServiceClient(conn)
			clientPools[i].clients = append(clientPools[i].clients, client)
		}
		clientPools[i].lastUsed = -1
	}
}

func getClient() (bs.SurvivalServiceClient, int) {
	if !hasChanged {
		rand.Seed(time.Now().Unix()) // initialize global pseudo random generator
		index := rand.Intn(serverNum)
		connIndex := clientPools[index].lastUsed
		connIndex = (connIndex + 1) % connNum
		clientPools[index].lastUsed = connIndex
		return clientPools[index].clients[connIndex], index
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
	connIndex := clientPools[bestIndex].lastUsed
	connIndex = (connIndex + 1) % connNum
	clientPools[bestIndex].lastUsed = connIndex
	return clientPools[bestIndex].clients[connIndex], bestIndex
}
func SendRequest(req *bs.BustSurvivalRequest, durChan chan time.Duration, errChan chan error) {
	client, index := getClient()

	begin := time.Now()
	resp, err := client.BustSurvival(ctx, req)
	dur := time.Since(begin)

	durChan <- dur
	errChan <- err
	calcChan <- calcPkg{
		err:   err != nil,
		dur:   dur,
		index: index,
	}
	_ = resp // ignore resp
}
