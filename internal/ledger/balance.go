package ledger

import "sort"

// Transfer records a payment from a debtor to a creditor.
type Transfer struct {
	From  string
	To    string
	Cents int64
}

// Balances computes each person's net position in cents. Positive balances are
// owed money and negative balances owe money.
func Balances(records []Record) map[string]int64 {
	balances := make(map[string]int64)
	for _, record := range records {
		switch record.Kind {
		case "paid":
			balances[record.From] += record.Cents
			shares := Shares(record.Cents, record.People)
			for index, person := range record.People {
				balances[person.Name] -= shares[index]
			}
		case "sent":
			balances[record.From] += record.Cents
			balances[record.To] -= record.Cents
		}
	}
	return balances
}

// SortedNames returns the map's names in ascending byte order.
func SortedNames(balances map[string]int64) []string {
	names := make([]string, 0, len(balances))
	for name := range balances {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Settle returns deterministic transfers that bring the supplied balances to
// zero. It does not modify the balances map.
func Settle(balances map[string]int64) []Transfer {
	type party struct {
		name   string
		amount int64
	}
	debtors := make([]party, 0)
	creditors := make([]party, 0)
	for name, balance := range balances {
		switch {
		case balance < 0:
			debtors = append(debtors, party{name: name, amount: -balance})
		case balance > 0:
			creditors = append(creditors, party{name: name, amount: balance})
		}
	}

	transfers := make([]Transfer, 0)
	for len(debtors) > 0 && len(creditors) > 0 {
		sort.Slice(debtors, func(i, j int) bool {
			if debtors[i].amount != debtors[j].amount {
				return debtors[i].amount > debtors[j].amount
			}
			return debtors[i].name < debtors[j].name
		})
		sort.Slice(creditors, func(i, j int) bool {
			if creditors[i].amount != creditors[j].amount {
				return creditors[i].amount > creditors[j].amount
			}
			return creditors[i].name < creditors[j].name
		})

		debtor, creditor := &debtors[0], &creditors[0]
		amount := debtor.amount
		if creditor.amount < amount {
			amount = creditor.amount
		}
		transfers = append(transfers, Transfer{
			From:  debtor.name,
			To:    creditor.name,
			Cents: amount,
		})
		debtor.amount -= amount
		creditor.amount -= amount
		if debtor.amount == 0 {
			debtors = debtors[1:]
		}
		if creditor.amount == 0 {
			creditors = creditors[1:]
		}
	}
	return transfers
}
