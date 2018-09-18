# grpool
[![Build Status](https://travis-ci.org/ivpusic/grpool.svg?branch=master)](https://travis-ci.org/ivpusic/grpool)

Lightweight Goroutine pool

Clients can submit jobs. Dispatcher takes job, and sends it to first available worker.
When worker is done with processing job, will be returned back to worker pool.

Number of workers and Job queue size is configurable.

## Docs
https://godoc.org/github.com/ivpusic/grpool

## Installation
```
go get github.com/wanghongfei/grpool
```



## 添加Future机制

```go
pool := NewPool(2, 10)

f := pool.SubmitFuture(func() interface{} {
    time.Sleep(time.Second * 3)
    return "ok"
})

// block直到任务完成
fmt.Println(<-f.ResultChan)
```



## Example with waiting jobs to finish
```Go
package main

import (
  "fmt"
  "runtime"

  "github.com/ivpusic/grpool"
)

func main() {
  // number of workers, and size of job queue
  pool := grpool.NewPool(100, 50)
  defer pool.Release()

  // how many jobs we should wait
  pool.WaitCount(10)

  // submit one or more jobs to pool
  for i := 0; i < 10; i++ {
    count := i

    pool.Submit(func() interface{} {
		defer pool.JobDone()

		time.Sleep(time.Second * 1)

		// 返回值会被忽略
		return nil
	})
  }

  // wait until we call JobDone for all jobs
  pool.WaitAll()
}
```





## License

*MIT*
