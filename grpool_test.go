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

func TestPool_SubmitFuture(t *testing.T) {
	pool := NewPool(2, 10)

	f := pool.SubmitFuture(func() interface{} {
		time.Sleep(time.Second * 3)
		return "ok"
	})

	fmt.Println(<-f.ResultChan)
}

func TestPool_Submit(t *testing.T) {
	pool := NewPool(2, 10)

	pool.WaitCount(1)
	pool.Submit(func() interface{} {
		defer pool.JobDone()

		time.Sleep(time.Second * 1)

		// 返回值会被忽略
		return nil
	})

	// 等待所有任务完成
	pool.WaitAll()
}
