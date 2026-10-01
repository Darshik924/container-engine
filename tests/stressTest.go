package main

import (
	"fmt"
	"os"
	"time"
)

func main() {
    time.Sleep(time.Second) 
   
	Sample()
    Stress()
}

func Stress()  {
	var chunks [][]byte
    pgSz := os.Getpagesize()

    for alloc := 0; alloc < 80<<20; alloc += 1 << 20 {
        
        chunk := make([]byte, 1<<20)
        for i := 0; i < len(chunk); i += pgSz {
            chunk[i] = 1
        }

        chunks = append(chunks, chunk)

        fmt.Printf("memory at %d MiB\n", alloc+1)
        time.Sleep(100 * time.Millisecond)
    }

    fmt.Println("finished: the 50 MiB limit did not stop this workload")
}