// cpu-stress-lab runs a CPU-bound workload on a chosen number of Go workers.
// It is a load demonstration, not a benchmark. Stop it with Ctrl+C.
package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

var (
	workersFlag  = flag.Int("workers", envInt("WORKERS", runtime.GOMAXPROCS(0)), "number of CPU workers")
	durationFlag = flag.Duration("duration", envDuration("DURATION", 30*time.Second), "run time; 0 runs until stopped (examples: 30s, 2m)")
)

var batches uint64
var checksum uint64

func main() {
	flag.Parse()
	if *workersFlag < 1 {
		fmt.Fprintln(os.Stderr, "workers must be at least 1")
		os.Exit(2)
	}
	if *durationFlag < 0 {
		fmt.Fprintln(os.Stderr, "duration cannot be negative")
		os.Exit(2)
	}

	visibleCPUs := runtime.NumCPU()
	runtime.GOMAXPROCS(*workersFlag)
	fmt.Printf("CPU Stress Lab | OS/arch: %s/%s | visible CPUs: %d | Go workers: %d | GOMAXPROCS: %d\n",
		runtime.GOOS, runtime.GOARCH, visibleCPUs, *workersFlag, runtime.GOMAXPROCS(0))
	fmt.Println("This workload intentionally uses CPU. Press Ctrl+C to stop early.")
	if *workersFlag > visibleCPUs {
		fmt.Printf("Note: %d workers exceed the %d CPUs visible to the process; the OS will time-slice them.\n", *workersFlag, visibleCPUs)
	}

	stop := make(chan struct{})
	var stopOnce sync.Once
	stopWorkers := func() { stopOnce.Do(func() { close(stop) }) }

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(signalCh)

	var wg sync.WaitGroup
	wg.Add(*workersFlag)
	for id := 0; id < *workersFlag; id++ {
		go func(workerID int) {
			defer wg.Done()
			state := uint64(workerID+1) * 0x9e3779b97f4a7c15
			for {
				select {
				case <-stop:
					return
				default:
				}
				state = work(state, 200_000)
				atomic.AddUint64(&checksum, state)
				atomic.AddUint64(&batches, 1)
			}
		}(id)
	}

	start := time.Now()
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	var deadline <-chan time.Time
	var timer *time.Timer
	if *durationFlag > 0 {
		timer = time.NewTimer(*durationFlag)
		defer timer.Stop()
		deadline = timer.C
	}

	previous := uint64(0)
	for {
		select {
		case <-ticker.C:
			current := atomic.LoadUint64(&batches)
			fmt.Printf("elapsed=%s batches=%d (+%d last second)\n",
				time.Since(start).Truncate(time.Second), current, current-previous)
			previous = current
		case <-deadline:
			stopWorkers()
			wg.Wait()
			fmt.Println("Duration reached.")
			printSummary(start, *workersFlag)
			return
		case <-signalCh:
			stopWorkers()
			wg.Wait()
			fmt.Println("Stopped by signal.")
			printSummary(start, *workersFlag)
			return
		}
	}
}

func printSummary(start time.Time, workers int) {
	fmt.Printf("Finished | workers=%d | elapsed=%s | batches=%d | checksum=%d\n",
		workers, time.Since(start).Truncate(time.Second), atomic.LoadUint64(&batches), atomic.LoadUint64(&checksum))
}

//go:noinline
func work(x uint64, n int) uint64 {
	for i := 0; i < n; i++ {
		x ^= x << 13
		x ^= x >> 7
		x ^= x << 17
	}
	return x
}

func envInt(name string, fallback int) int {
	value := os.Getenv(name)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid %s=%q: %v\n", name, value, err)
		os.Exit(2)
	}
	return parsed
}

func envDuration(name string, fallback time.Duration) time.Duration {
	value := os.Getenv(name)
	if value == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid %s=%q: %v\n", name, value, err)
		os.Exit(2)
	}
	return parsed
}
