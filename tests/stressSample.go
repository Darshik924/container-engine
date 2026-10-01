package main

import (
	"fmt"
	"os"
	"time"
)

func main() {
    time.Sleep(time.Second) // Give the parent time to move us into the cgroup.
   
	Sample()
}
func stress()  {
	var chunks [][]byte
    pageSize := os.Getpagesize()

    for allocated := 0; allocated < 80<<20; allocated += 1 << 20 {
        chunk := make([]byte, 1<<20)
        for i := 0; i < len(chunk); i += pageSize {
            chunk[i] = 1
        }
        chunks = append(chunks, chunk)
        fmt.Printf("touched %d MiB\n", allocated+1)
        time.Sleep(100 * time.Millisecond)
    }

    fmt.Println("finished: the 50 MiB limit did not stop this workload")
}