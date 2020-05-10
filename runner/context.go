package runner

import "github.com/rliebz/tusk/ui"

// Context contains contextual information about a run.
type Context struct {
	// Logger is responsible for logging actions as they occur. It is required to
	// be defined for a Context.
	Logger *ui.Logger

	// Interpreter specifies how a command is meant to be executed.
	Interpreter []string

	taskStack []*Task
}

// PushTask adds a sub-task to the task stack.
func (r *Context) PushTask(t *Task) {
	r.taskStack = append(r.taskStack, t)
}

// Tasks returns the list of tasks in the stack, in order.
func (r *Context) Tasks() []string {
	output := make([]string, len(r.taskStack))
	for i, t := range r.taskStack {
		output[i] = t.Name
	}
	return output
}
