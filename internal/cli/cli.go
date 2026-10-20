package cli

import (
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/Gulliby-development/centsplit/internal/ledger"
	"github.com/Gulliby-development/centsplit/internal/money"
)

const usage = "usage: centsplit balance FILE\n       centsplit settle FILE\n       centsplit split AMOUNT PERSON...\n"

// Run executes a centsplit command using the supplied streams.
func Run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		return reportError(stderr, "missing command (balance, settle or split)")
	}
	if len(args) == 1 && args[0] == "--help" {
		_, _ = io.WriteString(stdout, usage)
		return 0
	}

	switch args[0] {
	case "balance", "settle":
		return runLedgerCommand(args[0], args[1:], stdin, stdout, stderr)
	case "split":
		return runSplit(args[1:], stdout, stderr)
	default:
		return reportError(stderr, fmt.Sprintf("unknown command %q", args[0]))
	}
}

func runLedgerCommand(command string, args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	file, message := ledgerFileArgument(args)
	if message != "" {
		return reportError(stderr, message)
	}

	display, text, message := readLedgerFile(file, stdin)
	if message != "" {
		return reportError(stderr, message)
	}
	records, err := ledger.ParseFile(display, text)
	if err != nil {
		return reportError(stderr, err.Error())
	}

	balances := ledger.Balances(records)
	var output strings.Builder
	if command == "balance" {
		for _, name := range ledger.SortedNames(balances) {
			output.WriteString(name)
			output.WriteString(": ")
			output.WriteString(money.Format(balances[name], true))
			output.WriteByte('\n')
		}
	} else {
		transfers := ledger.Settle(balances)
		if len(transfers) == 0 {
			output.WriteString("all settled\n")
		} else {
			for _, transfer := range transfers {
				output.WriteString(transfer.From)
				output.WriteString(" pays ")
				output.WriteString(transfer.To)
				output.WriteString(": ")
				output.WriteString(money.Format(transfer.Cents, false))
				output.WriteByte('\n')
			}
			if len(transfers) == 1 {
				output.WriteString("1 transfer\n")
			} else {
				output.WriteString(fmt.Sprintf("%d transfers\n", len(transfers)))
			}
		}
	}
	_, _ = io.WriteString(stdout, output.String())
	return 0
}

func ledgerFileArgument(args []string) (string, string) {
	files := make([]string, 0, len(args))
	for _, arg := range args {
		if strings.HasPrefix(arg, "-") && arg != "-" {
			return "", fmt.Sprintf("unknown option %q", arg)
		}
		files = append(files, arg)
	}
	if len(files) == 0 {
		return "", "missing FILE"
	}
	if len(files) > 1 {
		return "", "too many arguments"
	}
	return files[0], ""
}

func readLedgerFile(file string, stdin io.Reader) (string, string, string) {
	display := displayPath(file)
	var (
		data []byte
		err  error
	)
	if file == "-" {
		display = "<stdin>"
		data, err = io.ReadAll(stdin)
	} else {
		data, err = os.ReadFile(file)
	}
	if err != nil {
		return display, "", display + ": cannot read the file"
	}
	if !utf8.Valid(data) {
		return display, "", display + ": input is not valid UTF-8"
	}
	return display, string(data), ""
}

func displayPath(file string) string {
	return strings.ReplaceAll(file, "\\", "/")
}

func runSplit(args []string, stdout, stderr io.Writer) int {
	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
			return reportError(stderr, fmt.Sprintf("unknown option %q", arg))
		}
	}
	if len(args) == 0 {
		return reportError(stderr, "missing AMOUNT")
	}
	amount, ok := money.ParseAmount(args[0])
	if !ok {
		return reportError(stderr, fmt.Sprintf("invalid amount %q", args[0]))
	}
	if len(args) == 1 {
		return reportError(stderr, "missing PERSON")
	}
	people, err := ledger.ParsePeople(args[1:])
	if err != nil {
		return reportError(stderr, err.Error())
	}
	if err := ledger.CheckExact(amount, people); err != nil {
		return reportError(stderr, err.Error())
	}

	shares := ledger.Shares(amount, people)
	var output strings.Builder
	for index, person := range people {
		output.WriteString(person.Name)
		output.WriteString(": ")
		output.WriteString(money.Format(shares[index], false))
		output.WriteByte('\n')
	}
	output.WriteString("total: ")
	output.WriteString(money.Format(amount, false))
	output.WriteByte('\n')
	_, _ = io.WriteString(stdout, output.String())
	return 0
}

func reportError(stderr io.Writer, message string) int {
	_, _ = io.WriteString(stderr, "centsplit: "+message+"\n")
	return 2
}
