package task

// RunContext contains contextual information about a run.
type RunContext struct {
	taskStack []*Task
}

// PushTask adds a sub-task to the task stack.
func (r *RunContext) PushTask(t *Task) {
	r.taskStack = append(r.taskStack, t)
}

// Tasks returns the list of tasks in the stack, in order.
func (r *RunContext) Tasks() []string {
	output := make([]string, len(r.taskStack))
	for i, t := range r.taskStack {
		output[i] = t.Name
	}
	return output
}
