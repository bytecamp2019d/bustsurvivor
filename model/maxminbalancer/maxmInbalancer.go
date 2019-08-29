package maxminbalancer

import (
	"context"
	"fmt"
	//"golang.org/x/text/unicode/rangetable"
	"google.golang.org/grpc"
	"math"
	"math/rand"
	"sync"
	"time"

	bs "github.com/bytecamp2019d/bustsurvivor/api/bustsurvivor"
	"github.com/bytecamp2019d/bustsurvivor/model/calculator"
)

var IPs = [...]string{
	"10.108.18.57:8080",
	"10.108.18.57:8081",
	"10.108.18.134:8080",
	"10.108.18.135:8080",
}

const (
	BEST = iota
	NORMAL
	WEAK
	SHIT
)

const serverNum = len(IPs)

var totalDurs [serverNum]time.Duration
var errCnts [serverNum]int
var hints [serverNum]int

var weights [serverNum]float64

var COUNT = 0
var weightSS [10][10]float64

var ctx = context.Background()

type clientPool struct {
	clients  []bs.SurvivalServiceClient
	lastUsed int
}

var connNum = 1

var clientPools [serverNum]clientPool

var calcChan = make(chan calculator.CalcPkg)

func InitBalancer(totalConnNum int, isFirst bool) {
	calculator.InitCalculator()
	fmt.Println(isFirst)
	connNum = totalConnNum
	for i := 0; i < serverNum; i++ {
		weights[i] = 100
		totalDurs[i] = 0
		errCnts[i] = 0
		hints[i] = 0
		clientPools[i].clients = nil
		clientPools[i].lastUsed = -1
	}
	initClients(totalConnNum)
	go requestStatistic()
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

func serverCMP(x int, y int) int {
	errorrateX := durationRequestErrorCount[x] * 1.0 / durationRequestCount[x]
	errorrateY := durationRequestErrorCount[y] * 1.0 / durationRequestCount[y]
	if errorrateX < errorrateY {
		return 1
	}
	if errorrateX > errorrateY {
		return -1
	}
	avgLatencyX := durationRequestLatency[x] * 1.0 / time.Duration(durationRequestCount[x])
	avgLatencyY := durationRequestLatency[y] * 1.0 / time.Duration(durationRequestCount[y])
	if avgLatencyX < avgLatencyY {
		return 1
	}
	if avgLatencyX > avgLatencyY {
		return -1
	}
	return 0
}
func updateWeight() {
	sum := 0.0
	var score []float64
	for i := 0; i < serverNum; i++ {
		var tmp [][]float64
		for j := 0; j < COUNT; j++ {
			tmp = append(tmp, []float64{weightSS[i][j+1] - weightSS[i][j], 0})
		}
		//fmt.Println(tmp)
		tt := calculator.GetRes(tmp, COUNT-1, 9)
		score = append(score, tt)
		sum += tt
	}
	for i := 0; i < serverNum; i++ {
		weights[i] *= (1.0 - score[i]/sum)
	}

}
func weightUpdateLittle() {
	for i := 0; i < serverNum; i++ {
		if durationRequestCount[i] < 100 {
			return
		}
	}
	for i := 0; i < serverNum; i++ {
		fmt.Println("Server ==========================", i)
		fmt.Print("处理请求数量：   ")
		fmt.Println(durationRequestCount[i])
		fmt.Print("平均时延： ")
		fmt.Println(durationRequestLatency[i])
	}
	for i := 0; i < serverNum; i++ {
		fmt.Print(weights[i], "   ")
	}
	fmt.Println()
	fmt.Println("==============================")
	bestServer := 0
	worstServer := 0
	for i := 1; i < serverNum; i++ {
		if serverCMP(worstServer, i) == 1 {
			worstServer = i
		}
		if serverCMP(i, bestServer) == 1 {
			bestServer = i
		}
	}
	var diffRatio float64
	if durationRequestErrorCount[worstServer] > 0 || durationRequestErrorCount[bestServer] > 0 {
		diffRatio = 0.05
	} else {
		avgLatencyBAD := durationRequestLatency[worstServer] * 1.0 / time.Duration(durationRequestCount[worstServer])
		avgLatencyGOOD := durationRequestLatency[bestServer] * 1.0 / time.Duration(durationRequestCount[bestServer])
		diffRatio = math.Abs((avgLatencyBAD - avgLatencyGOOD).Seconds() / (avgLatencyBAD + avgLatencyGOOD).Seconds())
	}
	if diffRatio > 0.05 {
		diffRatio = 0.05
	}
	weights[bestServer] += weights[worstServer] * diffRatio
	weights[worstServer] = weights[worstServer] * (1 - diffRatio)
	for i := 0; i < serverNum; i++ {
		fmt.Print(" ", durationRequestLatency[i]/time.Duration(durationRequestCount[i]))
	}
	for i := 0; i < serverNum; i++ {
		weightSS[COUNT][i] = (durationRequestLatency[i] / time.Duration(durationRequestCount[i])).Seconds() * 1000
	}
	for i := 0; i < serverNum; i++ {
		durationRequestCount[i] = 0
		durationRequestLatency[i] = 0
		durationRequestErrorCount[i] = 0
		time.Duration(22).Nanoseconds()
	}

	if COUNT == 8 {
		updateWeight()
		COUNT = 0
		return
	}
	COUNT++

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
		durationRequestCount[pkg.Index]++
		if pkg.Err {
			durationRequestErrorCount[pkg.Index]++
			errCnts[pkg.Index]++
		}
		durationRequestLatency[pkg.Index] += pkg.Dur
		totalDurs[pkg.Index] += pkg.Dur
		weightUpdateLittle()
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
	sum1 := 0.0
	rand.Seed(time.Now().UnixNano())
	for _, w := range weights {
		sum1 += w
	}
	randNum := rand.Float64() * sum1 // range in [0, sum)
	bestIndex := 0
	sum := 0.0
	for i, w := range weights {
		sum += w
		if randNum < sum {
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
