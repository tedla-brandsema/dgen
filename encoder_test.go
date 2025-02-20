package main

import (
	"strings"
	"testing"
)

// testCase defines the input and expected behavior for a test.
type testCase struct {
	name          string
	input         []byte
	config        *Config
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
func checkDeterminism(t *testing.T, input []byte, config *Config) []byte {
	pass1, err1 := GeneratePassword(input, config)
	if err1 != nil {
		t.Fatalf("first call error: %v", err1)
	}
	pass2, err2 := GeneratePassword(input, config)
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
func checkRequiredGroups(t *testing.T, pass []byte, config *Config) {
	for _, group := range config.Groups {
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
			config:        DefaultConfig(),
			wantErr:       false,
			checkRequired: true,
		},
		{
			name:          "Min length matching required groups",
			input:         []byte("min-length"),
			config:        DefaultConfig(), // Default config has 4 required groups
			wantErr:       false,
			checkRequired: true,
		},
		{
			name:  "Error when no characters available",
			input: []byte("anything"),
			config: &Config{
				Length: 10,
				Groups: []CharGroup{},
			},
			wantErr: true,
		},
		{
			name:  "Error when required group is empty",
			input: []byte("test"),
			config: &Config{
				Length: 6,
				Groups: []CharGroup{
					{"lowercase", "", true}, // Required but empty
					{"uppercase", "ABCDEFGHIJKLMNOPQRSTUVWXYZ", true},
				},
			},
			wantErr: true,
		},
		{
			name:  "Custom character groups",
			input: []byte("custom-test"),
			config: &Config{
				Length: 10,
				Groups: []CharGroup{
					{"vowels", "aeiou", true},
					{"consonants", "bcdfghjklmnpqrstvwxyz", true},
					{"numbers", "0123456789", false},
				},
			},
			wantErr:       false,
			checkRequired: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			pass, err := GeneratePassword(tc.input, tc.config)
			context := tc.name + " initial call"
			checkError(t, err, tc.wantErr, context)
			if tc.wantErr {
				return // Skip further checks if an error was expected
			}
			checkPasswordLength(t, pass, tc.config.Length)
			pass = checkDeterminism(t, tc.input, tc.config)
			if tc.checkRequired {
				checkRequiredGroups(t, pass, tc.config)
			}
		})
	}
}
