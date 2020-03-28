package runner

import (
	"reflect"
	"testing"
)

func TestContext_Tasks(t *testing.T) {
	ctx := Context{}

	if len(ctx.Tasks()) != 0 {
		t.Fatalf("want 0 tasks, got %d", len(ctx.Tasks()))
	}

	ctx.PushTask(&Task{Name: "foo"})
	ctx.PushTask(&Task{Name: "bar"})

	expected := []string{"foo", "bar"}
	actual := ctx.Tasks()
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("want %v, got %v", expected, actual)
	}
}
