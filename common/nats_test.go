package common

import (
	"testing"
)

func TestNormalizeNatsSubject(t *testing.T) {
	var tests = []struct {
		name     string
		indata   string
		expected string
	}{
		{"root", ".", ""},
		{"empty", "", ""},
		{"camel", "a.TestSubject", "a.testsubject"},
		{"asterisk_start", "*.test.subject", "*.test.subject"},
		{"asterisk_end", "test.subject.*", "test.subject.*"},
		{"asterisk_mid", "test.*.subject", "test.*.subject"},
		{"many_dots", ".....test.subject....", "test.subject"},
		{"many_dots_only", ".........", ""},
		{"camel_asterisk", "ImATestSubject.Yes.I.Am.*.caMel", "imatestsubject.yes.i.am.*.camel"},
		{"camel_wildcard", "ImATestSubject.Yes.I.Am.>", "imatestsubject.yes.i.am.>"},
		{"trailing_dot_1", "test.subject.>.", "test.subject.>"},
		{"trailing_dot_2", "test.subject.*", "test.subject.*"},
		{"starting_dot_1", ".test.subject.>", "test.subject.>"},
		{"starting_dot_2", ".test.subject.*", "test.subject.*"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeNatsSubject(tt.indata)
			if got != tt.expected {
				t.Fatalf("got '%s', expected '%s'", got, tt.expected)
			}
		})
	}
}

func TestNormalizeIsTheSame(t *testing.T) {
	var tests = []struct {
		name   string
		indata string
	}{
		{"root", "."},
		{"empty", ""},
		{"camel", "a.TestSubject"},
		{"asterisk_start", "*.test.subject"},
		{"asterisk_end", "test.subject.*"},
		{"asterisk_mid", "test.*.subject"},
		{"many_dots", ".....test.subject...."},
		{"many_dots_only", "........."},
		{"camel_asterisk", "ImATestSubject.Yes.I.Am.*.caMel"},
		{"camel_wildcard", "ImATestSubject.Yes.I.Am.>"},
		{"trailing_dot_1", "test.subject.>."},
		{"trailing_dot_2", "test.subject.*"},
		{"starting_dot_1", ".test.subject.>"},
		{"starting_dot_2", ".test.subject.*"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got1 := NormalizeNatsSubject(tt.indata)
			got2 := NormalizeNatsSubjectPrefix(tt.indata)
			if got1 != got2 {
				t.Fatalf("Normalized subject: '%s', Normalized prefix '%s'", got1, got2)
			}
		})
	}
}
