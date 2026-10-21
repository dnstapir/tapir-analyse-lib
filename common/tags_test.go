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

func TestConvertStringToTag(t *testing.T) {
	var tests = []struct {
		name     string
		indata   string
		expected DNSTapirTag
	}{
		{"test_anomalous_qtype_distribution", "anomalous_qtype_distribution", TagAnomalousQtypeDistribution},
		{"test_internal_testing", "internal_testing", TagInternalTesting},
		{"test_undefined", "undefined", TagUndefined},
		{"test_undefined_garbage", "d12380djiuwefi", TagUndefined},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StringToTag(tt.indata)
			if got != tt.expected {
				t.Fatalf("Got: '%d', Expected: '%d'", got, tt.expected)
			}
		})
	}
}
