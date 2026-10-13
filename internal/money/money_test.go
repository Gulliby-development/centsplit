package money

import "testing"

func TestParseAmount(t *testing.T) {
	tests := []struct {
		name  string
		text  string
		cents int64
		ok    bool
	}{
		{name: "whole amount", text: "12", cents: 1200, ok: true},
		{name: "one fractional digit", text: "12.5", cents: 1250, ok: true},
		{name: "two fractional digits", text: "12.50", cents: 1250, ok: true},
		{name: "leading zero before decimal", text: "0.05", cents: 5, ok: true},
		{name: "whole-number leading zeros", text: "007", cents: 700, ok: true},
		{name: "fraction with leading zero whole part", text: "000000000.01", cents: 1, ok: true},
		{name: "fraction ending in zero", text: "1.0", cents: 100, ok: true},
		{name: "smallest nonzero amount", text: "0.01", cents: 1, ok: true},
		{name: "nine whole digits", text: "999999999", cents: 99999999900, ok: true},
		{name: "maximum amount", text: "999999999.99", cents: 99999999999, ok: true},

		{name: "empty", text: "", ok: false},
		{name: "zero whole amount", text: "0", ok: false},
		{name: "zero with one fractional digit", text: "0.0", ok: false},
		{name: "zero with two fractional digits", text: "0.00", ok: false},
		{name: "missing whole digits", text: ".5", ok: false},
		{name: "missing fractional digits", text: "5.", ok: false},
		{name: "too many fractional digits", text: "1.234", ok: false},
		{name: "comma decimal separator", text: "1,5", ok: false},
		{name: "negative sign", text: "-5", ok: false},
		{name: "positive sign", text: "+5", ok: false},
		{name: "ten whole digits", text: "1234567890", ok: false},
		{name: "ten whole digits with fraction", text: "1234567890.00", ok: false},
		{name: "multiple decimal points", text: "1..0", ok: false},
		{name: "leading whitespace", text: " 1", ok: false},
		{name: "trailing whitespace", text: "1 ", ok: false},
		{name: "non-ASCII digit", text: "１", ok: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gotCents, gotOK := ParseAmount(test.text)
			if gotOK != test.ok {
				t.Fatalf("ParseAmount(%q) ok = %t, want %t", test.text, gotOK, test.ok)
			}
			if gotOK && gotCents != test.cents {
				t.Errorf("ParseAmount(%q) cents = %d, want %d", test.text, gotCents, test.cents)
			}
		})
	}
}

func TestFormat(t *testing.T) {
	tests := []struct {
		name  string
		cents int64
		sign  bool
		want  string
	}{
		{name: "zero without sign", cents: 0, want: "0.00"},
		{name: "zero does not gain plus sign", cents: 0, sign: true, want: "0.00"},
		{name: "one cent", cents: 1, want: "0.01"},
		{name: "positive one cent with sign", cents: 1, sign: true, want: "+0.01"},
		{name: "single digit fractional part", cents: 10, want: "0.10"},
		{name: "two digit fractional part", cents: 99, want: "0.99"},
		{name: "whole amount", cents: 100, want: "1.00"},
		{name: "positive amount", cents: 250, want: "2.50"},
		{name: "positive amount with sign", cents: 250, sign: true, want: "+2.50"},
		{name: "negative amount", cents: -250, want: "-2.50"},
		{name: "negative amount ignores plus flag", cents: -250, sign: true, want: "-2.50"},
		{name: "negative sub-dollar amount", cents: -1, want: "-0.01"},
		{name: "maximum parsed amount", cents: 99999999999, want: "999999999.99"},
		{name: "maximum int64", cents: 1<<63 - 1, want: "92233720368547758.07"},
		{name: "minimum int64", cents: -1 << 63, want: "-92233720368547758.08"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := Format(test.cents, test.sign); got != test.want {
				t.Errorf("Format(%d, %t) = %q, want %q", test.cents, test.sign, got, test.want)
			}
		})
	}
}
