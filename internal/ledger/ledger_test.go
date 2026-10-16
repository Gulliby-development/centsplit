package ledger

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestParsePeople(t *testing.T) {
	tests := []struct {
		name   string
		tokens []string
		want   []Person
		err    string
	}{
		{
			name:   "default and explicit weights",
			tokens: []string{"alice", "Bob-2", "u_9*99"},
			want:   []Person{{Name: "alice", Weight: 1}, {Name: "Bob-2", Weight: 1}, {Name: "u_9", Weight: 99}},
		},
		{name: "empty list", tokens: []string{}, want: []Person{}},
		{name: "minimum weight", tokens: []string{"a*1"}, want: []Person{{Name: "a", Weight: 1}}},
		{name: "maximum name length", tokens: []string{"abcdefghijklmnopqrst"}, want: []Person{{Name: "abcdefghijklmnopqrst", Weight: 1}}},
		{name: "name too long", tokens: []string{"abcdefghijklmnopqrstu"}, err: "invalid name \"abcdefghijklmnopqrstu\""},
		{name: "empty name", tokens: []string{"*2"}, err: "invalid name \"\""},
		{name: "invalid name character", tokens: []string{"bo!b"}, err: "invalid name \"bo!b\""},
		{name: "weight zero", tokens: []string{"bob*0"}, err: "invalid weight \"bob*0\""},
		{name: "weight leading zero", tokens: []string{"bob*07"}, err: "invalid weight \"bob*07\""},
		{name: "weight too large", tokens: []string{"bob*100"}, err: "invalid weight \"bob*100\""},
		{name: "empty weight", tokens: []string{"bob*"}, err: "invalid weight \"bob*\""},
		{name: "weight with extra separator", tokens: []string{"bob*1*2"}, err: "invalid weight \"bob*1*2\""},
		{name: "duplicate plain name", tokens: []string{"bob", "bob"}, err: "duplicate person \"bob\""},
		{name: "duplicate with weight", tokens: []string{"bob*2", "bob*3"}, err: "duplicate person \"bob\""},
		{name: "case-sensitive names", tokens: []string{"Ana", "ana"}, want: []Person{{Name: "Ana", Weight: 1}, {Name: "ana", Weight: 1}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := ParsePeople(test.tokens)
			if test.err != "" {
				if err == nil || err.Error() != test.err {
					t.Fatalf("ParsePeople(%q) error = %v, want %q", test.tokens, err, test.err)
				}
				if got != nil {
					t.Fatalf("ParsePeople(%q) returned %#v on error, want nil", test.tokens, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParsePeople(%q) returned unexpected error: %v", test.tokens, err)
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Fatalf("ParsePeople(%q) = %#v, want %#v", test.tokens, got, test.want)
			}
		})
	}
}

func TestParseFileRecordsAndWhitespace(t *testing.T) {
	input := "\uFEFF# trip records\r\n\r\n \tana\tpaid\t12.5\tfor\tana\tbob*2 \t\r\n\t# comment\r\nbob sent cara 0.05\r\n"
	got, err := ParseFile("trip.txt", input)
	if err != nil {
		t.Fatalf("ParseFile returned unexpected error: %v", err)
	}
	want := []Record{
		{Kind: "paid", From: "ana", Cents: 1250, People: []Person{{Name: "ana", Weight: 1}, {Name: "bob", Weight: 2}}},
		{Kind: "sent", From: "bob", To: "cara", Cents: 5},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ParseFile records = %#v, want %#v", got, want)
	}
}

func TestParseFileErrors(t *testing.T) {
	tests := []struct {
		name  string
		input string
		line  int
		msg   string
	}{
		{name: "invalid first name", input: "!x paid 1 for a", line: 1, msg: "invalid name \"!x\""},
		{name: "missing verb", input: "ana", line: 1, msg: "missing verb"},
		{name: "unknown verb", input: "ana owes 5 for bob", line: 1, msg: "unknown verb \"owes\""},
		{name: "paid missing amount", input: "ana paid", line: 1, msg: "missing amount"},
		{name: "paid invalid zero amount", input: "ana paid 0 for bob", line: 1, msg: "invalid amount \"0\""},
		{name: "paid invalid amount form", input: "ana paid 1.234 for bob", line: 1, msg: "invalid amount \"1.234\""},
		{name: "paid missing for", input: "ana paid 5", line: 1, msg: "missing \"for\""},
		{name: "paid wrong for word", input: "ana paid 5 bob", line: 1, msg: "expected \"for\", found \"bob\""},
		{name: "paid without people", input: "ana paid 5 for", line: 1, msg: "no people after \"for\""},
		{name: "invalid person name", input: "ana paid 5 for bo!b", line: 1, msg: "invalid name \"bo!b\""},
		{name: "empty person name", input: "ana paid 5 for *2", line: 1, msg: "invalid name \"\""},
		{name: "invalid weight zero", input: "ana paid 5 for bob*0", line: 1, msg: "invalid weight \"bob*0\""},
		{name: "invalid weight three digits", input: "ana paid 5 for bob*100", line: 1, msg: "invalid weight \"bob*100\""},
		{name: "invalid weight leading zero", input: "ana paid 5 for bob*007", line: 1, msg: "invalid weight \"bob*007\""},
		{name: "invalid weight empty", input: "ana paid 5 for bob*", line: 1, msg: "invalid weight \"bob*\""},
		{name: "duplicate person", input: "ana paid 5 for bob bob", line: 1, msg: "duplicate person \"bob\""},
		{name: "sent missing recipient", input: "ana sent", line: 1, msg: "missing recipient"},
		{name: "sent invalid recipient", input: "ana sent !bob 1", line: 1, msg: "invalid name \"!bob\""},
		{name: "sent to self", input: "ana sent ana 5", line: 1, msg: "cannot send to the same person"},
		{name: "sent missing amount", input: "ana sent bob", line: 1, msg: "missing amount"},
		{name: "sent invalid amount", input: "ana sent bob 1,5", line: 1, msg: "invalid amount \"1,5\""},
		{name: "sent unexpected word", input: "ana sent bob 5 now", line: 1, msg: "unexpected \"now\""},
		{name: "error line counts blank lines", input: "ana paid 1 for bob\n\n!x paid 1 for a", line: 3, msg: "invalid name \"!x\""},
		{name: "hash inside a word is part of the name", input: "ana paid 1 for bob#1 ana", line: 1, msg: "invalid name \"bob#1\""},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			records, err := ParseFile("e.txt", test.input)
			want := fmt.Sprintf("e.txt:%d: %s", test.line, test.msg)
			if err == nil || err.Error() != want {
				t.Fatalf("ParseFile error = %v, want %q", err, want)
			}
			if records != nil {
				t.Fatalf("ParseFile returned %#v on error, want nil", records)
			}
		})
	}
}

func TestParseFileEmptyAndRecordLimit(t *testing.T) {
	records, err := ParseFile("empty.txt", "\n \t\n# comment\n")
	if err != nil {
		t.Fatalf("ParseFile on empty content returned error: %v", err)
	}
	if !reflect.DeepEqual(records, []Record{}) {
		t.Fatalf("ParseFile on empty content = %#v, want an empty record slice", records)
	}

	validLine := "a paid 1 for b"
	exactlyLimit := strings.TrimSuffix(strings.Repeat(validLine+"\n", 1000), "\n")
	records, err = ParseFile("many.txt", exactlyLimit)
	if err != nil {
		t.Fatalf("ParseFile rejected 1000 records: %v", err)
	}
	if len(records) != 1000 {
		t.Fatalf("ParseFile returned %d records, want 1000", len(records))
	}

	overLimit := exactlyLimit + "\n\n! this 1001st record is invalid"
	records, err = ParseFile("many.txt", overLimit)
	if err == nil || err.Error() != "many.txt:1002: more than 1000 records" {
		t.Fatalf("ParseFile error = %v, want %q", err, "many.txt:1002: more than 1000 records")
	}
	if records != nil {
		t.Fatalf("ParseFile returned %#v on record limit error, want nil", records)
	}

	overLimit = strings.TrimSuffix(strings.Repeat(validLine+"\n", 1000), "\n") + "\n! this 1001st record is invalid"
	records, err = ParseFile("many.txt", overLimit)
	if err == nil || err.Error() != "many.txt:1001: more than 1000 records" {
		t.Fatalf("ParseFile error = %v, want %q", err, "many.txt:1001: more than 1000 records")
	}
}

func TestSharesExamples(t *testing.T) {
	tests := []struct {
		name   string
		cents  int64
		people []Person
		want   []int64
	}{
		{
			name:   "equal thirds with name tie break",
			cents:  1000,
			people: []Person{{Name: "zed", Weight: 1}, {Name: "amy", Weight: 1}, {Name: "bo", Weight: 1}},
			want:   []int64{333, 334, 333},
		},
		{
			name:   "five cents with a double weight",
			cents:  5,
			people: []Person{{Name: "ana", Weight: 1}, {Name: "bob", Weight: 2}, {Name: "cara", Weight: 1}},
			want:   []int64{1, 3, 1},
		},
		{
			name:   "weighted half third and sixth",
			cents:  100,
			people: []Person{{Name: "a", Weight: 3}, {Name: "b", Weight: 2}, {Name: "c", Weight: 1}},
			want:   []int64{50, 33, 17},
		},
		{
			name:   "four equal weights tie by byte-order name",
			cents:  1001,
			people: []Person{{Name: "ana", Weight: 1}, {Name: "bob", Weight: 1}, {Name: "cara", Weight: 1}, {Name: "dave", Weight: 1}},
			want:   []int64{251, 250, 250, 250},
		},
		{
			name:   "large amount and unequal weights",
			cents:  99999999999,
			people: []Person{{Name: "a", Weight: 99}, {Name: "b", Weight: 98}, {Name: "c", Weight: 97}, {Name: "d", Weight: 1}},
			want:   []int64{33559322033, 33220338983, 32881355932, 338983051},
		},
		{name: "zero amount", cents: 0, people: []Person{{Name: "a", Weight: 1}, {Name: "b", Weight: 2}}, want: []int64{0, 0}},
		{name: "no people", cents: 100, people: []Person{}, want: []int64{}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := Shares(test.cents, test.people)
			if !reflect.DeepEqual(got, test.want) {
				t.Fatalf("Shares(%d, %#v) = %#v, want %#v", test.cents, test.people, got, test.want)
			}
		})
	}
}

func TestSharesConserveCents(t *testing.T) {
	people := [][]Person{
		{{Name: "a", Weight: 1}},
		{{Name: "a", Weight: 1}, {Name: "b", Weight: 1}},
		{{Name: "z", Weight: 99}, {Name: "a", Weight: 98}, {Name: "m", Weight: 97}, {Name: "b", Weight: 1}},
		{{Name: "alice", Weight: 17}, {Name: "bob", Weight: 4}, {Name: "carol", Weight: 23}, {Name: "dan", Weight: 1}, {Name: "eve", Weight: 99}},
	}
	amounts := []int64{0, 1, 2, 5, 99, 100, 1001, 99999999999}
	for _, group := range people {
		for _, amount := range amounts {
			shares := Shares(amount, group)
			if len(shares) != len(group) {
				t.Fatalf("Shares(%d, %#v) returned %d shares, want %d", amount, group, len(shares), len(group))
			}
			var total int64
			for _, share := range shares {
				total += share
			}
			if total != amount {
				t.Fatalf("Shares(%d, %#v) sum to %d", amount, group, total)
			}
		}
	}
}

func TestBalancesExamplesAndConservation(t *testing.T) {
	tests := []struct {
		name string
		text string
		want map[string]int64
	}{
		{
			name: "trip example",
			text: "ana paid 90.00 for ana bob cara\n" +
				"bob paid 30.50 for ana bob\n" +
				"cara paid 12 for bob*2 cara\n" +
				"dave paid 10.01 for ana bob cara dave\n" +
				"bob sent ana 10.00\n",
			want: map[string]int64{"ana": 3224, "bob": -1525, "cara": -2450, "dave": 751},
		},
		{
			name: "three equal debtors",
			text: "amy paid 10.00 for amy bo cy dee",
			want: map[string]int64{"amy": 750, "bo": -250, "cy": -250, "dee": -250},
		},
		{
			name: "two independent groups",
			text: "p1 paid 8 for p1 d1 d2\nq1 paid 3 for q1 e1 e2 e3",
			want: map[string]int64{"d1": -267, "d2": -267, "e1": -75, "e2": -75, "e3": -75, "p1": 534, "q1": 225},
		},
		{
			name: "repayment nets balances to zero",
			text: "ana paid 5 for bob\nbob sent ana 5",
			want: map[string]int64{"ana": 0, "bob": 0},
		},
		{
			name: "payer may not participate",
			text: "payer paid 1 for person",
			want: map[string]int64{"payer": 100, "person": -100},
		},
		{
			name: "BOM CRLF and tabbed weighted expense",
			text: "\uFEFFx paid 1.00 for y\r\n\r\n  # c\r\n\ty paid 2 for x*2 y\r\n",
			want: map[string]int64{"x": -33, "y": 33},
		},
		{
			name: "large weighted expense",
			text: "ana paid 999999999.99 for ana*99 bob*98 cara*97\nbob paid 0.01 for ana",
			want: map[string]int64{"ana": 66326530611, "bob": -33333333332, "cara": -32993197279},
		},
		{name: "no records", text: "", want: map[string]int64{}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			records, err := ParseFile("balances.txt", test.text)
			if err != nil {
				t.Fatalf("ParseFile returned unexpected error: %v", err)
			}
			got := Balances(records)
			if !reflect.DeepEqual(got, test.want) {
				t.Fatalf("Balances(ParseFile(%q)) = %#v, want %#v", test.text, got, test.want)
			}
			var total int64
			for _, balance := range got {
				total += balance
			}
			if total != 0 {
				t.Fatalf("balances sum to %d, want 0", total)
			}
		})
	}
}

func TestSortedNames(t *testing.T) {
	got := SortedNames(map[string]int64{"b": 0, "a": 1, "A": -1, "aa": 2})
	want := []string{"A", "a", "aa", "b"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("SortedNames returned %#v, want %#v", got, want)
	}
	if got := SortedNames(map[string]int64{}); !reflect.DeepEqual(got, []string{}) {
		t.Fatalf("SortedNames(empty) = %#v, want an empty slice", got)
	}
}

func TestSettleExamples(t *testing.T) {
	tests := []struct {
		name     string
		balances map[string]int64
		want     []Transfer
	}{
		{
			name:     "trip example",
			balances: map[string]int64{"ana": 3224, "bob": -1525, "cara": -2450, "dave": 751},
			want: []Transfer{
				{From: "cara", To: "ana", Cents: 2450},
				{From: "bob", To: "ana", Cents: 774},
				{From: "bob", To: "dave", Cents: 751},
			},
		},
		{
			name:     "three equal debtors",
			balances: map[string]int64{"amy": 750, "bo": -250, "cy": -250, "dee": -250},
			want: []Transfer{
				{From: "bo", To: "amy", Cents: 250},
				{From: "cy", To: "amy", Cents: 250},
				{From: "dee", To: "amy", Cents: 250},
			},
		},
		{
			name:     "two independent groups",
			balances: map[string]int64{"d1": -267, "d2": -267, "e1": -75, "e2": -75, "e3": -75, "p1": 534, "q1": 225},
			want: []Transfer{
				{From: "d1", To: "p1", Cents: 267},
				{From: "d2", To: "p1", Cents: 267},
				{From: "e1", To: "q1", Cents: 75},
				{From: "e2", To: "q1", Cents: 75},
				{From: "e3", To: "q1", Cents: 75},
			},
		},
		{
			name:     "BOM CRLF weighted settlement",
			balances: map[string]int64{"x": -33, "y": 33},
			want:     []Transfer{{From: "x", To: "y", Cents: 33}},
		},
		{
			name: "debtors and creditors are re-sorted after each transfer",
			balances: map[string]int64{
				"a": -1000, "b": -900,
				"x": 300, "y": 800, "z": 800,
			},
			want: []Transfer{
				{From: "a", To: "y", Cents: 800},
				{From: "b", To: "z", Cents: 800},
				{From: "a", To: "x", Cents: 200},
				{From: "b", To: "x", Cents: 100},
			},
		},
		{
			name:     "equal debtor and creditor amounts use name order",
			balances: map[string]int64{"b": -500, "a": -500, "z": 500, "c": 500},
			want: []Transfer{
				{From: "a", To: "c", Cents: 500},
				{From: "b", To: "z", Cents: 500},
			},
		},
		{name: "zero and empty balances", balances: map[string]int64{"a": 0, "b": 0}, want: []Transfer{}},
		{name: "empty map", balances: map[string]int64{}, want: []Transfer{}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			before := make(map[string]int64, len(test.balances))
			for name, balance := range test.balances {
				before[name] = balance
			}
			got := Settle(test.balances)
			if !reflect.DeepEqual(got, test.want) {
				t.Fatalf("Settle(%#v) = %#v, want %#v", test.balances, got, test.want)
			}
			if !reflect.DeepEqual(test.balances, before) {
				t.Fatalf("Settle modified balances: got %#v, want %#v", test.balances, before)
			}
		})
	}
}
