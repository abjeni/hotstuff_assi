package fuzz

// this file includes a couple of extra unnecessary tests
// used for propability tests, profiling and other things

import (
	"fmt"
	"testing"
)

func TestFrequencyErrorFuzz(t *testing.T) {

	frequency := make(map[string]int, 0)

	f := initFuzz()
	for j := 0; j < 1000; j++ {
		errorInfo := new(ErrorInfo)
		errorInfo.Init()

		iterations := 1

		for i := 0; i < iterations; i++ {
			fuzzMessage := createFuzzMessage(f, errorInfo, nil)
			useFuzzMessage(t, errorInfo, fuzzMessage, nil)
		}

		for key := range errorInfo.panics {
			frequency[key]++
		}
	}

	sum := 0

	for key, val := range frequency {
		sum += val
		fmt.Println(key)
		fmt.Println(val)
		fmt.Println()
	}

	fmt.Println(sum)

}
