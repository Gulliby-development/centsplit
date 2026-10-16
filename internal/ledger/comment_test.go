package ledger

import (
	"reflect"
	"strings"
	"testing"
)

func TestParseFileRecordLineComments(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []Record
	}{
		{
			name:  "paid record comment",
			input: "ana paid 30.00 for ana bob # dinner",
			want:  []Record{{Kind: "paid", From: "ana", Cents: 3000, People: []Person{{Name: "ana", Weight: 1}, {Name: "bob", Weight: 1}}}},
		},
		{
			name:  "sent record comment",
			input: "bob sent ana 5.00 #refund",
			want:  []Record{{Kind: "sent", From: "bob", To: "ana", Cents: 500}},
		},
		{
			name:  "tab before multiple comment words",
			input: "cara paid 9 for bob*2 cara\t# taxi # ride home",
			want:  []Record{{Kind: "paid", From: "cara", Cents: 900, People: []Person{{Name: "bob", Weight: 2}, {Name: "cara", Weight: 1}}}},
		},
		{
			name:  "bare hash and comment-only lines with CRLF",
			input: "# heading\r\n \t# note\r\nana paid 1 for bob #\r\n\r\n",
			want:  []Record{{Kind: "paid", From: "ana", Cents: 100, People: []Person{{Name: "bob", Weight: 1}}}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := ParseFile("comments.txt", test.input)
			if err != nil {
				t.Fatalf("ParseFile returned unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Fatalf("ParseFile records = %#v, want %#v", got, test.want)
			}
		})
	}
}

func TestParseFileCommentErrors(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "no people after for before comment",
			input: "# leading comment\n\nana paid 5 for # bob",
			want:  "comments.txt:3: no people after \"for\"",
		},
		{
			name:  "hash word does not satisfy for",
			input: "ana paid 5 #for bob",
			want:  "comments.txt:1: missing \"for\"",
		},
		{
			name:  "hash inside word remains part of the token",
			input: "ana paid 5 for bob#1 ana",
			want:  "comments.txt:1: invalid name \"bob#1\"",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			records, err := ParseFile("comments.txt", test.input)
			if err == nil || err.Error() != test.want {
				t.Fatalf("ParseFile error = %v, want %q", err, test.want)
			}
			if records != nil {
				t.Fatalf("ParseFile returned %#v on error, want nil", records)
			}
		})
	}

	_, err := ParseFile("comments.txt", "ana paid 5#1 for bob")
	if err == nil || err.Error() != "comments.txt:1: invalid amount \"5#1\"" {
		t.Fatalf("ParseFile amount error = %v, want %q", err, "comments.txt:1: invalid amount \"5#1\"")
	}
}

func TestParseFileCommentLinesDoNotCountTowardRecordLimit(t *testing.T) {
	commentLines := strings.Repeat("# ignored comment\n", 1000)
	records, err := ParseFile("comments.txt", commentLines)
	if err != nil {
		t.Fatalf("ParseFile rejected 1000 comment-only lines: %v", err)
	}
	if !reflect.DeepEqual(records, []Record{}) {
		t.Fatalf("ParseFile records = %#v, want an empty slice", records)
	}

	input := commentLines + strings.Repeat("a paid 1 for b # note\n", 1000)
	records, err = ParseFile("comments.txt", input)
	if err != nil {
		t.Fatalf("ParseFile rejected 1000 records after comment-only lines: %v", err)
	}
	if len(records) != 1000 {
		t.Fatalf("ParseFile returned %d records, want 1000", len(records))
	}

	input += "a paid 1 for b # 1001st record"
	records, err = ParseFile("comments.txt", input)
	if err == nil || err.Error() != "comments.txt:2001: more than 1000 records" {
		t.Fatalf("ParseFile over-limit error = %v, want %q", err, "comments.txt:2001: more than 1000 records")
	}
	if records != nil {
		t.Fatalf("ParseFile returned %#v on record limit error, want nil", records)
	}
}
