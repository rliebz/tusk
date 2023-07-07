package runner

import (
	"fmt"
	"runtime"
	"strconv"
	"testing"

	"github.com/rliebz/ghost"
	"github.com/rliebz/ghost/be"
	yaml "gopkg.in/yaml.v2"
)

func TestWhen_UnmarshalYAML(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  When
	}{
		{
			"short notation",
			`foo`,
			createWhen(withWhenEqual("foo", "true")),
		},
		{
			"list short notation",
			`[foo, bar, baz]`,
			createWhen(
				withWhenEqual("foo", "true"),
				withWhenEqual("bar", "true"),
				withWhenEqual("baz", "true"),
			),
		},
		{
			"not-equal",
			`not-equal: {foo: bar}`,
			createWhen(withWhenNotEqual("foo", "bar")),
		},
		{
			"not-exists",
			`not-exists: file.txt`,
			createWhen(withWhenNotExists("file.txt")),
		},
		{
			"null environment",
			`environment: {foo: null}`,
			createWhen(withoutWhenEnv("foo")),
		},
		{
			"environment list with null",
			`environment: {foo: ["a", null, "b"]}`,
			createWhen(
				withWhenEnv("foo", "a"),
				withoutWhenEnv("foo"),
				withWhenEnv("foo", "b"),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := ghost.New(t)

			var got When
			err := yaml.UnmarshalStrict([]byte(tt.input), &got)
			g.NoError(err)

			g.Should(be.DeepEqual(tt.want, got))
		})
	}
}

func TestWhen_Dependencies(t *testing.T) {
	tests := []struct {
		name string
		when When
		want []string
	}{
		{
			name: "empty",
			when: When{},
			want: []string{},
		},
		{
			name: "equal one",
			when: createWhen(withWhenEqual("foo", "true")),
			want: []string{"foo"},
		},
		{
			name: "equal multi",
			when: createWhen(withWhenEqual("foo", "true"), withWhenEqual("bar", "true")),
			want: []string{"foo", "bar"},
		},
		{
			name: "not equal one",
			when: createWhen(withWhenNotEqual("foo", "true")),
			want: []string{"foo"},
		},
		{
			name: "not equal multi",
			when: createWhen(withWhenNotEqual("foo", "true"), withWhenNotEqual("bar", "true")),
			want: []string{"foo", "bar"},
		},
		{
			name: "both one",
			when: createWhen(withWhenEqual("foo", "true"), withWhenNotEqual("foo", "true")),
			want: []string{"foo"},
		},
		{
			name: "both multi",
			when: createWhen(withWhenEqual("foo", "true"), withWhenNotEqual("bar", "true")),
			want: []string{"foo", "bar"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := ghost.New(t)

			got := tt.when.Dependencies()
			g.Should(beEqualUnordered(tt.want, got))
		})
	}
}

var whenValidateTests = []struct {
	when      When
	options   map[string]string
	shouldErr bool
}{
	// Empty
	{When{}, nil, false},
	{When{}, map[string]string{"foo": "bar"}, false},

	// Command Clauses
	{createWhen(withWhenCommandSuccess), nil, false},
	{createWhen(withWhenCommandFailure), nil, true},
	{createWhen(withWhenCommandSuccess, withWhenCommandSuccess), nil, false},
	{createWhen(withWhenCommandSuccess, withWhenCommandFailure), nil, false},
	{createWhen(withWhenCommandFailure, withWhenCommandFailure), nil, true},

	// Exist Clauses
	{createWhen(withWhenExists("when_test.go")), nil, false},
	{createWhen(withWhenExists("fakefile")), nil, true},
	{createWhen(withWhenExists("fakefile"), withWhenExists("when_test.go")), nil, false},
	{createWhen(withWhenExists("when_test.go"), withWhenExists("fakefile")), nil, false},
	{createWhen(withWhenExists("fakefile"), withWhenExists("fakefile2")), nil, true},

	// Not Exist Clauses
	{createWhen(withWhenNotExists("when_test.go")), nil, true},
	{createWhen(withWhenNotExists("fakefile")), nil, false},
	{createWhen(withWhenNotExists("fakefile"), withWhenNotExists("when_test.go")), nil, false},
	{createWhen(withWhenNotExists("when_test.go"), withWhenNotExists("fakefile")), nil, false},
	{createWhen(withWhenNotExists("fakefile"), withWhenNotExists("fakefile2")), nil, false},
	{createWhen(withWhenNotExists("when.go"), withWhenNotExists("when_test.go")), nil, true},

	// OS Clauses
	{createWhen(withWhenOSSuccess), nil, false},
	{createWhen(withWhenOSFailure), nil, true},
	{createWhen(withWhenOSSuccess, withWhenOSFailure), nil, false},
	{createWhen(withWhenOSFailure, withWhenOSSuccess), nil, false},
	{createWhen(withWhenOSFailure, withWhenOSFailure), nil, true},

	// Environment Clauses
	{createWhen(withWhenEnvSuccess), nil, false},
	{createWhen(withoutWhenEnvSuccess), nil, false},
	{createWhen(withWhenEnvFailure), nil, true},
	{createWhen(withoutWhenEnvFailure), nil, true},
	{createWhen(withWhenEnvSuccess, withoutWhenEnvFailure), nil, false},
	{createWhen(withWhenEnvFailure, withoutWhenEnvSuccess), nil, false},
	{createWhen(withWhenEnvFailure, withoutWhenEnvFailure), nil, true},

	// Equal Clauses
	{
		createWhen(withWhenEqual("foo", "true")),
		map[string]string{"foo": "true"},
		false,
	},
	{
		createWhen(withWhenEqual("foo", "true"), withWhenEqual("bar", "false")),
		map[string]string{"foo": "true", "bar": "false"},
		false,
	},
	{
		createWhen(withWhenEqual("foo", "true"), withWhenEqual("bar", "false")),
		map[string]string{"foo": "true", "bar": "true"},
		false,
	},
	{
		createWhen(withWhenEqual("foo", "true"), withWhenEqual("bar", "false")),
		map[string]string{"foo": "true"},
		false,
	},
	{
		createWhen(withWhenEqual("foo", "true")),
		map[string]string{"foo": "false"},
		true,
	},
	{
		createWhen(withWhenEqual("foo", "true")),
		map[string]string{},
		true,
	},
	{
		createWhen(withWhenEqual("foo", "true")),
		map[string]string{"bar": "true"},
		true,
	},
	{
		createWhen(withWhenEqual("foo", "true"), withWhenEqual("bar", "false")),
		map[string]string{"bar": "true"},
		true,
	},

	// NotEqual Clauses
	{
		createWhen(withWhenNotEqual("foo", "true")),
		map[string]string{"foo": "true"},
		true,
	},
	{
		createWhen(withWhenNotEqual("foo", "true")),
		map[string]string{"foo": "false"},
		false,
	},
	{
		createWhen(withWhenNotEqual("foo", "true"), withWhenNotEqual("bar", "false")),
		map[string]string{"foo": "true", "bar": "true"},
		false,
	},
	{
		createWhen(withWhenNotEqual("foo", "true"), withWhenNotEqual("bar", "false")),
		map[string]string{"foo": "false"},
		false,
	},
	{
		createWhen(withWhenNotEqual("foo", "true")),
		map[string]string{},
		true,
	},
	{
		createWhen(withWhenNotEqual("foo", "true")),
		map[string]string{"bar": "true"},
		true,
	},
	{
		createWhen(withWhenNotEqual("foo", "true"), withWhenNotEqual("bar", "true")),
		map[string]string{"bar": "true"},
		true,
	},
	{
		createWhen(withWhenNotEqual("foo", "true"), withWhenNotEqual("bar", "true")),
		map[string]string{"foo": "false", "bar": "false"},
		false,
	},

	// Combinations
	{
		createWhen(
			withWhenCommandFailure,
			withWhenExists("fakefile"),
			withWhenNotExists("when_test.go"),
			withWhenOSFailure,
			withWhenEnvFailure,
			withWhenEqual("foo", "wrong"),
			withWhenNotEqual("foo", "true"),
		),
		map[string]string{"foo": "true"},
		true,
	},
	{
		createWhen(
			withWhenCommandSuccess,
			withWhenExists("fakefile"),
			withWhenNotExists("when_test.go"),
			withWhenOSFailure,
			withWhenEnvFailure,
			withWhenEqual("foo", "wrong"),
			withWhenNotEqual("foo", "true"),
		),
		map[string]string{"foo": "true"},
		false,
	},
	{
		createWhen(
			withWhenCommandFailure,
			withWhenExists("when_test.go"),
			withWhenNotExists("when_test.go"),
			withWhenOSFailure,
			withWhenEnvFailure,
			withWhenEqual("foo", "wrong"),
			withWhenNotEqual("foo", "true"),
		),
		map[string]string{"foo": "true"},
		false,
	},
	{
		createWhen(
			withWhenCommandFailure,
			withWhenExists("fakefile"),
			withWhenNotExists("fakefile"),
			withWhenOSFailure,
			withWhenEnvFailure,
			withWhenEqual("foo", "wrong"),
			withWhenNotEqual("foo", "true"),
		),
		map[string]string{"foo": "true"},
		false,
	},
	{
		createWhen(
			withWhenCommandFailure,
			withWhenExists("fakefile"),
			withWhenNotExists("when_test.go"),
			withWhenOSSuccess,
			withWhenEnvFailure,
			withWhenEqual("foo", "wrong"),
			withWhenNotEqual("foo", "true"),
		),
		map[string]string{"foo": "true"},
		false,
	},
	{
		createWhen(
			withWhenCommandFailure,
			withWhenExists("fakefile"),
			withWhenNotExists("when_test.go"),
			withWhenOSFailure,
			withWhenEnvSuccess,
			withWhenEqual("foo", "wrong"),
			withWhenNotEqual("foo", "true"),
		),
		map[string]string{"foo": "true"},
		false,
	},
	{
		createWhen(
			withWhenCommandFailure,
			withWhenExists("fakefile"),
			withWhenNotExists("when_test.go"),
			withWhenOSFailure,
			withWhenEnvFailure,
			withWhenEqual("foo", "true"),
			withWhenNotEqual("foo", "true"),
		),
		map[string]string{"foo": "true"},
		false,
	},
	{
		createWhen(
			withWhenCommandFailure,
			withWhenExists("fakefile"),
			withWhenNotExists("when_test.go"),
			withWhenOSFailure,
			withWhenEnvFailure,
			withWhenEqual("foo", "wrong"),
			withWhenNotEqual("foo", "fake"),
		),
		map[string]string{"foo": "true"},
		false,
	},
}

func TestWhen_Validate(t *testing.T) {
	for i, tt := range whenValidateTests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			g := ghost.New(t)

			err := tt.when.Validate(Context{}, tt.options)
			g.Should(be.Equal(tt.shouldErr, err != nil))
		})
	}

	// TODO: Combine this table with above tests
	tests := []struct {
		name    string
		context Context
		when    When
	}{
		{
			name:    "directory exists",
			context: Context{Dir: "testdata"},
			when: createWhen(
				withWhenExists("exists.txt"), // file in testdata
			),
		},
		{
			name:    "directory not exists",
			context: Context{Dir: "testdata"},
			when: createWhen(
				withWhenNotExists("when_test.go"), // file in working dir, not in testdata
			),
		},
		{
			name:    "directory command",
			context: Context{Dir: "testdata"},
			when: createWhen(
				withWhenCommand(`test "$(basename "$(pwd)")" = "testdata"`),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := ghost.New(t)

			err := tt.when.Validate(tt.context, nil)
			g.NoError(err)
		})
	}
}

func TestNormalizeOS(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"nonsense", "nonsense"},
		{"darwin", "darwin"},
		{"Darwin", "darwin"},
		{"OSX", "darwin"},
		{"macOS", "darwin"},
		{"win", "windows"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			g := ghost.New(t)

			got := normalizeOS(tt.input)
			g.Should(be.Equal(tt.want, got))
		})
	}
}

func TestList_UnmarshalYAML(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  WhenList
	}{
		{
			"single item",
			"os: linux",
			WhenList{createWhen(withWhenOS("linux"))},
		},
		{
			"list length 1",
			"[os: linux]",
			WhenList{createWhen(withWhenOS("linux"))},
		},
		{
			"single item short",
			"foo",
			WhenList{createWhen(withWhenEqual("foo", "true"))},
		},
		{
			"list implies multiple whens",
			"[foo, bar]",
			WhenList{createWhen(withWhenEqual("foo", "true")), createWhen(withWhenEqual("bar", "true"))},
		},
		{
			"nested short lists",
			"[[foo, bar], [baz]]",
			WhenList{
				createWhen(withWhenEqual("foo", "true"), withWhenEqual("bar", "true")),
				createWhen(withWhenEqual("baz", "true")),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := ghost.New(t)

			var got WhenList
			err := yaml.UnmarshalStrict([]byte(tt.input), &got)
			g.NoError(err)

			g.Should(be.DeepEqual(tt.want, got))
		})
	}
}

func TestList_Dependencies(t *testing.T) {
	tests := []struct {
		name string
		list WhenList
		want []string
	}{
		{
			"empty list",
			WhenList{},
			[]string{},
		},
		{
			"single item list",
			WhenList{
				createWhen(withWhenEqual("foo", "true"), withWhenEqual("bar", "true")),
			},
			[]string{"foo", "bar"},
		},
		{
			"duplicate across lists",
			WhenList{
				createWhen(withWhenEqual("foo", "true")),
				createWhen(withWhenEqual("foo", "true")),
			},
			[]string{"foo"},
		},
		{
			"different items per list",
			WhenList{
				createWhen(withWhenEqual("foo", "true")),
				createWhen(withWhenEqual("bar", "true")),
			},
			[]string{"foo", "bar"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := ghost.New(t)

			got := tt.list.Dependencies()
			g.Should(beEqualUnordered(tt.want, got))
		})
	}
}

func TestList_Dependencies_nil(t *testing.T) {
	g := ghost.New(t)

	var l *WhenList
	g.Should(be.SliceLen(0, l.Dependencies()))
}

func TestList_Validate(t *testing.T) {
	tests := []struct {
		name    string
		list    WhenList
		options map[string]string
		wantErr string
	}{
		{
			"all valid",
			WhenList{whenTrue, whenTrue, whenTrue},
			nil,
			"",
		},
		{
			"all invalid",
			WhenList{whenFalse, whenFalse, whenFalse},
			nil,
			fmt.Sprintf("current OS (%s) not listed in [fake]", runtime.GOOS),
		},
		{
			"some invalid",
			WhenList{whenTrue, whenFalse, whenTrue},
			nil,
			fmt.Sprintf("current OS (%s) not listed in [fake]", runtime.GOOS),
		},
		{
			"passes requirements",
			WhenList{
				createWhen(withWhenEqual("foo", "true")),
				createWhen(withWhenEqual("bar", "false")),
			},
			map[string]string{"foo": "true", "bar": "false"},
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := ghost.New(t)

			err := tt.list.Validate(Context{}, tt.options)
			if tt.wantErr != "" {
				g.Should(be.ErrorEqual(tt.wantErr, err))
				return
			}
			g.NoError(err)
		})
	}
}

func TestList_Validate_nil(t *testing.T) {
	g := ghost.New(t)

	var l *WhenList
	err := l.Validate(Context{}, map[string]string{})
	g.NoError(err)
}
