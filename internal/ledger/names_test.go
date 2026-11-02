package ledger

import (
	"reflect"
	"testing"
)

func TestParseNames(t *testing.T) {
	tests := []struct {
		list string
		want []string
	}{
		{"ana", []string{"ana"}},
		{"ana,bob", []string{"ana", "bob"}},
		{"bob,ana,cara_1,d-2", []string{"bob", "ana", "cara_1", "d-2"}},
		{"Ana,ana", []string{"Ana", "ana"}},
		{"abcdefghijklmnopqrst", []string{"abcdefghijklmnopqrst"}},
	}
	for _, test := range tests {
		got, err := ParseNames(test.list)
		if err != nil || !reflect.DeepEqual(got, test.want) {
			t.Errorf("ParseNames(%q) = (%q, %v), want (%q, nil)", test.list, got, err, test.want)
		}
	}
}

func TestParseNamesErrors(t *testing.T) {
	tests := []struct {
		list    string
		message string
	}{
		{"", "invalid name \"\""},
		{",ana", "invalid name \"\""},
		{"ana,", "invalid name \"\""},
		{"ana,,bob", "invalid name \"\""},
		{"ana bob", "invalid name \"ana bob\""},
		{"ana;bob", "invalid name \"ana;bob\""},
		{"ana*2", "invalid name \"ana*2\""},
		{"ana=5", "invalid name \"ana=5\""},
		{"abcdefghijklmnopqrstu", "invalid name \"abcdefghijklmnopqrstu\""},
		{"ana,bob,bob", "duplicate person \"bob\""},
		{"bob,x y,bob", "invalid name \"x y\""},
	}
	for _, test := range tests {
		got, err := ParseNames(test.list)
		if err == nil || err.Error() != test.message || got != nil {
			t.Errorf("ParseNames(%q) = (%q, %v), want error %q", test.list, got, err, test.message)
		}
	}
}
