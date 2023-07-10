package async

// Unit is a type that represents a void
type Unit struct{}

var (
	// Done is a unit that represents a void
	Done = Unit{}
)

// Result is a struct that contains the result or error of a task
type Result[T any] struct {
	// The result of the task
	Result T

	// The error of the task
	Error error
}

// Task is an asynchronous value
type Task[T any] struct {
	channel chan Result[T]
}

// New creates a new task from an asynchronous action
func New[T any](action func() (T, error)) Task[T] {
	channel := make(chan Result[T])
	go func() {
		result, err := action()
		channel <- Result[T]{
			Result: result,
			Error:  err,
		}
	}()
	return Task[T]{channel}
}

// Promise creates a new task that can be resolved or rejected later
func Promise[T any]() Task[T] {
	return Task[T]{make(chan Result[T])}
}

// Resolve resolves the task with a result
func (t Task[T]) Resolve(result T) {
	t.channel <- Result[T]{Result: result}
}

// Reject rejects the task with an error
func (t Task[T]) Reject(err error) {
	t.channel <- Result[T]{Error: err}
}

// Await waits for the task to be resolved or rejected
func (t Task[T]) Await() (T, error) {
	res := <-t.channel
	return res.Result, res.Error
}
