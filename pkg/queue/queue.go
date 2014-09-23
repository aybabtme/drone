package queue

import (
	"expvar"
	"fmt"
	. "github.com/drone/drone/pkg/model"
)

// A Queue dispatches tasks to workers.
type Queue struct {
	queueLen *expvar.Int
	tasks    chan<- *BuildTask
}

// BuildTasks represents a build that is pending
// execution.
type BuildTask struct {
	Repo   *Repo
	Commit *Commit
	Build  *Build
}

// Start N workers with the given build runner.
func Start(workers int, runner BuildRunner) *Queue {
	tasks := make(chan *BuildTask)

	metrics := expvar.NewMap("queue")
	queueLen := &expvar.Int{}
	workerCount := &expvar.Int{}
	workerWaiting := &expvar.Int{}
	workerWorking := &expvar.Int{}
	metrics.Set("queueLen", queueLen)
	metrics.Set("workerCount", workerCount)
	metrics.Set("workerWaiting", workerWaiting)
	metrics.Set("workerWorking", workerWorking)

	queue := &Queue{
		queueLen: queueLen,
		tasks:    tasks,
	}

	for i := 0; i < workers; i++ {

		workerMet := &expvar.Map{}
		workerMet.Init()
		repo := &expvar.String{}
		branch := &expvar.String{}
		commit := &expvar.String{}
		workerMet.Set("repo", repo)
		workerMet.Set("branch", branch)
		workerMet.Set("commit", commit)

		worker := worker{
			waiting: workerWaiting,
			working: workerWorking,
			repo:    repo,
			branch:  branch,
			commit:  commit,
			runner:  runner,
		}

		metrics.Set(fmt.Sprintf("worker%d", i), workerMet)
		workerCount.Add(1)

		go worker.work(tasks)
	}

	return queue
}

// Add adds the task to the build queue.
func (q *Queue) Add(task *BuildTask) {
	q.queueLen.Add(1)
	q.tasks <- task
	q.queueLen.Add(-1)
}
