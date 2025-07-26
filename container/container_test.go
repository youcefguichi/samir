package main

import (
	"strconv"
	"testing"
)

func TestValidateMemoryInputAndExtractValue(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		wantErr  bool
	}{
		{"100Mb", "100", false},
		{"0Mb", "0", false},
		{"-10Mb", "", true},
		{"100mb", "", true},
		{"100", "", true},
		{"Mb", "", true},
	}

	for _, tt := range tests {
		got, err := ValidateMemoryInputAndExtractValue(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("ValidateMemoryInputAndExtractValue(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
		}
		if got != tt.expected {
			t.Errorf("ValidateMemoryInputAndExtractValue(%q) = %v, want %v", tt.input, got, tt.expected)
		}
	}
}

func TestValidateCpuInputAndExtractValue(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		wantErr  bool
	}{
		{"100m", "100", false},
		{"0m", "0", false},
		{"-10m", "", true},
		{"100M", "", true},
		{"100", "", true},
		{"m", "", true},
	}

	for _, tt := range tests {
		got, err := validateCpuInputAndExtractValue(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("validateCpuInputAndExtractValue(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
		}
		if got != tt.expected {
			t.Errorf("validateCpuInputAndExtractValue(%q) = %v, want %v", tt.input, got, tt.expected)
		}
	}
}

func TestConvertFromMbtoBytes(t *testing.T) {
	tests := []struct {
		input     string
		expected  int64
		wantPanic bool
	}{
		{"1", 1024 * 1024, false},
		{"0", 0, false},
		{"-1", 0, true},
		{"abc", 0, true},
	}

	for _, tt := range tests {
		func() {
			defer func() {
				if r := recover(); r != nil {
					if !tt.wantPanic {
						t.Errorf("convertFromMbtoBytes(%q) panicked unexpectedly: %v", tt.input, r)
					}
				} else if tt.wantPanic {
					t.Errorf("convertFromMbtoBytes(%q) did not panic as expected", tt.input)
				}
			}()
			got := convertFromMbtoBytes(tt.input)
			if !tt.wantPanic {
				val, err := strconv.ParseInt(got, 10, 64)
				if err != nil {
					t.Errorf("convertFromMbtoBytes(%q) returned invalid int: %v", tt.input, got)
				}
				if val != tt.expected {
					t.Errorf("convertFromMbtoBytes(%q) = %v, want %v", tt.input, val, tt.expected)
				}
			}
		}()
	}
}

func TestIsRootOrGuest(t *testing.T) {
	entries := []struct {
		username string
		expected bool
	}{
		{"root", true},
		{"guest", true},
		{"admin", false},
		{"user", false},
		{"ROOT", false},
		{"Guest", false},
		{"", false},
	}

	for _, entry := range entries {
		result := isRootOrGuest(entry.username)
		if result != entry.expected {
			t.Errorf("isRootOrGuest(%q) = %v, want %v", entry.username, result, entry.expected)
		}
	}
}
