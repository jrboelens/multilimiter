package main

import (
	"context"
	"fmt"
	"time"

	"github.com/jrboelens/multilimiter"
)

func DoWork(lim multilimiter.Limiter, iterations, sleepMs int) {
	tracker := &multilimiter.ConcurrencyTracker{}
	tracker.Start()
	for i := 0; i < iterations; i++ {
		lim.Execute(func(ctx context.Context) {
			tracker.Add()
			time.Sleep(time.Duration(sleepMs) * time.Millisecond)
			tracker.Subtract()
		}, context.Background())
	}
	tracker.Stop()
	printTracker(tracker)
}

func RunRateSuite() {
	rate := 10.00 // in seconds
	concurrency := 2
	iterations := 101
	sleepMs := 0

	RunRateLimiter(rate, concurrency, iterations, sleepMs)
	RunRateLimiter(rate*10, concurrency, iterations*10, sleepMs)
	RunRateLimiter(rate*100, concurrency, iterations*100, sleepMs)
}

func RunRateLimiter(rate float64, concurrency, iterations, sleepMs int) {
	rateOpt := &multilimiter.RateLimitOption{multilimiter.NewRateLimiter(rate)}
	lim := multilimiter.NewLimiter(rateOpt)
	fmt.Printf("Starting RateLimiter %d iterations at %f/s with %d concurrency\n", iterations, rate, concurrency)
	DoWork(lim, iterations, sleepMs)
	fmt.Printf("\n")
}

func RunConcSuite() {
	rate := 1000.00 // in seconds
	concurrency := 1
	iterations := 10010
	sleepMs := 10

	RunConcLimiter(rate, concurrency, iterations, sleepMs)
	RunConcLimiter(rate, concurrency*2, iterations, sleepMs)
	RunConcLimiter(rate, concurrency*4, iterations, sleepMs)
	RunConcLimiter(rate, concurrency*8, iterations, sleepMs)
	RunConcLimiter(rate, concurrency*16, iterations, sleepMs)
}

func RunConcLimiter(rate float64, concurrency, iterations, sleepMs int) {
	// rateOpt := &multilimiter.RateLimitOption{multilimiter.NewRateLimiter(rate)}
	// concOpt := &multilimiter.ConcLimitOption{multilimiter.NewConcLimiter(concurrency)}
	lim := multilimiter.DefaultLimiter(rate, concurrency)
	fmt.Printf("Starting RateLimiter %d iterations at %f/s with %d concurrency\n", iterations, rate, concurrency)
	DoWork(lim, iterations, sleepMs)
	fmt.Printf("\n")
}

func printTracker(tracker *multilimiter.ConcurrencyTracker) {
	fmt.Printf("Elapsed Time: %dms\n", tracker.Elapsed().Milliseconds())
	fmt.Printf("Total iterations: %d\n", tracker.Total())
	fmt.Printf("Max Concurrency: %d\n", tracker.Max())
	fmt.Printf("Rate: %f\n", tracker.Rate())
}

func main() {
	//	RunRateSuite()
	RunConcSuite()
}
