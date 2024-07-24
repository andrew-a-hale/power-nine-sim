package main

import (
	"fmt"
	"math"
	"math/rand"
	"sync"
)

const (
	RARES              = 112
	STARTER_RARE_COUNT = 2
	POWER_NINE_RARES   = 9
	TOTAL_RUNS         = 100000
	WORKERS            = 10
)

var POWER_NINE = []string{
	"ANCESTRAL RECALL",
	"BLACK LOTUS",
	"MOX EMERALD",
	"MOX JET",
	"MOX PEARL",
	"MOX RUBY",
	"MOX SAPPHIRE",
	"TIMETWISTER",
	"TIME WALK",
}

func isDone(pulls map[int]int) bool {
	for _, p := range pulls {
		if p == 0 {
			return false
		}
	}

	return true
}

func simulation(wg *sync.WaitGroup, ch chan<- int, show bool) {
	defer wg.Done()

	packs := 0
	pulls := make(map[int]int)
	for i := 0; i < POWER_NINE_RARES; i++ {
		pulls[i] = 0
	}

	for !isDone(pulls) {
		packs++
		for _, p := range rand.Perm(RARES)[0:STARTER_RARE_COUNT] {
			if p < POWER_NINE_RARES {
				pulls[p]++
			}
		}
	}

	if show {
		fmt.Printf("Sample Took: %d packs\n", packs)
		for j, n := range pulls {
			fmt.Printf("%s: %d\n", POWER_NINE[j], n)
		}
		fmt.Println()
	}

	ch <- packs
}

func mean(xs []int) float64 {
	acc := 0
	for _, x := range xs {
		acc += x
	}

	return float64(acc) / float64(len(xs))
}

func variance(xs []int) float64 {
	acc := 0.0
	xbar := mean(xs)
	for _, x := range xs {
		acc += math.Pow(float64(x)-xbar, 2)
	}

	return acc / float64(len(xs)-1)
}

func confidenceInterval(xbar float64, variance float64, n int, z float64) (float64, float64) {
	interval := z * math.Sqrt(variance/float64(n))
	return xbar - interval, xbar + interval
}

func main() {
	var runs []int
	ch := make(chan int, WORKERS)
	sample := int(rand.Int31n(TOTAL_RUNS))

	var wg sync.WaitGroup
	var i int
	for i < TOTAL_RUNS {
		for j := 0; j < WORKERS; j++ {
			wg.Add(1)
			go simulation(&wg, ch, i == sample)

			i++
			if i >= TOTAL_RUNS {
				break
			}
		}

		go func(ch chan int, runs *[]int) {
			for x := range ch {
				*runs = append(*runs, x)
			}
		}(ch, &runs)
		wg.Wait()
	}
	close(ch)

	xbar := mean(runs)
	variance := variance(runs)
	lb, ub := confidenceInterval(xbar, variance, len(runs), 1.96)
	fmt.Printf("Expected Runs (95%% CI): %0.2f (%0.2f, %0.2f)\n", xbar, lb, ub)
}
