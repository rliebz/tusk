package appcli

import "testing"

var flagPrefixerTests = []struct {
	flags       string
	placeholder string
	expected    string
}{
	{"a", "", "-a"},
	{"a", "foo", "-a <foo>"},
	{"aa", "", "    --aa"},
	{"aa", "foo", "    --aa <foo>"},
	{"a, aa", "", "-a, --aa"},
	{"aa, a", "", "-a, --aa"},
	{"a, aa", "foo", "-a, --aa <foo>"},
}

func TestFlagPrefixer(t *testing.T) {
	for _, tt := range flagPrefixerTests {
		actual := flagPrefixer(tt.flags, tt.placeholder)
		if tt.expected != actual {
			t.Errorf(
				`flagPrefixer("%s", "%s"): expected "%s", got "%s"`,
				tt.flags, tt.placeholder, tt.expected, actual,
			)
		}
	}
}
