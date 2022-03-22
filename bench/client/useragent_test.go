package client_test

import (
	"regexp"
	"sync"
	"testing"

	"github.com/isucon10-qualify/isucon10-qualify/bench/client"
)

func Test_UserAgent(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < 10000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ua := client.GenerateUserAgent()
		}()
	}

	wg.Wait()
}
