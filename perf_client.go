package perf_client

import (
	"fmt"
	"math"
	"time"

	"github.com/golang/glog"
)

type Task func() error

type Metric struct {
	Duration    time.Duration
	MaxResponse time.Duration
	MinResponse time.Duration
	NumErrors   int
	NumRequests int
	NumRoutines int
}

func runtask(fn Task, testduration time.Duration) Metric {
	m := Metric{MaxResponse: time.Duration(math.MinInt64), MinResponse: time.Duration(math.MaxInt64)}
	var err error
	var funcduration time.Duration
	start := time.Now()
	var requestTime time.Time
	for time.Now().Sub(start) < testduration {
		requestTime = time.Now()
		if err = fn(); err != nil {
			m.NumErrors++
			continue
		}
		funcduration = time.Now().Sub(requestTime)
		m.Duration += funcduration
		m.NumRequests++
		if m.MaxResponse < funcduration {
			m.MaxResponse = funcduration
		}
		if m.MinResponse > funcduration {
			m.MinResponse = funcduration
		}
	}
	return m
}

func RunPerfTest(fn Task, testduration time.Duration, numroutines int) Metric {
	glog.Infoln("------ Running Performance Test ------")
	defer glog.Infoln("------ Finished Performance Test ------")
	metrics := make(chan Metric)
	for i := 0; i < numroutines; i++ {
		go func() { metrics <- runtask(fn, testduration) }()
	}
	var globalmetrics = Metric{MaxResponse: time.Duration(math.MinInt64), MinResponse: time.Duration(math.MaxInt64)}
	for i := 0; i < numroutines; i++ {
		metric := <-metrics
		globalmetrics.Duration += metric.Duration
		globalmetrics.NumErrors += metric.NumErrors
		globalmetrics.NumRequests += metric.NumRequests
		if globalmetrics.MaxResponse < metric.MaxResponse {
			globalmetrics.MaxResponse = metric.MaxResponse
		}
		if globalmetrics.MinResponse > metric.MinResponse {
			globalmetrics.MinResponse = metric.MinResponse
		}
		globalmetrics.NumRoutines++
	}
	glog.Infoln("Performance Results: \n", globalmetrics)
	return globalmetrics
}

func (metric Metric) AverageDurationSecs() float64 {
	return metric.Duration.Seconds() / float64(metric.NumRoutines)
}

func (metric Metric) AverageRequestRate() float64 {
	return float64(metric.NumRequests) / metric.AverageDurationSecs()
}

func (metric Metric) AverageRequestTime() float64 {
	return metric.Duration.Seconds() / float64(metric.NumRequests)
}

func (metric Metric) String() string {
	return fmt.Sprintln("Num. Routines:", metric.NumRoutines, "\nNum. Errors:", metric.NumErrors, "\nAvg. Request Rate:", metric.AverageRequestRate(), "request/s\nAvg. Request Time:", time.Duration(float64(int64(time.Second))*metric.AverageRequestTime()), "\nFastest Response:", metric.MinResponse, "\nSlowest Response:", metric.MaxResponse, "\nAvg: Duration:", time.Duration(metric.AverageDurationSecs()*float64(int64(time.Second))))
}
