package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/jrboelens/multilimiter"
)

func DoWork(lim multilimiter.Limiter, iterations, sleepMs int) {
	tracker := &multilimiter.ConcurrencyTracker{}
	tracker.Start()
	for i := 0; i < iterations; i++ {
		lim.Execute(context.Background(), func(ctx context.Context) {
			tracker.Add()
			if sleepMs > 0 {
				time.Sleep(time.Duration(sleepMs) * time.Millisecond)
			}
			tracker.Subtract()
		})
	}
	lim.Wait()
	tracker.Stop()
	printTracker(tracker)
}

func printTracker(tracker *multilimiter.ConcurrencyTracker) {
	fmt.Printf("Elapsed Time: %dms\n", tracker.Elapsed().Milliseconds())
	fmt.Printf("Total iterations: %d\n", tracker.Total())
	fmt.Printf("Max Concurrency: %d\n", tracker.Max())
	fmt.Printf("Rate: %f\n", tracker.Rate())
}

func main() {
	var rate float64
	var concurrency, iterations, sleepMs int

	flag.Float64Var(&rate, "rate", 1.0, "rate limit")
	flag.IntVar(&concurrency, "concurrency", 1, "concurrency limit")
	flag.IntVar(&iterations, "iterations", 1, "number of iterations")
	flag.IntVar(&sleepMs, "sleepMs", 1, "number of milliseconds to sleep during each Execute()")
	flag.Parse()

	rateOpt := &multilimiter.RateLimitOption{multilimiter.NewRateLimiter(rate)}
	concOpt := &multilimiter.ConcLimitOption{multilimiter.NewConcLimiter(concurrency)}
	lim := multilimiter.NewLimiter(rateOpt, concOpt)

	// Default Limiter offers a more easy way to create a Limiter
	//lim := multilimiter.DefaultLimiter(rate, concurrency)

	fmt.Printf("Starting Limiter for %d iterations at %f/s with %d concurrency\n", iterations, rate, concurrency)
	DoWork(lim, iterations, sleepMs)
	fmt.Printf("\n")
}
