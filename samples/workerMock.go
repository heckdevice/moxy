package main

import (
	"flag"
	"fmt"
	log "github.com/sirupsen/logrus"
	"math/rand"
	"os"
	"time"
)

var (
	args *CLIArgs
)

// WorkerResponse - The result of a workers workitem latency
type WorkerResponse struct {
	workerNumber    int
	workerLatency   time.Duration
	workerStartTime time.Time
}

// CLIArgs - Command line arguments for the main function
type CLIArgs struct {
	workers             int
	simulatedMinLatency float64
	simulatedMaxLatency float64
}

func executionDuration(msg string, startTime time.Time) {
	log.Println(fmt.Sprintf("%v : %v", msg, time.Since(startTime)))
}

func parallelCounter(workerNumber int, countsToFinish int, workerChan chan WorkerResponse) {
	startTime := time.Now()
	log.Info(fmt.Sprintf("Worker number %v started", workerNumber))
	i := 0
	for i <= countsToFinish {
		//Counter job
		i++
	}
	duration := time.Since(startTime)
	resp := WorkerResponse{workerNumber, duration, startTime}
	workerChan <- resp
}

//Provide latencies in ms rounded to next int
func simluateLatencies(workerNumber int, latencyMin, latencyMax float64, workerChan chan WorkerResponse) {
	startTime := time.Now()
	simulatedLatency := latencyMin + rand.Float64()*float64(rand.Intn(int(latencyMax-latencyMin)))
	log.Debug(fmt.Sprintf("Simulate latency is %v", simulatedLatency))
	waitDuration := time.Duration(int(simulatedLatency*1000*1000)) * time.Nanosecond
	log.Debug(fmt.Sprintf("Wait is %v", waitDuration))
	time.Sleep(waitDuration)
	resp := WorkerResponse{workerNumber, waitDuration, startTime}
	workerChan <- resp
}

func init() {
	args = &CLIArgs{}
	flag.IntVar(&args.workers, "w", 10, "number of workers to simulate")
	flag.Float64Var(&args.simulatedMinLatency, "minLat", 1.234, "minimum latency to simulate")
	flag.Float64Var(&args.simulatedMaxLatency, "maxLat", 10.234, "maximum sqlatency to simulate")
}

func main() {
	log.SetOutput(os.Stdout)
	log.SetOutput(os.Stderr)
	log.SetFormatter(&log.JSONFormatter{})
	flag.Parse()
	if args.simulatedMaxLatency <= args.simulatedMinLatency {
		log.Fatal(fmt.Sprintf("Invalid argument values :: simulatedMinLatency:%v, simulatedMaxLatency:%v", args.simulatedMinLatency, args.simulatedMaxLatency))
		return
	}
	workersChan := make(chan WorkerResponse)
	jobsDone := 0
	fmt.Println(fmt.Sprintf("Starting %v workers with min latecy = %v, max latency = %v", args.workers, args.simulatedMinLatency, args.simulatedMaxLatency))
	startTime := time.Now()
	defer executionDuration(fmt.Sprintf("--------All %v workers returned in", args.workers), startTime)
	workerNumber := 1
	for workerNumber <= args.workers {
		//go parallelCounter(workerNumber, counts, workersChan)
		go simluateLatencies(workerNumber, args.simulatedMinLatency, args.simulatedMaxLatency, workersChan)
		workerNumber++
	}
	maxLatencyWorker := WorkerResponse{-1, time.Duration(0 * time.Millisecond), time.Now()}
	for {
		select {
		case resp := <-workersChan:
			jobsDone++
			executionDuration(fmt.Sprintf("worker #%v returned with simulated latency of %v in ", resp.workerNumber, resp.workerLatency), resp.workerStartTime)
			if resp.workerLatency > maxLatencyWorker.workerLatency {
				maxLatencyWorker = resp
			}
			if jobsDone == args.workers {
				fmt.Println("****************************************************")
				log.Info(fmt.Sprintf("worker #%v was last worker to return with latency %v", resp.workerNumber, resp.workerLatency))
				log.Info(fmt.Sprintf("worker #%v returned with highest latency of %v", maxLatencyWorker.workerNumber, maxLatencyWorker.workerLatency))
				fmt.Println("****************************************************")
				return
			}
		default:
			//fmt.Println(fmt.Sprintf("Waiting for workers to return, till now %v returned", jobsDone))
		}
	}
}
