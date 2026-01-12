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
		// Basic cases
		{"empty", "", nil},
		{"no vars", "hello world", nil},
		{"single $VAR", "$FOO", []string{"FOO"}},
		{"single ${VAR}", "${BAR}", []string{"BAR"}},
		{"multiple vars", "$FOO $BAR ${BAZ}", []string{"FOO", "BAR", "BAZ"}},
		{"adjacent vars", "$FOO${BAR}$BAZ", []string{"FOO", "BAR", "BAZ"}},
		{"with text", "hello $FOO world ${BAR} end", []string{"FOO", "BAR"}},

		// Escaping
		{"escaped $$", "$$FOO $BAR", []string{"BAR"}},
		{"double escaped $$$$", "$$$$FOO $BAR", []string{"BAR"}},
		{"escaped in middle", "$FOO $$BAR $BAZ", []string{"FOO", "BAZ"}},

		// Underscore handling
		{"underscore ignored", "$_ $FOO ${_}", []string{"FOO"}},
		{"underscore in name", "$FOO_BAR", []string{"FOO_BAR"}},
		{"leading underscore", "$_FOO", []string{"_FOO"}},

		// Operators in ${} syntax
		{"with operators stripped", "${FOO:-default} ${BAR:=value}", []string{"FOO", "BAR"}},
		{"plus operator", "${FOO:+alternate}", []string{"FOO"}},
		{"dash operator", "${FOO-default}", []string{"FOO"}},
		{"equals operator", "${FOO=default}", []string{"FOO"}},

		// Edge cases
		{"trailing $", "hello$", nil},
		{"$ followed by space", "$ FOO $BAR", []string{"BAR"}},
		{"$ followed by special", "$! $@ $# $BAR", []string{"BAR"}},
		{"empty braces ${}", "${} $FOO", []string{"FOO"}},
		{"unclosed brace", "${FOO $BAR", []string{"FOO $BAR"}}, // reads to end
		{"just $", "$", nil},
		{"just ${", "${", nil},

		// Numeric vars: a8m/envsubst allows by default (unlike GNU which ignores)
		// This is intentional - use -no-digit flag if you want GNU behavior
		{"numeric var $1", "$1 $FOO", []string{"1", "FOO"}},
		{"numeric braced ${1}", "${1} $FOO", []string{"1", "FOO"}},
		{"numeric start ${123ABC}", "${123ABC}", []string{"123ABC"}},

		// Duplicates (not deduplicated - that's OK)
		{"duplicate vars", "$FOO $FOO $FOO", []string{"FOO", "FOO", "FOO"}},

		// Real-world ARGOCD_ENV_ patterns
		{"complex", "$ARGOCD_ENV_FOO ${ARGOCD_ENV_BAR:-default}", []string{"ARGOCD_ENV_FOO", "ARGOCD_ENV_BAR"}},
		{"argocd pattern", "$ARGOCD_ENV_CLUSTER $ARGOCD_ENV_DOMAIN $ARGOCD_ENV_SHARD", []string{"ARGOCD_ENV_CLUSTER", "ARGOCD_ENV_DOMAIN", "ARGOCD_ENV_SHARD"}},
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
