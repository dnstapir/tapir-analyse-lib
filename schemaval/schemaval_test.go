package schemaval

import (
	"errors"
	"testing"

	"github.com/dnstapir/tapir-analyse-lib/common"
)

func TestCreateEmptyPaths(t *testing.T) {
	var tests = []struct {
		name     string
		in1      bool
		in2      bool
		expected error
	}{
		{"false-false", false, false, common.ErrBadParam},
		{"false-true", false, true, common.ErrBadParam},
		{"true-false", true, false, common.ErrBadParam},
		{"true-true", true, true, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Conf{
				AllowNoSchema:  tt.in1,
				AllowNoVerKeys: tt.in2,
			}

			_, err := Create(c)
			if !errors.Is(err, tt.expected) {
				t.Fatalf("Expected: %s, Got: %s", tt.expected, err)
			}
		})
	}
}

func TestCreateSigningKeys(t *testing.T) {
	var tests = []struct {
		name     string
		in1      string
		expected error
	}{
		{"good", "testdata/signkeys/a.json", nil},
		{"noalg", "testdata/signkeys/noalg.json", common.ErrBadJWK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Conf{
				AllowNoSchema:  true,
				AllowNoVerKeys: true,
				SigningKey:     tt.in1,
			}

			_, err := Create(c)
			if !errors.Is(err, tt.expected) {
				t.Fatalf("Expected: %s, Got: %s", tt.expected, err)
			}
		})
	}
}

func TestSign(t *testing.T) {
	var tests = []struct {
		name     string
		indata   []byte
		expected []byte
	}{
		{"simple go-jose v4.1.3", testdataSimple, testdataSimpleSigned},
		{"new_qname go-jose v4.1.3", testdataNewQnameGood, testdataNewQnameGoodSigned},
	}

	c := Conf{
		SigningKey:     "testdata/signkeys/a.json",
		AllowNoSchema:  true,
		AllowNoVerKeys: true,
	}

	s, err := Create(c)
	if err != nil {
		t.Fatalf("Couldn't create handle: %s", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := s.SignData(tt.indata)
			if err != nil {
				t.Fatalf("Couldn't sign data: %s", err)
			}

			if string(got) != string(tt.expected) {
				t.Fatalf("Expected: %s, Got: %s", string(tt.expected), string(got))
			}
		})
	}
}

func TestVerifySignatureExternal(t *testing.T) {
	var tests = []struct {
		name     string
		indata   []byte
		expected []byte
	}{
		{"simple go-jose v4.1.3", testdataSimpleSigned, testdataSimple},
		{"new_qname go-jose v4.1.3", testdataNewQnameGoodSigned, testdataNewQnameGood},
	}

	c := Conf{
		VerificationKeys: "testdata/verkeys/a.pub.json",
		AllowNoSchema:    true,
	}

	s, err := Create(c)
	if err != nil {
		t.Fatalf("Couldn't create handle: %s", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			/* Verify a signature created with a different JOSE/JWS library using the same input data */
			got, err := s.VerifySignature(tt.indata)
			if err != nil {
				t.Fatalf("Couldn't verify signature: %s", err)
			}

			if string(got) != string(tt.expected) {
				t.Fatalf("Expected: %s, Got: %s", string(tt.expected), string(got))
			}
		})
	}
}

func TestVerifySignatureNoVerkeys(t *testing.T) {
	c := Conf{
		AllowNoSchema:  true,
		AllowNoVerKeys: true,
	}

	s, err := Create(c)
	if err != nil {
		t.Fatalf("Couldn't create handle: %s", err)
	}

	_, err = s.VerifySignature(testdataSimpleSigned)
	if !errors.Is(err, common.ErrNotCompleted) {
		t.Fatalf("Expected: %s, Got: %s", common.ErrNotCompleted, err)
	}
}

func TestValidateWithSchema(t *testing.T) {
	var tests = []struct {
		name     string
		indata   []byte
		schema   string
		expected bool
	}{
		{"accept_all", testdataSimple, "https://schema.dnstapir.se/v1/accept-all", true},
		{"edge-observation-bad", testdataSimple, "https://schema.dnstapir.se/v1/core_observation", false},
		{"new-qname-bad", testdataSimple, "https://schema.dnstapir.se/v1/new_qname", false},
	}

	c := Conf{
		SchemaDir:      "testdata/schemas/",
		AllowNoVerKeys: true,
	}

	s, err := Create(c)
	if err != nil {
		t.Fatalf("Couldn't create handle: %s", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := s.ValidateWithID(tt.indata, tt.schema)
			if got != tt.expected {
				t.Fatalf("got '%t', expected '%t'", got, tt.expected)
			}
		})
	}
}
