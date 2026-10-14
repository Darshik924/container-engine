package main

import (
	"fmt"
	"os"
	"time"
)

// Simple Program that counts the number of 1 bits in 1 to 10 inclu
func Sample()  {
	var nums []int
	for i:=1; i<11; i++ {
		nums = append(nums, i)
	}

	cnt := 0
	for _, i := range nums {
		for k:=31; k>=0; k-- {
			if (1 << k) & i >= 1 {
				cnt++;
			} 
		}
	}
	fmt.Println(cnt)
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

        fmt.Printf("memory at %d MiB\n", (alloc>>20)+1)
        time.Sleep(30 * time.Millisecond)
    }

    fmt.Println("finished: the 50 MiB limit did not stop this workload")
}