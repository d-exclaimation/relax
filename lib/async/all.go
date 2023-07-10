package async

// AwaitAll waits for all tasks to be resolved or rejected
func AwaitAll[T any](tasks ...Task[T]) []Result[T] {
	res := make([]Result[T], len(tasks))
	for i, task := range tasks {
		res[i] = <-task.channel
	}
	return res
}

// AwaitAllUnit waits for all tasks to be resolved or rejected
func AwaitAllUnit(tasks ...Task[Unit]) []error {
	res := make([]error, len(tasks))
	for i, task := range tasks {
		_, res[i] = task.Await()
	}
	return res
}
