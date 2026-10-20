package cli

import "testing"

func TestRunSplitWithExactShares(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"exact plus equal rest", []string{"split", "30.00", "ana=12.00", "bob", "cara"}, "ana: 12.00\nbob: 9.00\ncara: 9.00\ntotal: 30.00\n"},
		{"exact plus weighted rest", []string{"split", "10.01", "ana=0.01", "bob*2", "cara"}, "ana: 0.01\nbob: 6.67\ncara: 3.33\ntotal: 10.01\n"},
		{"exact equals amount", []string{"split", "10", "ana=10", "bob"}, "ana: 10.00\nbob: 0.00\ntotal: 10.00\n"},
		{"everybody exact", []string{"split", "10.00", "ana=3.00", "bob=7.00"}, "ana: 3.00\nbob: 7.00\ntotal: 10.00\n"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			code, stdout, stderr := runCLI(test.args, "")
			if code != 0 || stdout != test.want || stderr != "" {
				t.Fatalf("Run(%q) = (%d, %q, %q), want (0, %q, empty)",
					test.args, code, stdout, stderr, test.want)
			}
		})
	}
}

func TestRunSplitExactShareErrors(t *testing.T) {
	tests := []struct {
		args    []string
		message string
	}{
		{[]string{"split", "10.00", "ana=4.00", "bob=7.00"}, "exact shares add up to 11.00, more than 10.00"},
		{[]string{"split", "10.00", "ana=4.00", "bob=5.00"}, "exact shares add up to 9.00, not 10.00"},
		{[]string{"split", "10", "ana=0", "bob"}, "invalid amount \"ana=0\""},
		{[]string{"split", "10", "ana*2=5", "bob"}, "invalid name \"ana*2\""},
		{[]string{"split", "10", "ana=5", "ana=5"}, "duplicate person \"ana\""},
	}
	for _, test := range tests {
		assertCLIError(t, test.args, "", test.message)
		code, stdout, stderr := runCLI(test.args, "")
		wantErr := "centsplit: " + test.message + "\n"
		if code != 2 || stdout != "" || stderr != wantErr {
			t.Fatalf("Run(%q) = (%d, %q, %q), want (2, empty, %q)",
				test.args, code, stdout, stderr, wantErr)
		}
	}
}

func TestRunBalanceWithExactShares(t *testing.T) {
	const twoLines = "ana paid 60.00 for ana=24.00 bob cara\nbob paid 20.00 for ana bob\n"
	const threeLines = twoLines + "cara paid 9.00 for ana=4.50 cara=4.50\n"
	tests := []struct {
		name  string
		args  []string
		input string
		want  string
	}{
		{"stdin balance", []string{"balance", "-"}, twoLines, "ana: +26.00\nbob: -8.00\ncara: -18.00\n"},
		{"three-line balance", []string{"balance", "-"}, threeLines, "ana: +21.50\nbob: -8.00\ncara: -13.50\n"},
		{"three-line settle", []string{"settle", "-"}, threeLines, "cara pays ana: 13.50\nbob pays ana: 8.00\n2 transfers\n"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			code, stdout, stderr := runCLI(test.args, test.input)
			if code != 0 || stdout != test.want || stderr != "" {
				t.Fatalf("Run(%q) = (%d, %q, %q), want (0, %q, empty)",
					test.args, code, stdout, stderr, test.want)
			}
		})
	}
	t.Run("file error", func(t *testing.T) {
		assertCLIError(t, []string{"balance", "-"}, "ana paid 5 for ana=3 bob=3",
			"<stdin>:1: exact shares add up to 6.00, more than 5.00")
		code, stdout, stderr := runCLI([]string{"balance", "-"}, "ana paid 5 for ana=3 bob=3")
		wantErr := "centsplit: <stdin>:1: exact shares add up to 6.00, more than 5.00\n"
		if code != 2 || stdout != "" || stderr != wantErr {
			t.Fatalf("got (%d, %q, %q), want (2, empty, %q)", code, stdout, stderr, wantErr)
		}
	})
}
