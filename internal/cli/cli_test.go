package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const tripLedger = `# trip to the lake
ana paid 90.00 for ana bob cara
bob paid 30.50 for ana bob
cara paid 12 for bob*2 cara
dave paid 10.01 for ana bob cara dave
bob sent ana 10.00
`

func runCLI(args []string, input string) (int, string, string) {
	var stdout, stderr bytes.Buffer
	code := Run(args, strings.NewReader(input), &stdout, &stderr)
	return code, stdout.String(), stderr.String()
}

func assertCLIError(t *testing.T, args []string, input, message string) {
	t.Helper()
	code, stdout, stderr := runCLI(args, input)
	if code != 2 {
		t.Errorf("Run(%q) exit code = %d, want 2", args, code)
	}
	if stdout != "" {
		t.Errorf("Run(%q) stdout = %q, want empty", args, stdout)
	}
	want := "centsplit: " + message + "\n"
	if stderr != want {
		t.Errorf("Run(%q) stderr = %q, want %q", args, stderr, want)
	}
}

func TestRunHelp(t *testing.T) {
	code, stdout, stderr := runCLI([]string{"--help"}, "")
	if code != 0 {
		t.Errorf("Run(--help) exit code = %d, want 0", code)
	}
	const want = "usage: centsplit balance FILE\n       centsplit settle FILE\n       centsplit split AMOUNT PERSON...\n"
	if stdout != want {
		t.Errorf("Run(--help) stdout = %q, want %q", stdout, want)
	}
	if stderr != "" {
		t.Errorf("Run(--help) stderr = %q, want empty", stderr)
	}
}

func TestRunTripExamplesFromStandardInput(t *testing.T) {
	weightedLedger := "\uFEFFx paid 1.00 for y\r\n\r\n  # c\r\n\ty paid 2 for x*2 y\r\n"
	largeLedger := "ana paid 999999999.99 for ana*99 bob*98 cara*97\nbob paid 0.01 for ana"
	manyRecords := strings.Repeat("a paid 1 for b\n", 1000)
	independentLedgers := "p1 paid 8 for p1 d1 d2\nq1 paid 3 for q1 e1 e2 e3"
	amyLedger := "amy paid 10.00 for amy bo cy dee"
	tests := []struct {
		name  string
		args  []string
		input string
		want  string
	}{
		{
			name:  "trip balance",
			args:  []string{"balance", "-"},
			input: tripLedger,
			want:  "ana: +32.24\nbob: -15.25\ncara: -24.50\ndave: +7.51\n",
		},
		{
			name:  "trip settlement",
			args:  []string{"settle", "-"},
			input: tripLedger,
			want:  "cara pays ana: 24.50\nbob pays ana: 7.74\nbob pays dave: 7.51\n3 transfers\n",
		},
		{
			name:  "amy ledger balance",
			args:  []string{"balance", "-"},
			input: amyLedger,
			want:  "amy: +7.50\nbo: -2.50\ncy: -2.50\ndee: -2.50\n",
		},
		{
			name:  "amy ledger settlement",
			args:  []string{"settle", "-"},
			input: amyLedger,
			want:  "bo pays amy: 2.50\ncy pays amy: 2.50\ndee pays amy: 2.50\n3 transfers\n",
		},
		{
			name:  "independent ledgers balance",
			args:  []string{"balance", "-"},
			input: independentLedgers,
			want:  "d1: -2.67\nd2: -2.67\ne1: -0.75\ne2: -0.75\ne3: -0.75\np1: +5.34\nq1: +2.25\n",
		},
		{
			name:  "independent ledgers settlement",
			args:  []string{"settle", "-"},
			input: independentLedgers,
			want:  "d1 pays p1: 2.67\nd2 pays p1: 2.67\ne1 pays q1: 0.75\ne2 pays q1: 0.75\ne3 pays q1: 0.75\n5 transfers\n",
		},
		{
			name:  "BOM CRLF weighted ledger balance",
			args:  []string{"balance", "-"},
			input: weightedLedger,
			want:  "x: -0.33\ny: +0.33\n",
		},
		{
			name:  "BOM CRLF weighted ledger settlement",
			args:  []string{"settle", "-"},
			input: weightedLedger,
			want:  "x pays y: 0.33\n1 transfer\n",
		},
		{
			name:  "large amount ledger balance",
			args:  []string{"balance", "-"},
			input: largeLedger,
			want:  "ana: +663265306.11\nbob: -333333333.32\ncara: -329931972.79\n",
		},
		{
			name:  "1000-record ledger balance",
			args:  []string{"balance", "-"},
			input: manyRecords,
			want:  "a: +1000.00\nb: -1000.00\n",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			code, stdout, stderr := runCLI(test.args, test.input)
			if code != 0 || stdout != test.want || stderr != "" {
				t.Fatalf("Run(%q) = (%d, %q, %q), want (0, %q, %q)",
					test.args, code, stdout, stderr, test.want, "")
			}
		})
	}
}

func TestRunSplitExamples(t *testing.T) {
	tests := []struct {
		args []string
		want string
	}{
		{
			args: []string{"split", "100", "ana", "bob", "cara"},
			want: "ana: 33.34\nbob: 33.33\ncara: 33.33\ntotal: 100.00\n",
		},
		{
			args: []string{"split", "0.05", "ana", "bob*2", "cara"},
			want: "ana: 0.01\nbob: 0.03\ncara: 0.01\ntotal: 0.05\n",
		},
		{
			args: []string{"split", "10.00", "zed", "amy", "bo"},
			want: "zed: 3.33\namy: 3.34\nbo: 3.33\ntotal: 10.00\n",
		},
		{
			args: []string{"split", "1.00", "a*3", "b*2", "c"},
			want: "a: 0.50\nb: 0.33\nc: 0.17\ntotal: 1.00\n",
		},
		{
			args: []string{"split", "999999999.99", "a*99", "b*98", "c*97", "d"},
			want: "a: 335593220.33\nb: 332203389.83\nc: 328813559.32\nd: 3389830.51\ntotal: 999999999.99\n",
		},
	}
	for _, test := range tests {
		code, stdout, stderr := runCLI(test.args, "")
		if code != 0 || stdout != test.want || stderr != "" {
			t.Errorf("Run(%q) = (%d, %q, %q), want (0, %q, %q)",
				test.args, code, stdout, stderr, test.want, "")
		}
	}
}

func TestRunStandardInputAndSettledBalances(t *testing.T) {
	input := "ana paid 5 for bob\nbob sent ana 5\n"
	code, stdout, stderr := runCLI([]string{"balance", "-"}, input)
	if code != 0 || stdout != "ana: 0.00\nbob: 0.00\n" || stderr != "" {
		t.Errorf("balance from stdin = (%d, %q, %q), want (0, %q, %q)",
			code, stdout, stderr, "ana: 0.00\nbob: 0.00\n", "")
	}

	code, stdout, stderr = runCLI([]string{"settle", "-"}, input)
	if code != 0 || stdout != "all settled\n" || stderr != "" {
		t.Errorf("settle from stdin = (%d, %q, %q), want (0, %q, %q)",
			code, stdout, stderr, "all settled\n", "")
	}

	code, stdout, stderr = runCLI([]string{"settle", "-"}, "ana paid 5 for ana bob\n")
	if code != 0 || stdout != "bob pays ana: 2.50\n1 transfer\n" || stderr != "" {
		t.Errorf("settle from stdin = (%d, %q, %q), want (0, %q, %q)",
			code, stdout, stderr, "bob pays ana: 2.50\n1 transfer\n", "")
	}
}

func TestRunEmptyLedger(t *testing.T) {
	code, stdout, stderr := runCLI([]string{"balance", "-"}, "")
	if code != 0 || stdout != "" || stderr != "" {
		t.Errorf("empty balance = (%d, %q, %q), want (0, empty, empty)", code, stdout, stderr)
	}
	code, stdout, stderr = runCLI([]string{"settle", "-"}, "")
	if code != 0 || stdout != "all settled\n" || stderr != "" {
		t.Errorf("empty settle = (%d, %q, %q), want (0, %q, empty)",
			code, stdout, stderr, "all settled\n")
	}
}

func TestRunFileInputIsNotModified(t *testing.T) {
	path := filepath.Join(t.TempDir(), "expenses.txt")
	input := []byte("ana paid 5 for ana bob\n")
	if err := os.WriteFile(path, input, 0600); err != nil {
		t.Fatal(err)
	}

	for _, command := range []string{"balance", "settle"} {
		code, stdout, stderr := runCLI([]string{command, path}, "")
		if code != 0 || stderr != "" {
			t.Fatalf("%s file command = (%d, %q), want success", command, code, stderr)
		}
		if command == "balance" && stdout != "ana: +2.50\nbob: -2.50\n" {
			t.Errorf("balance stdout = %q", stdout)
		}
		if command == "settle" && stdout != "bob pays ana: 2.50\n1 transfer\n" {
			t.Errorf("settle stdout = %q", stdout)
		}
	}

	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(after, input) {
		t.Errorf("input file changed: got %q, want %q", after, input)
	}
}

func TestRunUsageAndSplitValidationOrder(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		message string
	}{
		{name: "no command", message: "missing command (balance, settle or split)"},
		{name: "unknown command", args: []string{"bogus"}, message: `unknown command "bogus"`},
		{name: "help is not special after another argument", args: []string{"--help", "balance"}, message: `unknown command "--help"`},
		{name: "balance help option", args: []string{"balance", "--help"}, message: `unknown option "--help"`},
		{name: "missing balance file", args: []string{"balance"}, message: "missing FILE"},
		{name: "balance option before positional", args: []string{"balance", "-x", "expenses.txt"}, message: `unknown option "-x"`},
		{name: "balance option after positional", args: []string{"balance", "expenses.txt", "-x"}, message: `unknown option "-x"`},
		{name: "first balance option wins", args: []string{"balance", "-x", "-y"}, message: `unknown option "-x"`},
		{name: "too many balance files", args: []string{"balance", "a", "b"}, message: "too many arguments"},
		{name: "too many settle files", args: []string{"settle", "a", "b"}, message: "too many arguments"},
		{name: "missing split amount", args: []string{"split"}, message: "missing AMOUNT"},
		{name: "split option checked first", args: []string{"split", "-x", "5", "ana"}, message: `unknown option "-x"`},
		{name: "split negative amount token is an option", args: []string{"split", "-5", "ana"}, message: `unknown option "-5"`},
		{name: "split lone dash token is an option", args: []string{"split", "-", "ana"}, message: `unknown option "-"`},
		{name: "split later option checked first", args: []string{"split", "x", "-x"}, message: `unknown option "-x"`},
		{name: "invalid split amount", args: []string{"split", "x", "ana"}, message: `invalid amount "x"`},
		{name: "missing split person", args: []string{"split", "5"}, message: "missing PERSON"},
		{name: "duplicate split person", args: []string{"split", "5", "ana", "ana"}, message: `duplicate person "ana"`},
		{name: "invalid split weight", args: []string{"split", "5", "ana*0"}, message: `invalid weight "ana*0"`},
		{name: "invalid split name", args: []string{"split", "5", "bo!b"}, message: `invalid name "bo!b"`},
		{name: "empty split name", args: []string{"split", "5", "*2"}, message: `invalid name ""`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assertCLIError(t, test.args, "", test.message)
		})
	}
}

func TestRunFileReadAndParseErrors(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing.txt")
	backslashPath := strings.ReplaceAll(missing, "/", "\\")
	displayPath := strings.ReplaceAll(backslashPath, "\\", "/")
	assertCLIError(t, []string{"balance", backslashPath}, "", displayPath+": cannot read the file")

	invalidUTF8 := filepath.Join(t.TempDir(), "invalid.txt")
	if err := os.WriteFile(invalidUTF8, []byte{'x', ' ', 0xff, '\n'}, 0600); err != nil {
		t.Fatal(err)
	}
	assertCLIError(t, []string{"balance", invalidUTF8}, "", filepath.ToSlash(invalidUTF8)+": input is not valid UTF-8")

	parseErrors := []struct {
		name    string
		line    string
		message string
	}{
		{name: "invalid first name", line: "!x paid 1 for a", message: `invalid name "!x"`},
		{name: "missing verb", line: "ana", message: "missing verb"},
		{name: "unknown verb", line: "ana owes 5 for bob", message: `unknown verb "owes"`},
		{name: "paid missing amount", line: "ana paid", message: "missing amount"},
		{name: "paid zero amount", line: "ana paid 0 for bob", message: `invalid amount "0"`},
		{name: "paid invalid amount", line: "ana paid 1.234 for bob", message: `invalid amount "1.234"`},
		{name: "paid missing for", line: "ana paid 5", message: `missing "for"`},
		{name: "paid expected for", line: "ana paid 5 bob", message: `expected "for", found "bob"`},
		{name: "paid no participants", line: "ana paid 5 for", message: `no people after "for"`},
		{name: "invalid participant name", line: "ana paid 5 for bo!b", message: `invalid name "bo!b"`},
		{name: "empty participant name", line: "ana paid 5 for *2", message: `invalid name ""`},
		{name: "zero participant weight", line: "ana paid 5 for bob*0", message: `invalid weight "bob*0"`},
		{name: "three digit participant weight", line: "ana paid 5 for bob*100", message: `invalid weight "bob*100"`},
		{name: "leading zero participant weight", line: "ana paid 5 for bob*007", message: `invalid weight "bob*007"`},
		{name: "empty participant weight", line: "ana paid 5 for bob*", message: `invalid weight "bob*"`},
		{name: "duplicate participant", line: "ana paid 5 for bob bob", message: `duplicate person "bob"`},
		{name: "sent missing recipient", line: "ana sent", message: "missing recipient"},
		{name: "sent invalid recipient", line: "ana sent !bob 1", message: `invalid name "!bob"`},
		{name: "sent to self", line: "ana sent ana 5", message: "cannot send to the same person"},
		{name: "sent missing amount", line: "ana sent bob", message: "missing amount"},
		{name: "sent invalid amount", line: "ana sent bob 1,5", message: `invalid amount "1,5"`},
		{name: "sent unexpected word", line: "ana sent bob 5 now", message: `unexpected "now"`},
	}
	for _, test := range parseErrors {
		t.Run(test.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "bad.txt")
			if err := os.WriteFile(path, []byte("\n"+test.line+"\n"), 0600); err != nil {
				t.Fatal(err)
			}
			assertCLIError(t, []string{"balance", path}, "", filepath.ToSlash(path)+":2: "+test.message)
		})
	}

	t.Run("record limit precedes parsing the 1001st record", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "many.txt")
		input := strings.Repeat("a paid 1 for b\n", 1000) + "! invalid record"
		if err := os.WriteFile(path, []byte(input), 0600); err != nil {
			t.Fatal(err)
		}
		assertCLIError(t, []string{"balance", path}, "", filepath.ToSlash(path)+":1001: more than 1000 records")
	})

	assertCLIError(t, []string{"settle", "-"}, "ana paid 1 for bob\nana sent\n",
		"<stdin>:2: missing recipient")
}

func TestRunDocumentationContainsSpecifiedContent(t *testing.T) {
	readme, err := os.ReadFile(filepath.Join("..", "..", "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(readme)
	required := []string{
		"usage: centsplit balance FILE\n       centsplit settle FILE\n       centsplit split AMOUNT PERSON...",
		`# trip to the lake
ana paid 90.00 for ana bob cara
bob paid 30.50 for ana bob
cara paid 12 for bob*2 cara
dave paid 10.01 for ana bob cara dave
bob sent ana 10.00`,
		"ana: +32.24\nbob: -15.25\ncara: -24.50\ndave: +7.51",
		"cara pays ana: 24.50\nbob pays ana: 7.74\nbob pays dave: 7.51\n3 transfers",
		"## Install",
		"## Usage",
		"## File format",
		"## How shares and settlements are decided",
		"## Exit codes",
		"## Development",
	}
	for _, text := range required {
		if !strings.Contains(content, text) {
			t.Errorf("README.md does not contain %q", text)
		}
	}
}
