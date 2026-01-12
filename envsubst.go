package envsubst

import (
	"io/ioutil"
	"os"

	"github.com/a8m/envsubst/parse"
)

// String returns the parsed template string after processing it.
// If the parser encounters invalid input, it returns an error describing the failure.
func String(s string) (string, error) {
	return StringRestricted(s, false, false)
}

// StringRestricted returns the parsed template string after processing it.
// If the parser encounters invalid input, or a restriction is violated, it returns
// an error describing the failure.
// Errors on first failure or returns a collection of failures if failOnFirst is false
func StringRestricted(s string, noUnset, noEmpty bool) (string, error) {
	return StringRestrictedNoDigit(s, noUnset, noEmpty , false)
}

// Like StringRestricted but additionally allows to ignore env variables which start with a digit.
func StringRestrictedNoDigit(s string, noUnset, noEmpty bool, noDigit bool) (string, error) {
	return parse.New("string", os.Environ(),
		&parse.Restrictions{noUnset, noEmpty, noDigit}).Parse(s)
}

// StringWithVars returns the parsed template string after processing it,
// only substituting variables specified in the shellFormat string (GNU envsubst style).
// If shellFormat is empty, all variables are substituted (default behavior).
// Example: StringWithVars(input, "$FOO $BAR") only substitutes $FOO and $BAR.
func StringWithVars(s string, shellFormat string) (string, error) {
	return StringRestrictedWithVars(s, false, false, shellFormat)
}

// StringRestrictedWithVars returns the parsed template string with variable filtering.
// Only variables listed in shellFormat are substituted; others are left as literal text.
func StringRestrictedWithVars(s string, noUnset, noEmpty bool, shellFormat string) (string, error) {
	return StringRestrictedNoDigitWithVars(s, noUnset, noEmpty, false, shellFormat)
}

// StringRestrictedNoDigitWithVars returns the parsed template string with all options.
func StringRestrictedNoDigitWithVars(s string, noUnset, noEmpty, noDigit bool, shellFormat string) (string, error) {
	allowedVars := parse.ShellFormatToMap(parse.ParseShellFormat(shellFormat))
	p := &parse.Parser{
		Name:        "string",
		Env:         os.Environ(),
		Restrict:    &parse.Restrictions{noUnset, noEmpty, noDigit},
		AllowedVars: allowedVars,
	}
	return p.Parse(s)
}

// Bytes returns the bytes represented by the parsed template after processing it.
// If the parser encounters invalid input, it returns an error describing the failure.
func Bytes(b []byte) ([]byte, error) {
	return BytesRestricted(b, false, false)
}

// BytesRestricted returns the bytes represented by the parsed template after processing it.
// If the parser encounters invalid input, or a restriction is violated, it returns
// an error describing the failure.
func BytesRestricted(b []byte, noUnset, noEmpty bool) ([]byte, error) {
	return BytesRestrictedNoDigit(b, noUnset, noEmpty, false)
}

// Like BytesRestricted but additionally allows to ignore env variables which start with a digit.
func BytesRestrictedNoDigit(b []byte, noUnset, noEmpty bool, noDigit bool) ([]byte, error) {
	s, err := parse.New("bytes", os.Environ(),
		&parse.Restrictions{noUnset, noEmpty, noDigit}).Parse(string(b))
	if err != nil {
		return nil, err
	}
	return []byte(s), nil
}

// BytesWithVars returns the parsed bytes, only substituting variables in shellFormat.
func BytesWithVars(b []byte, shellFormat string) ([]byte, error) {
	return BytesRestrictedWithVars(b, false, false, shellFormat)
}

// BytesRestrictedWithVars returns the parsed bytes with variable filtering.
func BytesRestrictedWithVars(b []byte, noUnset, noEmpty bool, shellFormat string) ([]byte, error) {
	return BytesRestrictedNoDigitWithVars(b, noUnset, noEmpty, false, shellFormat)
}

// BytesRestrictedNoDigitWithVars returns the parsed bytes with all options.
func BytesRestrictedNoDigitWithVars(b []byte, noUnset, noEmpty, noDigit bool, shellFormat string) ([]byte, error) {
	s, err := StringRestrictedNoDigitWithVars(string(b), noUnset, noEmpty, noDigit, shellFormat)
	if err != nil {
		return nil, err
	}
	return []byte(s), nil
}

// ReadFile call io.ReadFile with the given file name.
// If the call to io.ReadFile failed it returns the error; otherwise it will
// call envsubst.Bytes with the returned content.
func ReadFile(filename string) ([]byte, error) {
	return ReadFileRestricted(filename, false, false)
}

// ReadFileRestricted calls io.ReadFile with the given file name.
// If the call to io.ReadFile failed it returns the error; otherwise it will
// call envsubst.Bytes with the returned content.
func ReadFileRestricted(filename string, noUnset, noEmpty bool) ([]byte, error) {
	return ReadFileRestrictedNoDigit(filename, noUnset, noEmpty, false)
}

// Like ReadFileRestricted but additionally allows to ignore env variables which start with a digit.
func ReadFileRestrictedNoDigit(filename string, noUnset, noEmpty bool, noDigit bool) ([]byte, error) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return BytesRestrictedNoDigit(b, noUnset, noEmpty, noDigit)
}

// ReadFileWithVars reads and parses a file, only substituting variables in shellFormat.
func ReadFileWithVars(filename string, shellFormat string) ([]byte, error) {
	return ReadFileRestrictedWithVars(filename, false, false, shellFormat)
}

// ReadFileRestrictedWithVars reads and parses a file with variable filtering.
func ReadFileRestrictedWithVars(filename string, noUnset, noEmpty bool, shellFormat string) ([]byte, error) {
	return ReadFileRestrictedNoDigitWithVars(filename, noUnset, noEmpty, false, shellFormat)
}

// ReadFileRestrictedNoDigitWithVars reads and parses a file with all options.
func ReadFileRestrictedNoDigitWithVars(filename string, noUnset, noEmpty, noDigit bool, shellFormat string) ([]byte, error) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return BytesRestrictedNoDigitWithVars(b, noUnset, noEmpty, noDigit, shellFormat)
}
