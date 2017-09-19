package caddy

import (
	"fmt"
	"strings"
	"testing"

	"github.com/mholt/caddy"
)

func TestParse(t *testing.T) {
	tests := []struct {
		input     string
		shouldErr bool
		expected  []string
	}{
		{
			`
			txtdirect {
				wrong keyword
			}
			`,
			true,
			[]string{},
		},
		{
			`
			txtdirect {
				enable
			}
			`,
			true,
			[]string{},
		},
		{
			`
			txtdirect {
				enable this
				disable that
			}
			`,
			true,
			[]string{},
		},
		{
			`txtdirect`,
			false,
			[]string{"host", "gometa"},
		},
		{
			`
			txtdirect {
				enable host
			}
			`,
			false,
			[]string{"host"},
		},
		{
			`
			txtdirect {
				disable host
			}
			`,
			false,
			[]string{"gometa"},
		},
	}

	for i, test := range tests {
		c := caddy.NewTestController("http", test.input)
		options, err := parse(c)
		if !test.shouldErr && err != nil {
			t.Errorf("Test %d: Unexpected error %s", i, err)
			continue
		}
		if test.shouldErr {
			if err == nil {
				t.Errorf("Test %d: Expected error", i)
			}
			continue
		}

		if !identical(options, test.expected) {
			options := fmt.Sprintf("[ %s ]", strings.Join(options, ", "))
			expected := fmt.Sprintf("[ %s ]", strings.Join(test.expected, ", "))
			t.Errorf("Test %d: Expected options %s, got %s", i, options, expected)
		}
	}
}

func identical(s1, s2 []string) bool {
	if s1 == nil {
		if s2 == nil {
			return true
		}
		return false
	}
	if s2 == nil {
		return false
	}

	if len(s1) != len(s2) {
		return false
	}

	for i := range s1 {
		if s1[i] != s2[i] {
			return false
		}
	}
	return true
}