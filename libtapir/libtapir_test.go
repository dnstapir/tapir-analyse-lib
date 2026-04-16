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
		{"whitespace", "    myexample.xa.\n", "myexample.xa."},
		{"whitespace-trailing", "myexample.xa.\n", "myexample.xa."},
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

func TestHasValidETLD(t *testing.T) {
	var tests = []struct {
		name     string
		indata   string
		expected bool
	}{
		{"root", ".", false},
		{"empty", "", false},
		{"many-dots", ".........", false},
		{"lan", "test.lan", false},
		{"lan-fqdn", "test.lan.", false},
		{"single", "org", true},
		{"single-fqdn", "org.", true},
		{"internal", "test.internal", false},
		{"internal-fqdn", "test.internal.", false},
		{"se", "test.se", true},
		{"se-fqdn", "test.se.", true},
		{"gov.uk", "test.gov.uk", true},
		{"gov.uk-fqdn", "test.gov.uk.", true},
		{"long", "ndioqudh89u2inwref98hsd129d834fby.net", true},
		{"long-with-many-dots", "......ndioqudh89u2inwref98hsd129d834fby.net....", true},
		{"reverse", "a.9.7.1.4.1.8.6.0.0.0.0.0.0.0.0.0.0.0.0.0.1.0.0.0.0.7.4.6.0.6.2.ip6.arpa", true},
		{"reverse-fqdn", "a.9.7.1.4.1.8.6.0.0.0.0.0.0.0.0.0.0.0.0.0.1.0.0.0.0.7.4.6.0.6.2.ip6.arpa.", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HasValidETLD(tt.indata)
			if got != tt.expected {
				t.Fatalf("got '%t', expected '%t'", got, tt.expected)
			}
		})
	}
}

func TestGetETLDPlusOne(t *testing.T) {
	var tests = []struct {
		name     string
		indata   string
		expected string
	}{
		{"root", ".", "."},
		{"empty", "", "."},
		{"many-dots", ".........", "."},
		{"lan", "test.lan", "test.lan."},
		{"lan-fqdn", "test.lan.", "test.lan."},
		{"single", "org", "."},
		{"single-fqdn", "org.", "."},
		{"internal", "test.internal", "test.internal."},
		{"internal-fqdn", "test.internal.", "test.internal."},
		{"se", "test.se", "test.se."},
		{"se-fqdn", "test.se.", "test.se."},
		{"gov.uk", "test.gov.uk", "test.gov.uk."},
		{"gov.uk-fqdn", "test.gov.uk.", "test.gov.uk."},
		{"long", "ndioqudh89u2inwref98hsd129d834fby.net", "ndioqudh89u2inwref98hsd129d834fby.net."},
		{"long-with-many-dots", "......ndioqudh89u2inwref98hsd129d834fby.net....", "ndioqudh89u2inwref98hsd129d834fby.net."},
		{"reverse", "a.9.7.1.4.1.8.6.0.0.0.0.0.0.0.0.0.0.0.0.0.1.0.0.0.0.7.4.6.0.6.2.ip6.arpa", "2.ip6.arpa."},
		{"reverse-fqdn", "a.9.7.1.4.1.8.6.0.0.0.0.0.0.0.0.0.0.0.0.0.1.0.0.0.0.7.4.6.0.6.2.ip6.arpa.", "2.ip6.arpa."},
		{"web-example", "www.example.com", "example.com."},
		{"web-example-fqdn", "www.example.com.", "example.com."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := GetETLDPlusOne(tt.indata)
			if got != tt.expected {
				t.Fatalf("got '%s', expected '%s'", got, tt.expected)
			}
		})
	}
}

func TestGetETLDPlusOneError(t *testing.T) {
	var tests = []struct {
		name        string
		indata      string
		errExpected bool
	}{
		{"root", ".", true},
		{"empty", "", true},
		{"many-dots", ".........", true},
		{"lan", "test.lan", false},
		{"lan-fqdn", "test.lan.", false},
		{"single", "org", true},
		{"single-fqdn", "org.", true},
		{"internal", "test.internal", false},
		{"internal-fqdn", "test.internal.", false},
		{"se", "test.se", false},
		{"se-fqdn", "test.se.", false},
		{"gov.uk", "test.gov.uk", false},
		{"gov.uk-fqdn", "test.gov.uk.", false},
		{"long", "ndioqudh89u2inwref98hsd129d834fby.net", false},
		{"long-with-many-dots", "......ndioqudh89u2inwref98hsd129d834fby.net....", false},
		{"reverse", "a.9.7.1.4.1.8.6.0.0.0.0.0.0.0.0.0.0.0.0.0.1.0.0.0.0.7.4.6.0.6.2.ip6.arpa", false},
		{"reverse-fqdn", "a.9.7.1.4.1.8.6.0.0.0.0.0.0.0.0.0.0.0.0.0.1.0.0.0.0.7.4.6.0.6.2.ip6.arpa.", false},
		{"web-example", "www.example.com", false},
		{"web-example-fqdn", "www.example.com.", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := GetETLDPlusOne(tt.indata)
			if (err != nil) != tt.errExpected {
				t.Fatalf("got '%s'. Was error expected? %t", err, tt.errExpected)
			}
		})
	}
}
