package stage

import "sync"

// runIndexedParallel executes fn for indices [0,n) using a worker pool and
// returns all results in completion order.
func runIndexedParallel[T any](n, workers int, fn func(int) T) []T {
	jobs := make(chan int)
	results := make(chan T)
	var wg sync.WaitGroup

	worker := func() {
		defer wg.Done()
		for idx := range jobs {
			results <- fn(idx)
		}
	}

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go worker()
	}

	go func() {
		for i := 0; i < n; i++ {
			jobs <- i
		}
		close(jobs)
	}()

	out := make([]T, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, <-results)
	}
	wg.Wait()
	return out
}
