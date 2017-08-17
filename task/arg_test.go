package task

import (
	"testing"
)

func TestCreateCLIFlag_undefined(t *testing.T) {
	arg := &Arg{
		Type: "wrong",
	}

	flag, err := CreateCLIFlag(arg)
	if err == nil {
		t.Fatalf("flag was wrongly created: %#v", flag)
	}
}
