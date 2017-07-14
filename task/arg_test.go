package task

import (
	"testing"

	"math"

	"github.com/urfave/cli"
)

func TestCreateCLIFlag_int(t *testing.T) {

	verifyArg := func(name string, value int) {
		arg := &Arg{
			Type:    name,
			Default: value,
		}

		flag, err := CreateCLIFlag(arg)
		if err != nil {
			t.Errorf("error creating flag: %s", err)
			return
		}

		intFlag, ok := flag.(cli.IntFlag)
		if !ok {
			t.Errorf("error converting to cli.IntFlag: %#v", flag)
			return
		}

		actual := intFlag.Value
		if actual != value {
			t.Errorf("expected %d, actual %d", value, actual)
		}
	}

	// Check the boundaries and spellings
	verifyArg("int", math.MinInt64)
	verifyArg("Integer", math.MaxInt64)

	// Try a range of numbers
	for i := math.MinInt8; i < math.MaxInt8; i++ {
		verifyArg("INT", i)
	}
}

func TestCreateCLIFlag_float(t *testing.T) {

	verifyArg := func(name string, value float64) {
		arg := &Arg{
			Type:    name,
			Default: value,
		}

		flag, err := CreateCLIFlag(arg)
		if err != nil {
			t.Errorf("error creating flag: %s", err)
			return
		}

		floatFlag, ok := flag.(cli.Float64Flag)
		if !ok {
			t.Errorf("error converting to cli.Float64Flag: %#v", flag)
			return
		}

		actual := floatFlag.Value
		if actual != value {
			t.Errorf("expected %f, actual %f", value, actual)
		}
	}

	// Check the boundaries and spellings
	verifyArg("float", math.SmallestNonzeroFloat64)
	verifyArg("Float64", math.MaxFloat64)
	verifyArg("floaT", -math.SmallestNonzeroFloat64)
	verifyArg("FlOaT64", -math.MaxFloat64)
	verifyArg("DOUBLE", 0)

	// Check various numbers
	verifyArg("float", 1)
	verifyArg("float", math.Pi)
	verifyArg("float", math.Sqrt2)
	verifyArg("float", math.E)
}

func TestCreateCLIFlag_bool(t *testing.T) {

	arg := &Arg{
		Type:    "bool",
		Default: false,
	}

	flag, err := CreateCLIFlag(arg)
	if err != nil {
		t.Fatalf("error creating flag: %s", err)
	}

	_, ok := flag.(cli.BoolFlag)
	if !ok {
		t.Errorf("error converting to cli.BoolFlag: %#v", flag)
	}
}

func TestCreateCLIFlag_bool_t(t *testing.T) {
	arg := &Arg{
		Type:    "bool",
		Default: true,
	}

	flag, err := CreateCLIFlag(arg)
	if err != nil {
		t.Fatalf("error creating flag: %s", err)
	}

	_, ok := flag.(cli.BoolTFlag)
	if !ok {
		t.Errorf("error converting to cli.BoolTFlag: %#v", flag)
	}
}

func TestCreateCLIFlag_string(t *testing.T) {

	verifyArg := func(name string, value string) {
		arg := &Arg{
			Type:    name,
			Default: value,
		}

		flag, err := CreateCLIFlag(arg)
		if err != nil {
			t.Errorf("error creating flag: %s", err)
			return
		}

		intFlag, ok := flag.(cli.StringFlag)
		if !ok {
			t.Errorf("error converting to cli.StringFlag: %#v", flag)
			return
		}

		actual := intFlag.Value
		if actual != value {
			t.Errorf("expected %s, actual %s", value, actual)
		}
	}

	verifyArg("string", "some string")
	verifyArg("String", "")
	verifyArg("", "Hello, 世界")
}

func TestCreateCLIFlag_undefined(t *testing.T) {
	arg := &Arg{
		Type: "wrong",
	}

	flag, err := CreateCLIFlag(arg)
	if err == nil {
		t.Fatalf("flag was wrongly created: %#v", flag)
	}
}
