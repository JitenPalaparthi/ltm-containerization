package main

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"time"
)

func main() {
	target := envInt("TARGET_MIB", 150)
	step := envInt("STEP_MIB", 10)
	if target < 1 || step < 1 || target > 2048 {
		panic("invalid TARGET_MIB or STEP_MIB")
	}
	var chunks [][]byte
	allocated := 0
	fmt.Printf("Starting memory demo: target=%d MiB step=%d MiB\n", target, step)
	for allocated < target {
		n := step
		if target-allocated < n {
			n = target - allocated
		}
		chunk := make([]byte, n*1024*1024)
		for i := 0; i < len(chunk); i += 4096 {
			chunk[i] = 1
		}
		chunks = append(chunks, chunk)
		allocated += n
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		fmt.Printf("Retained=%d MiB GoHeapAlloc=%d MiB\n", allocated, m.HeapAlloc/1024/1024)
		time.Sleep(700 * time.Millisecond)
	}
	fmt.Println("Target reached; keeping memory resident. Container should remain running if limit is sufficient.")
	for {
		runtime.KeepAlive(chunks)
		time.Sleep(5 * time.Second)
	}
}

func envInt(k string, fallback int) int {
	s := os.Getenv(k)
	if s == "" {
		return fallback
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		panic(fmt.Sprintf("invalid %s: %v", k, err))
	}
	return v
}
