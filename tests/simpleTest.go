package main

import (
	"fmt"
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