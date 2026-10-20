<p align="center"><img src="docs/logo.svg" width="96" alt="" /></p>

# centsplit

centsplit is an offline command-line tool for people sharing a flat, trip, or project budget. Record who paid and who shared each expense, then calculate exact cent balances or a deterministic list of repayments. The program uses no floating-point arithmetic, does not need network access, and never changes the expense file.

[![CI](https://github.com/Gulliby-development/centsplit/actions/workflows/ci.yml/badge.svg)](https://github.com/Gulliby-development/centsplit/actions/workflows/ci.yml) [![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

## Install

From the repository root, build the executable:

```sh
go build -o centsplit .
```

To run it directly with Go, replace the `FILE` placeholder with the path to an existing expense file:

```sh
go run . balance FILE
```

## Usage

Running `go run . --help` prints the command syntax:

```sh
go run . --help
```

```text
usage: centsplit balance FILE
       centsplit settle FILE
       centsplit split AMOUNT PERSON...
```

`balance FILE` lists every person's balance. Positive values are owed to that person; negative values are amounts they owe. `settle FILE` produces transfers that reconcile the balances. `split AMOUNT PERSON...` divides one amount among the listed people without reading a file. Use `-` as FILE to read the expense records from standard input.

Here is a shared trip ledger named `trip.txt`:

```text
# trip to the lake
ana paid 90.00 for ana bob cara
bob paid 30.50 for ana bob
cara paid 12 for bob*2 cara
dave paid 10.01 for ana bob cara dave
bob sent ana 10.00
```

Run `centsplit balance trip.txt` for the balances:

```text
ana: +32.24
bob: -15.25
cara: -24.50
dave: +7.51
```

Run `centsplit settle trip.txt` for the repayment plan:

```text
cara pays ana: 24.50
bob pays ana: 7.74
bob pays dave: 7.51
3 transfers
```

## File format

Each nonblank, non-comment line is one record. A paid-expense record has the form `NAME paid AMOUNT for PERSON...`; a repayment or loan record has the form `NAME sent NAME AMOUNT`. Names are case-sensitive and contain 1–20 ASCII letters, digits, underscores, or hyphens. A participant may have a weight from 1 to 99 written as `NAME*WEIGHT`; an omitted weight is 1.

A participant may instead have an exact share written as `NAME=AMOUNT`, with the amount in the same format as any other amount. Exact shares are taken first and the rest of the amount is divided among the participants without one by weight. The exact shares may not add up to more than the amount, and when every participant has one they must add up to exactly the amount. `NAME*WEIGHT=AMOUNT` is rejected as `invalid name`. The `split` command accepts the same tokens: `centsplit split 30.00 ana=12.00 bob cara` prints:

```text
ana: 12.00
bob: 9.00
cara: 9.00
total: 30.00
```

Amounts use 1–9 whole-number digits and may have one or two decimal digits after a period. They must be greater than zero; signs, commas, and other decimal forms are not accepted. Lines may use spaces or tabs between fields. Blank lines and lines whose first non-whitespace character is `#` are ignored. A `#` that starts a word (after spaces or tabs) also begins a comment that runs to the end of the line, as in `ana paid 30.00 for ana bob # dinner`; a `#` inside a word belongs to the word and is rejected, as in `invalid name "bob#1"`. Comments are part of the file format only: the `split` command takes none. A UTF-8 byte-order mark at the start of a file is accepted, as are LF and CRLF line endings. A file may contain at most 1000 records.

## How shares and settlements are decided

All values are integer cents. For a weighted split, cents are first allocated proportionally using integer division. Any remaining cents go to the participants with the largest fractional remainders; equal remainders are resolved by ascending byte-order name. The resulting shares always add up to the original amount, and the displayed split shares stay in the order the people were supplied.

A payment raises the payer's balance and lowers each participant's balance by their share. A `sent` record raises the sender's balance and lowers the recipient's balance. Balances therefore sum to zero. Settlement repeatedly pairs the largest debtor with the largest creditor, breaking equal amounts by ascending name, and transfers the smaller outstanding amount. Each choice is deterministic; the procedure does not promise the fewest possible transfers.

## Exit codes

Successful commands exit with code 0. Help is successful when `--help` is the only argument. Usage or input errors exit with code 2, print one `centsplit: MESSAGE` diagnostic to standard error, and leave standard output empty. File diagnostics identify the displayed file path and line number. A file path in a diagnostic uses forward slashes; standard input is displayed as `<stdin>`. The program does not include machine-dependent operating-system text in errors.

## Development

Run the checks from the repository root:

```sh
go vet ./...
gofmt -l .
go test ./...
```

## License

MIT, see [LICENSE](LICENSE).

---

Built with BAKED.
