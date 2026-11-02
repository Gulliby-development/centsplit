package cli

import "testing"

func TestRunBalanceOnly(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"one name", []string{"balance", "--only", "ana", "-"}, "ana: +32.24\n"},
		{"two names", []string{"balance", "--only", "ana,dave", "-"}, "ana: +32.24\ndave: +7.51\n"},
		{"list order does not matter", []string{"balance", "--only", "dave,ana", "-"}, "ana: +32.24\ndave: +7.51\n"},
		{"equals form", []string{"balance", "--only=bob,cara", "-"}, "bob: -15.25\ncara: -24.50\n"},
		{"after the file", []string{"balance", "-", "--only", "cara"}, "cara: -24.50\n"},
		{"last one wins", []string{"balance", "--only", "ana", "--only=dave", "-"}, "dave: +7.51\n"},
		{"all names", []string{"balance", "--only", "ana,bob,cara,dave", "-"}, "ana: +32.24\nbob: -15.25\ncara: -24.50\ndave: +7.51\n"},
		{"names are case sensitive", []string{"balance", "--only", "dave", "-"}, "dave: +7.51\n"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			code, stdout, stderr := runCLI(test.args, tripLedger)
			if code != 0 || stdout != test.want || stderr != "" {
				t.Fatalf("Run(%q) = (%d, %q, %q), want (0, %q, empty)", test.args, code, stdout, stderr, test.want)
			}
		})
	}
}

func TestRunBalanceOnlyShowsZeroBalances(t *testing.T) {
	code, stdout, stderr := runCLI([]string{"balance", "--only", "bo", "-"}, "ana paid 2 for ana bo\nbo sent ana 1\n")
	if code != 0 || stdout != "bo: 0.00\n" || stderr != "" {
		t.Fatalf("Run = (%d, %q, %q), want (0, %q, empty)", code, stdout, stderr, "bo: 0.00\n")
	}
}

func TestRunBalanceOnlyErrors(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		input   string
		message string
	}{
		{"missing value", []string{"balance", "-", "--only"}, tripLedger, "missing value for --only"},
		{"empty value", []string{"balance", "--only", "", "-"}, tripLedger, "invalid name \"\""},
		{"empty item", []string{"balance", "--only", "ana,,bob", "-"}, tripLedger, "invalid name \"\""},
		{"trailing comma", []string{"balance", "--only=ana,", "-"}, tripLedger, "invalid name \"\""},
		{"bad character", []string{"balance", "--only", "ana,b!b", "-"}, tripLedger, "invalid name \"b!b\""},
		{"weight is not a name", []string{"balance", "--only", "ana*2", "-"}, tripLedger, "invalid name \"ana*2\""},
		{"name too long", []string{"balance", "--only", "abcdefghijklmnopqrstu", "-"}, tripLedger, "invalid name \"abcdefghijklmnopqrstu\""},
		{"duplicate", []string{"balance", "--only", "ana,bob,ana", "-"}, tripLedger, "duplicate person \"ana\""},
		{"unknown person", []string{"balance", "--only", "ana,zed,yan", "-"}, tripLedger, "unknown person \"zed\""},
		{"name case", []string{"balance", "--only", "Ana", "-"}, tripLedger, "unknown person \"Ana\""},
		{"empty ledger", []string{"balance", "--only", "ana", "-"}, "# nothing\n", "unknown person \"ana\""},
		{"ledger error first", []string{"balance", "--only", "zed", "-"}, "ana paid 1 for\n", "<stdin>:1: no people after \"for\""},
		{"value error before file error", []string{"balance", "--only", "a b"}, "", "invalid name \"a b\""},
		{"missing file", []string{"balance", "--only", "ana"}, "", "missing FILE"},
		{"two files", []string{"balance", "--only", "ana", "-", "-"}, tripLedger, "too many arguments"},
		{"settle refuses it", []string{"settle", "--only", "ana", "-"}, tripLedger, "unknown option \"--only\""},
		{"settle refuses the equals form", []string{"settle", "--only=ana", "-"}, tripLedger, "unknown option \"--only=ana\""},
		{"split refuses it", []string{"split", "--only", "ana", "10", "ana"}, "", "unknown option \"--only\""},
		{"other options stay unknown", []string{"balance", "--only", "ana", "--all", "-"}, tripLedger, "unknown option \"--all\""},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assertCLIError(t, test.args, test.input, test.message)
		})
	}
}
