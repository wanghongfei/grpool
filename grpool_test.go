package grpool

import (
	"runtime"
	"testing"
	"fmt"
	"time"
)

func init() {
	println("using MAXPROC")
	numCPUs := runtime.NumCPU()
	runtime.GOMAXPROCS(numCPUs)
}

func TestPool_Submit(t *testing.T) {
	pool := NewPool(2, 10)

	f := pool.Submit(func() interface{} {
		time.Sleep(time.Second * 3)
		return "ok"
	})

	fmt.Println(<-f.ResultChan)
}
