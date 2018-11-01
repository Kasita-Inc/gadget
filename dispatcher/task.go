package dispatcher

import "time"

// Task is the unit of work to be executed by a worker in the pool.
type Task interface {
	// Execute function on the task on receipt and log the error.
	Execute() error
}

type retryTask struct {
	base    Task
	retry   func() bool
	retries int
	period  time.Duration
}

// NewRetryTask for the passed task that will retry execution up to the amount of retries specified
// whenever the passed retry function returns true.
func NewRetryTask(task Task, retry func() bool, retries int, period time.Duration) Task {
	return &retryTask{base: task, retry: retry, retries: retries, period: period}
}

func (rt *retryTask) Execute() error {
	retry := true
	for i := 0; i < rt.retries && retry; i++ {
		time.Sleep(rt.period * time.Duration(i*5))
		rt.base.Execute()
		retry = rt.retry()
	}
	return nil
}

// TimeoutTask that will fail after the specified duration has elapsed after execute
// is called on the task.
type TimeoutTask interface {
	Task
	GetTimeout() time.Duration
}

// NewTimeoutTask that will fail after the specified duration has elapsed after execute
// is called on the task.
func NewTimeoutTask(task Task, timeout time.Duration) TimeoutTask {
	return &timeoutTask{Task: task, timeout: timeout}
}

type timeoutTask struct {
	Task
	timeout time.Duration
}

func (task *timeoutTask) GetTimeout() time.Duration {
	return task.timeout
}