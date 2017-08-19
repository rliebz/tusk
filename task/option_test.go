package task

import (
	"testing"
)

func TestCreateCLIFlag_undefined(t *testing.T) {
	opt := &Option{
		Type: "wrong",
	}

	flag, err := CreateCLIFlag(opt)
	if err == nil {
		t.Fatalf("flag was wrongly created: %#v", flag)
	}
}
