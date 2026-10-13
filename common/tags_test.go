package common

import (
	"testing"
)

func TestConvertTagToString(t *testing.T) {
	var tests = []struct {
		name     string
		indata   DNSTapirTag
		expected string
	}{
		{"test_anomalous_qtype_distribution", TagAnomalousQtypeDistribution, "anomalous_qtype_distribution"},
		{"test_internal_testing", TagInternalTesting, "internal_testing"},
		{"test_undefined", TagUndefined, "undefined"},
		{"test_undefined_clamp", TagUndefined + 1, "undefined"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DNSTapirTag(tt.indata).String()
			if got != tt.expected {
				t.Fatalf("Got: '%s', Expected: '%s'", got, tt.expected)
			}
		})
	}
}
