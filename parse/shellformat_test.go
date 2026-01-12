package parse

import (
	"reflect"
	"testing"
)

func TestParseShellFormat(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{"empty", "", nil},
		{"no vars", "hello world", nil},
		{"single $VAR", "$FOO", []string{"FOO"}},
		{"single ${VAR}", "${BAR}", []string{"BAR"}},
		{"multiple vars", "$FOO $BAR ${BAZ}", []string{"FOO", "BAR", "BAZ"}},
		{"adjacent vars", "$FOO${BAR}$BAZ", []string{"FOO", "BAR", "BAZ"}},
		{"with text", "hello $FOO world ${BAR} end", []string{"FOO", "BAR"}},
		{"escaped $$", "$$FOO $BAR", []string{"BAR"}},
		{"underscore ignored", "$_ $FOO ${_}", []string{"FOO"}},
		{"with operators stripped", "${FOO:-default} ${BAR:=value}", []string{"FOO", "BAR"}},
		{"complex", "$ARGOCD_ENV_FOO ${ARGOCD_ENV_BAR:-default}", []string{"ARGOCD_ENV_FOO", "ARGOCD_ENV_BAR"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := ParseShellFormat(test.input)
			if !reflect.DeepEqual(result, test.expected) {
				t.Errorf("ParseShellFormat(%q) = %v, expected %v", test.input, result, test.expected)
			}
		})
	}
}

func TestShellFormatToMap(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected map[string]bool
	}{
		{"nil", nil, nil},
		{"empty", []string{}, nil},
		{"single", []string{"FOO"}, map[string]bool{"FOO": true}},
		{"multiple", []string{"FOO", "BAR", "BAZ"}, map[string]bool{"FOO": true, "BAR": true, "BAZ": true}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := ShellFormatToMap(test.input)
			if !reflect.DeepEqual(result, test.expected) {
				t.Errorf("ShellFormatToMap(%v) = %v, expected %v", test.input, result, test.expected)
			}
		})
	}
}

// Test variable filtering during parsing
func TestParseWithShellFormat(t *testing.T) {
	env := []string{
		"FOO=foo",
		"BAR=bar",
		"BAZ=baz",
	}

	tests := []struct {
		name        string
		input       string
		shellFormat string
		expected    string
	}{
		// Basic filtering
		{"no filter - all vars", "$FOO $BAR $BAZ", "", "foo bar baz"},
		{"filter single var", "$FOO $BAR $BAZ", "$FOO", "foo $BAR $BAZ"},
		{"filter two vars", "$FOO $BAR $BAZ", "$FOO $BAR", "foo bar $BAZ"},
		{"filter all vars", "$FOO $BAR $BAZ", "$FOO $BAR $BAZ", "foo bar baz"},
		{"filter none (empty vars)", "$FOO $BAR", "$NOTINPUT", "$FOO $BAR"},

		// Braced syntax
		{"filter ${VAR}", "${FOO} ${BAR}", "$FOO", "foo ${BAR}"},

		// Extended syntax preserved when not allowed
		{"filter with default", "${FOO:-default} ${BAR:-default}", "$FOO", "foo ${BAR:-default}"},
		{"filter with equals", "${FOO:=val} ${BAR:=val}", "$FOO", "foo ${BAR:=val}"},
		{"filter with plus", "${FOO:+other} ${BAR:+other}", "$FOO", "other ${BAR:+other}"},

		// Mixed text
		{"mixed with text", "hello $FOO world $BAR end", "$FOO", "hello foo world $BAR end"},

		// Unset vars in allowed list
		{"unset in allowed", "$FOO $NOTSET", "$FOO $NOTSET", "foo "},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			allowedVars := ShellFormatToMap(ParseShellFormat(test.shellFormat))
			p := &Parser{
				Name:        test.name,
				Env:         env,
				Restrict:    Relaxed,
				AllowedVars: allowedVars,
			}
			result, err := p.Parse(test.input)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if result != test.expected {
				t.Errorf("Parse(%q) with shellFormat %q = %q, expected %q",
					test.input, test.shellFormat, result, test.expected)
			}
		})
	}
}
