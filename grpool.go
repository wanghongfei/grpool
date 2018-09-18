package grpool

import "sync"

// Gorouting instance which can accept client jobs
type worker struct {
	workerPool 				chan *worker

	// 带有返回值的任务队列
	jobChanWithFuture		chan *Future
	// 无返回值的任务队列
	jobChan					chan Job

	stop					chan struct{}
}

func (w *worker) start() {
	go func() {
		var jobFuture *Future
		for {
			// worker free, add it to pool
			w.workerPool <- w

			select {
			case jobFuture = <-w.jobChanWithFuture:
				// 执行job
				result := jobFuture.job()
				// 将结果放到Future中
				jobFuture.ResultChan <- result

			case jobFunc := <- w.jobChan:
				jobFunc()

				// job()
			case <-w.stop:
				w.stop <- struct{}{}
				return
			}
		}
	}()
}

func newWorker(pool chan *worker) *worker {
	return &worker{
		workerPool: pool,

		jobChanWithFuture: make(chan *Future),
		jobChan: make(chan Job),

		stop: make(chan struct{}),
	}
}

// Accepts jobs from clients, and waits for first free worker to deliver job
type dispatcher struct {
	workerPool 				chan *worker
	jobQueueWithFuture   	chan *Future
	jobQueue				chan Job
	stop       				chan struct{}
}

func (d *dispatcher) dispatch() {
	for {
		select {
		case job := <-d.jobQueueWithFuture:
			worker := <-d.workerPool
			worker.jobChanWithFuture <- job

		case jobFunc := <- d.jobQueue:
			worker := <- d.workerPool
			worker.jobChan <- jobFunc

		case <-d.stop:
			for i := 0; i < cap(d.workerPool); i++ {
				worker := <-d.workerPool

				worker.stop <- struct{}{}
				<-worker.stop
			}

			d.stop <- struct{}{}
			return
		}
	}
}

func newDispatcher(workerPool chan *worker, jobQueueFuture chan *Future, jobQueue chan Job) *dispatcher {
	d := &dispatcher{
		workerPool: workerPool,
		jobQueueWithFuture: jobQueueFuture,
		jobQueue: jobQueue,
		stop: make(chan struct{}),
	}

	for i := 0; i < cap(d.workerPool); i++ {
		worker := newWorker(d.workerPool)
		worker.start()
	}

	go d.dispatch()
	return d
}

// Represents user request, function which should be executed in some worker.
type Job func() interface{}

type Pool struct {
	jobQueueFuture   	chan *Future
	jobQueue			chan Job

	dispatcher			*dispatcher
	wg         			sync.WaitGroup
}

type Future struct {
	// 可以从此channel中取出job结果
	ResultChan		chan interface{}
	job				Job
}

// Will make pool of gorouting workers.
// numWorkers - how many workers will be created for this pool
// queueLen - how many jobs can we accept until we block
//
// Returned object contains JobQueue reference, which you can use to send job to pool.
func NewPool(numWorkers int, jobQueueLen int) *Pool {
	jobQueueFuture := make(chan *Future, jobQueueLen)
	jobQueue := make(chan Job, jobQueueLen)

	workerPool := make(chan *worker, numWorkers)

	pool := &Pool{
		jobQueueFuture:   jobQueueFuture,
		jobQueue: jobQueue,

		dispatcher: newDispatcher(workerPool, jobQueueFuture, jobQueue),
	}

	return pool
}

// 提交有返回值任务, 返回Future指针
func (p *Pool) SubmitFuture(jobFunc Job) *Future {
	f := &Future{
		ResultChan: make(chan interface{}, 1),
		job: jobFunc,
	}

	p.jobQueueFuture <- f
	return f
}

// 提交无返回值的任务
func (p *Pool) Submit(jobFunc Job) {
	p.jobQueue <- jobFunc
	return
}

// In case you are using WaitAll fn, you should call this method
// every time your job is done.
//
// If you are not using WaitAll then we assume you have your own way of synchronizing.
func (p *Pool) JobDone() {
	p.wg.Done()
}

// How many jobs we should wait when calling WaitAll.
// It is using WaitGroup Add/Done/Wait
func (p *Pool) WaitCount(count int) {
	p.wg.Add(count)
}

// Will wait for all jobs to finish.
func (p *Pool) WaitAll() {
	p.wg.Wait()
}

// Will release resources used by pool
func (p *Pool) Release() {
	p.dispatcher.stop <- struct{}{}
	<-p.dispatcher.stop
}
