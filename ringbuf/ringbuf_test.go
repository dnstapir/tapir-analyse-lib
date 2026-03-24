package ringbuf

import (
	"testing"
)

var e struct{}

func TestAddCheckContents(t *testing.T) {
	var tests = []struct {
		name     string
		inSize   int
		inData   []string
		expected map[string]struct{}
	}{
		{"basic", 4, []string{"a", "b", "c"}, map[string]struct{}{"a": e, "b": e, "c": e}},
		{"wrap", 2, []string{"a", "b", "c"}, map[string]struct{}{"b": e, "c": e}},
		{"equal", 3, []string{"a", "b", "c"}, map[string]struct{}{"a": e, "b": e, "c": e}},
		{"empty", 3, []string{}, map[string]struct{}{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Conf{
				Size: tt.inSize,
			}

			r, err := Create[string](c)
			if err != nil {
				t.Fatalf("Unexpected error: %s", err)
			}

			for _, v := range tt.inData {
				r.Add(v)
			}

			got := r.Contents()

			if len(got) != len(tt.expected) {
				t.Fatalf("Expected: %d, Got: %d", len(tt.expected), len(got))
			}

			for _, data := range got {
				_, ok := tt.expected[data]
				if !ok {
					t.Fatalf("Expected contents missing '%s'", data)
				}
			}
		})
	}
}
