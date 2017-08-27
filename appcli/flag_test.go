package appcli

import (
	"testing"

	"gitlab.com/rliebz/tusk/task"
)

func TestCreateCLIFlag_undefined(t *testing.T) {
	opt := &task.Option{
		Type: "wrong",
	}

	flag, err := createCLIFlag(opt)
	if err == nil {
		t.Fatalf("flag was wrongly created: %#v", flag)
	}
}