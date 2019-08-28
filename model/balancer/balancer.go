package balancer

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"math"
	"math/rand"
	"sync"
	"time"

	bs "github.com/bytecamp2019d/bustsurvivor/api/bustsurvivor"
	"github.com/bytecamp2019d/bustsurvivor/model/calculator"
)

var IPs = [...]string{
	"127.0.0.1:8080",
	"127.0.0.1:8081",
	//"127.0.0.1:8082",
	//"127.0.0.1:8083",
}

const (
	BEST = iota
	NORMAL
	WEAK
	SHIT
)

const serverNum = len(IPs)
const diffRatio = 0.95

var totalDurs [serverNum]time.Duration
var errCnts [serverNum]int
var hints [serverNum]int

var weights [serverNum]float64

var ctx = context.Background()

type clientPool struct {
	clients  []bs.SurvivalServiceClient
	lastUsed int
}

var connNum = 1

var clientPools [serverNum]clientPool

var calcChan = make(chan calculator.CalcPkg)

func InitBalancer(totalConnNum int, firstTest bool) {
	connNum = totalConnNum
	for i := 0; i < serverNum; i++ {
		totalDurs[i] = 0
		errCnts[i] = 0
		hints[i] = 0
		clientPools[i].clients = nil
		clientPools[i].lastUsed = -1
		if firstTest {
			weights[i] = 100
		}
	}

	initClients(totalConnNum)
	go requestStatistic()
	go weightUpdate(1000)
}

var durationLock sync.RWMutex
var durationRequestCount [serverNum]int64
var durationRequestLatency [serverNum]time.Duration
var durationRequestErrorCount [serverNum]int64
var timeTable = [4][2]time.Duration{
	{time.Millisecond * 0, time.Millisecond * 20},
	{time.Millisecond * 20, time.Millisecond * 50},
	{time.Millisecond * 50, time.Millisecond * 100},
	{time.Millisecond * 100, time.Millisecond * 1000 * 1000},
}
var scoreTable = [4]float64{
	90, 60, 30, 0,
}

func getScore(avgDuration time.Duration, errorCount int64, requestCount int64) float64 {
	if errorCount*100 >= requestCount {
		return 0
	}
	for i := 0; i < 4; i++ {
		if avgDuration < timeTable[i][1] {
			return scoreTable[i]
		}
	}
	return 0
}

func weightUpdate(frequency time.Duration) {
	for {
		durationLock.Lock()
		serverIndexToRebalance := -1
		worstScore := math.MaxFloat64
		totalScore := 0.0
		score := [serverNum]float64{}
		hasFree := false
		for i := 0; i < serverNum; i++ {
			if durationRequestCount[i] == 0 {
				score[i] = scoreTable[BEST]
				hasFree = true
			} else {
				score[i] = getScore(
					durationRequestLatency[i]/time.Duration(durationRequestCount[i]),
					durationRequestErrorCount[i],
					durationRequestCount[i],
				)
			}
			totalScore += score[i]
			if score[i] < worstScore {
				worstScore = score[i]
				serverIndexToRebalance = i
			}
		}
		if worstScore <= scoreTable[NORMAL] || hasFree {
			niubiServers := make([]int, 0)
			totalBalanceScore := 0.0
			for index, s := range score {
				if s > scoreTable[NORMAL] && serverIndexToRebalance != index {
					totalBalanceScore += s
					niubiServers = append(niubiServers, index)
				}
			}

			if len(niubiServers) > 0 {
				weightToRebalance := weights[serverIndexToRebalance] * (1 - diffRatio)
				weights[serverIndexToRebalance] = weights[serverIndexToRebalance] * diffRatio
				for _, serverIndex := range niubiServers {
					weights[serverIndex] += weightToRebalance * (score[serverIndex] / totalBalanceScore)
				}
			}
		}

		for i := 0; i < serverNum; i++ {
			durationRequestCount[i] = 0
			durationRequestLatency[i] = 0
			durationRequestErrorCount[i] = 0
		}

		durationLock.Unlock()
		fmt.Println(weights)
		time.Sleep(frequency * time.Millisecond)
	}
}

func requestStatistic() {
	for i := 0; i < serverNum; i++ {
		durationRequestCount[i] = 0
		durationRequestLatency[i] = 0
		durationRequestErrorCount[i] = 0
	}

	for {
		pkg := <-calcChan
		hints[pkg.Index]++
		durationLock.Lock()
		durationRequestCount[pkg.Index]++
		if pkg.Err {
			durationRequestErrorCount[pkg.Index]++
			errCnts[pkg.Index]++
		}
		durationRequestLatency[pkg.Index] += pkg.Dur
		durationLock.Unlock()
		totalDurs[pkg.Index] += pkg.Dur
	}
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
	judgeList := make([]float64, 0)
	sum := 0.0

	for _, w := range weights {
		sum += w
		//sum += 20  // todo: same weight
		judgeList = append(judgeList, sum)
	}

	rand.Seed(time.Now().UnixNano())
	randNum := rand.Float64() * sum // range in [0, sum)
	bestIndex := 0
	for i := 0; i < len(judgeList); i++ {
		if randNum < judgeList[i] {
			bestIndex = i
			break
		}
	}
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
	calcChan <- calculator.CalcPkg{
		Err:   err != nil,
		Dur:   dur,
		Index: index,
	}
	_ = resp // ignore resp
}
