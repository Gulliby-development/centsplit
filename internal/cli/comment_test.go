package cli

import "testing"

const commentedTripLedger = "# trip\n" +
	"ana paid 30.00 for ana bob cara # dinner\n" +
	"bob sent ana 5.00 #refund\n" +
	"cara paid 9 for bob*2 cara\t# taxi\n"

const uncommentedTripLedger = "ana paid 30.00 for ana bob cara\n" +
	"bob sent ana 5.00\n" +
	"cara paid 9 for bob*2 cara\n"

func TestRunInlineCommentsFromStandardInput(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "balance example",
			args: []string{"balance", "-"},
			want: "ana: +15.00\nbob: -11.00\ncara: -4.00\n",
		},
		{
			name: "settle example",
			args: []string{"settle", "-"},
			want: "bob pays ana: 11.00\ncara pays ana: 4.00\n2 transfers\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for _, input := range []string{commentedTripLedger, uncommentedTripLedger} {
				code, stdout, stderr := runCLI(test.args, input)
				if code != 0 || stdout != test.want || stderr != "" {
					t.Fatalf("Run(%q) = (%d, %q, %q), want (0, %q, empty)",
						test.args, code, stdout, stderr, test.want)
				}
			}
		})
	}
}

func TestRunInlineCommentErrors(t *testing.T) {
	assertCLIError(t, []string{"balance", "-"},
		"# first comment\n\nana paid 5 for # bob",
		"<stdin>:3: no people after \"for\"")
	assertCLIError(t, []string{"split", "5", "ana", "#bob"},
		"",
		"invalid name \"#bob\"")
}
