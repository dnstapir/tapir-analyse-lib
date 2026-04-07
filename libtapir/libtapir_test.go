package libtapir

import (
	"testing"
)

func TestNormalizeDomainName(t *testing.T) {
	var tests = []struct {
		name     string
		indata   string
		expected string
	}{
		{"root", ".", "."},
		{"empty", "", "."},
		{"camel", "MyExample.Xa.", "myexample.xa."},
		{"no_trailing_dot", "example.xa", "example.xa."},
		{"asterisk", "*.example.xa", "example.xa."},
		{"asterisk_nodot", "*example.xa", "example.xa."},
		{"many_dots", ".....example.xa....", "example.xa."},
		{"many_dots_only", ".........", "."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeDomainName(tt.indata)
			if got != tt.expected {
				t.Fatalf("got '%s', expected '%s'", got, tt.expected)
			}
		})
	}
}

func TestNormalizeDomainNameSuffix(t *testing.T) {
	var tests = []struct {
		name     string
		indata   string
		expected string
	}{
		{"root", ".", "."},
		{"empty", "", "."},
		{"camel", "MyExample.Xa.", ".myexample.xa."},
		{"no_trailing_dot", "example.xa", ".example.xa."},
		{"asterisk", "*.example.xa", ".example.xa."},
		{"asterisk_nodot", "*example.xa", ".example.xa."},
		{"many_dots", ".....example.xa....", ".example.xa."},
		{"many_dots_only", ".........", "."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeDomainNameSuffix(tt.indata)
			if got != tt.expected {
				t.Fatalf("got '%s', expected '%s'", got, tt.expected)
			}
		})
	}
}

func TestFlipDomainName(t *testing.T) {
	var tests = []struct {
		name     string
		indata   string
		expected string
	}{
		{"root", ".", "."},
		{"empty", "", "."},
		{"camel", "MyExample.Xa.", "xa.myexample"},
		{"no_trailing_dot", "example.xa", "xa.example"},
		{"asterisk", "*.example.xa", "xa.example"},
		{"asterisk_nodot", "*example.xa", "xa.example"},
		{"many_dots", ".....example.xa....", "xa.example"},
		{"many_dots_only", ".........", "."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FlipDomainName(tt.indata)
			if got != tt.expected {
				t.Fatalf("got '%s', expected '%s'", got, tt.expected)
			}
		})
	}
}
