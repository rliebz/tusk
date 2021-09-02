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

// TaskNames returns the list of task names in the stack, in order. Private
// tasks are filtered out.
func (r *Context) TaskNames() []string {
	output := make([]string, 0, len(r.taskStack))
	for _, t := range r.taskStack {
		if !t.Private {
			output = append(output, t.Name)
		}
	}
	return output
}
