package dgen

import (
	"strings"
	"testing"
)

// testCase defines the input and expected behavior for a test.
type testCase struct {
	name          string
	input         []byte
	limit         int
	charSet       *CharSet
	wantErr       bool
	checkRequired bool // If true, output must contain at least one char from each required group.
}

// checkError asserts that an error was or was not returned as expected.
func checkError(t *testing.T, err error, wantErr bool, context string) {
	if wantErr && err == nil {
		t.Errorf("%s: expected an error but got none", context)
	}
	if !wantErr && err != nil {
		t.Errorf("%s: unexpected error: %v", context, err)
	}
}

// checkPasswordLength verifies that the generated password has the expected length.
func checkPasswordLength(t *testing.T, pass []byte, expected int) {
	if len(pass) != expected {
		t.Errorf("expected password length %d, got %d", expected, len(pass))
	}
}

// checkDeterminism calls encodeWithComplexity twice and ensures the outputs match.
func checkDeterminism(t *testing.T, input []byte, charSet *CharSet, limit int) []byte {
	pass1, err1 := Encode(input, charSet, limit)
	if err1 != nil {
		t.Fatalf("first call error: %v", err1)
	}
	pass2, err2 := Encode(input, charSet, limit)
	if err2 != nil {
		t.Fatalf("second call error: %v", err2)
	}
	if string(pass1) != string(pass2) {
		t.Errorf("non-determsnistic output: %q vs %q", pass1, pass2)
	}
	return pass1
}

// checkRequiredGroups ensures that the generated password contains at least one character
// from every required group.
func checkRequiredGroups(t *testing.T, pass []byte, charSet *CharSet) {
	for _, group := range charSet.Groups {
		if group.Required {
			if !strings.ContainsAny(string(pass), group.Chars) {
				t.Errorf("password %q does not contain any character from required group %q", pass, group.Name)
			}
		}
	}
}

func TestEncodeWithComplexity(t *testing.T) {
	tests := []testCase{
		{
			name:          "Default configuration",
			input:         []byte("test-input"),
			charSet:       DefaultCharSet(),
			limit:         16,
			wantErr:       false,
			checkRequired: true,
		},
		{
			name:          "Min length matching required groups",
			input:         []byte("min-length"),
			charSet:       DefaultCharSet(), // Default config has 4 required groups
			limit:         16,
			wantErr:       false,
			checkRequired: true,
		},
		{
			name:  "Error when no characters available",
			input: []byte("anything"),
			charSet: &CharSet{
				Groups: []CharGroup{},
			},
			limit:   10,
			wantErr: true,
		},
		{
			name:  "Error when required group is empty",
			input: []byte("test"),
			charSet: &CharSet{
				Groups: []CharGroup{
					{"lowercase", "", true}, // Required but empty
					{"uppercase", "ABCDEFGHIJKLMNOPQRSTUVWXYZ", true},
				},
			},
			limit:   6,
			wantErr: true,
		},
		{
			name:  "Custom character groups",
			input: []byte("custom-test"),
			charSet: &CharSet{
				Groups: []CharGroup{
					{"vowels", "aeiou", true},
					{"consonants", "bcdfghjklmnpqrstvwxyz", true},
					{"numbers", "0123456789", false},
				},
			},
			limit:         10,
			wantErr:       false,
			checkRequired: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			pass, err := Encode(tc.input, tc.charSet, tc.limit)
			context := tc.name + " initial call"
			checkError(t, err, tc.wantErr, context)
			if tc.wantErr {
				return // Skip further checks if an error was expected
			}
			checkPasswordLength(t, pass, tc.limit)
			pass = checkDeterminism(t, tc.input, tc.charSet, tc.limit)
			if tc.checkRequired {
				checkRequiredGroups(t, pass, tc.charSet)
			}
		})
	}
}
