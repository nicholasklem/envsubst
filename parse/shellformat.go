package parse

import (
	"strings"
	"unicode"
)

// ParseShellFormat extracts variable names from a GNU envsubst SHELL-FORMAT string.
// It handles both $VAR and ${VAR} forms.
// Example: "$FOO ${BAR} $BAZ" returns ["FOO", "BAR", "BAZ"]
func ParseShellFormat(shellFormat string) []string {
	if shellFormat == "" {
		return nil
	}
	var vars []string
	i := 0
	for i < len(shellFormat) {
		if shellFormat[i] == '$' {
			i++
			if i >= len(shellFormat) {
				break
			}
			// Handle ${VAR} form
			if shellFormat[i] == '{' {
				i++
				start := i
				for i < len(shellFormat) && shellFormat[i] != '}' {
					i++
				}
				if start < i {
					varName := shellFormat[start:i]
					// Strip any operator suffix (:-default, :=value, etc)
					if idx := strings.IndexAny(varName, ":-+="); idx != -1 {
						varName = varName[:idx]
					}
					if varName != "" && varName != "_" {
						vars = append(vars, varName)
					}
				}
				if i < len(shellFormat) {
					i++ // skip '}'
				}
			} else if shellFormat[i] == '$' {
				// Skip escaped $$
				i++
			} else if isAlphaNumericByte(shellFormat[i]) {
				// Handle $VAR form
				start := i
				for i < len(shellFormat) && isAlphaNumericByte(shellFormat[i]) {
					i++
				}
				varName := shellFormat[start:i]
				if varName != "_" {
					vars = append(vars, varName)
				}
			}
		} else {
			i++
		}
	}
	return vars
}

// ShellFormatToMap converts a slice of variable names to a map for O(1) lookup.
// Returns nil if the input slice is empty or nil, which signals "allow all vars".
func ShellFormatToMap(vars []string) map[string]bool {
	if len(vars) == 0 {
		return nil
	}
	m := make(map[string]bool, len(vars))
	for _, v := range vars {
		m[v] = true
	}
	return m
}

func isAlphaNumericByte(b byte) bool {
	r := rune(b)
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}
