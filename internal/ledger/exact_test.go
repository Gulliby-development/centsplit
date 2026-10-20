package ledger

import (
	"reflect"
	"testing"
)

func TestParsePeopleExactShares(t *testing.T) {
	tests := []struct {
		name   string
		tokens []string
		want   []Person
	}{
		{
			name:   "exact, weighted and plain together",
			tokens: []string{"ana=12.00", "bob*2", "cara"},
			want:   []Person{{Name: "ana", Weight: 1, Exact: 1200}, {Name: "bob", Weight: 2}, {Name: "cara", Weight: 1}},
		},
		{name: "smallest amount", tokens: []string{"ana=0.01"}, want: []Person{{Name: "ana", Weight: 1, Exact: 1}}},
		{
			name:   "whole amount without decimals",
			tokens: []string{"ana=10", "bob"},
			want:   []Person{{Name: "ana", Weight: 1, Exact: 1000}, {Name: "bob", Weight: 1}},
		},
		{
			name:   "one and two decimals",
			tokens: []string{"ana=4.5", "cara=4.50"},
			want:   []Person{{Name: "ana", Weight: 1, Exact: 450}, {Name: "cara", Weight: 1, Exact: 450}},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := ParsePeople(test.tokens)
			if err != nil {
				t.Fatalf("ParsePeople(%q) returned unexpected error: %v", test.tokens, err)
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("ParsePeople(%q) = %#v, want %#v", test.tokens, got, test.want)
			}
		})
	}
}

func TestParsePeopleExactShareErrors(t *testing.T) {
	tests := []struct {
		name   string
		tokens []string
		err    string
	}{
		{"zero amount", []string{"ana=0", "bob"}, `invalid amount "ana=0"`},
		{"empty amount", []string{"ana=", "bob"}, `invalid amount "ana="`},
		{"three decimals", []string{"ana=1.234", "bob"}, `invalid amount "ana=1.234"`},
		{"weight and exact together", []string{"ana*2=5", "bob"}, `invalid name "ana*2"`},
		{"bad name character", []string{"b!b=5"}, `invalid name "b!b"`},
		{"duplicate exact tokens", []string{"ana=5", "ana=5"}, `duplicate person "ana"`},
		{"duplicate in mixed forms", []string{"ana=5", "ana"}, `duplicate person "ana"`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := ParsePeople(test.tokens)
			if err == nil || err.Error() != test.err {
				t.Fatalf("ParsePeople(%q) error = %v, want %q", test.tokens, err, test.err)
			}
			if got != nil {
				t.Errorf("ParsePeople(%q) returned %#v on error, want nil", test.tokens, got)
			}
		})
	}
}

func TestSharesWithExactParts(t *testing.T) {
	tests := []struct {
		name   string
		cents  int64
		people []Person
		want   []int64
	}{
		{"exact plus weighted rest", 1001, []Person{{"ana", 1, 1}, {"bob", 2, 0}, {"cara", 1, 0}}, []int64{1, 667, 333}},
		{"everybody exact", 1000, []Person{{"ana", 1, 300}, {"bob", 1, 700}}, []int64{300, 700}},
		{"wine example", 3000, []Person{{"ana", 1, 1200}, {"bob", 1, 0}, {"cara", 1, 0}}, []int64{1200, 900, 900}},
		{"exact equals amount", 1000, []Person{{"ana", 1, 1000}, {"bob", 1, 0}}, []int64{1000, 0}},
		{"tie-break among non-exact only", 1002, []Person{{"ana", 1, 1}, {"cara", 1, 0}, {"bob", 1, 0}}, []int64{1, 500, 501}},
		{"no exact share", 1000, []Person{{"ana", 1, 0}, {"bob", 1, 0}, {"cara", 1, 0}}, []int64{334, 333, 333}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := Shares(test.cents, test.people)
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("Shares(%d, %#v) = %v, want %v", test.cents, test.people, got, test.want)
			}
		})
	}
}

func TestCheckExact(t *testing.T) {
	tests := []struct {
		name   string
		cents  int64
		people []Person
		err    string
	}{
		{"everybody exact, too much", 1000, []Person{{"ana", 1, 400}, {"bob", 1, 700}}, "exact shares add up to 11.00, more than 10.00"},
		{"everybody exact, too little", 1000, []Person{{"ana", 1, 400}, {"bob", 1, 500}}, "exact shares add up to 9.00, not 10.00"},
		{"file example sum", 500, []Person{{"ana", 1, 300}, {"bob", 1, 300}}, "exact shares add up to 6.00, more than 5.00"},
		{"too much with a non-exact person", 1000, []Person{{"ana", 1, 1100}, {"bob", 1, 0}}, "exact shares add up to 11.00, more than 10.00"},
		{"everybody exact, exact total", 1000, []Person{{"ana", 1, 400}, {"bob", 1, 600}}, ""},
		{"one person without exact share", 1000, []Person{{"ana", 1, 400}, {"bob", 1, 0}}, ""},
		{"exact sum equals amount", 1000, []Person{{"ana", 1, 1000}, {"bob", 1, 0}}, ""},
		{"no exact share", 1000, []Person{{"ana", 1, 0}, {"bob", 2, 0}}, ""},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := CheckExact(test.cents, test.people)
			if test.err == "" {
				if err != nil {
					t.Errorf("CheckExact(%d, %#v) = %v, want nil", test.cents, test.people, err)
				}
				return
			}
			if err == nil || err.Error() != test.err {
				t.Errorf("CheckExact(%d, %#v) = %v, want %q", test.cents, test.people, err, test.err)
			}
		})
	}
}

func TestParseFileExactSharesTotals(t *testing.T) {
	records, err := ParseFile("x.txt", "ana paid 60.00 for ana=24.00 bob cara\n")
	if err != nil {
		t.Fatalf("ParseFile returned unexpected error: %v", err)
	}
	want := []Record{{
		Kind:   "paid",
		From:   "ana",
		Cents:  6000,
		People: []Person{{Name: "ana", Weight: 1, Exact: 2400}, {Name: "bob", Weight: 1}, {Name: "cara", Weight: 1}},
	}}
	if !reflect.DeepEqual(records, want) {
		t.Errorf("ParseFile = %#v, want %#v", records, want)
	}

	records, err = ParseFile("x.txt", "ana paid 5 for ana=3 bob=3\n")
	wantErr := "x.txt:1: exact shares add up to 6.00, more than 5.00"
	if err == nil || err.Error() != wantErr {
		t.Fatalf("ParseFile error = %v, want %q", err, wantErr)
	}
	if records != nil {
		t.Errorf("ParseFile returned %#v on error, want nil", records)
	}
}
