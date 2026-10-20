package ledger

import (
	"fmt"
	"sort"

	"github.com/Gulliby-development/centsplit/internal/money"
)

// Shares gives people with an exact share that amount and divides the rest
// among the others by weighted largest remainder. The result stays in
// participant order.
func Shares(cents int64, people []Person) []int64 {
	var exactSum int64
	others := make([]Person, 0, len(people))
	indices := make([]int, 0, len(people))
	for index, person := range people {
		if person.Exact > 0 {
			exactSum += person.Exact
			continue
		}
		others = append(others, person)
		indices = append(indices, index)
	}
	if len(others) == len(people) {
		return weightedShares(cents, people)
	}
	shares := make([]int64, len(people))
	for index, person := range people {
		if person.Exact > 0 {
			shares[index] = person.Exact
		}
	}
	if len(others) == 0 {
		return shares
	}
	rest := cents - exactSum
	if rest < 0 {
		rest = 0
	}
	for position, share := range weightedShares(rest, others) {
		shares[indices[position]] = share
	}
	return shares
}

// CheckExact validates the exact shares against the amount in cents.
func CheckExact(cents int64, people []Person) error {
	var sum int64
	count := 0
	for _, person := range people {
		if person.Exact > 0 {
			sum += person.Exact
			count++
		}
	}
	if count == 0 {
		return nil
	}
	if sum > cents {
		return fmt.Errorf("exact shares add up to %s, more than %s", money.Format(sum, false), money.Format(cents, false))
	}
	if count == len(people) && sum != cents {
		return fmt.Errorf("exact shares add up to %s, not %s", money.Format(sum, false), money.Format(cents, false))
	}
	return nil
}

// weightedShares divides cents by weighted largest remainder; ties for a
// remaining cent are resolved by name.
func weightedShares(cents int64, people []Person) []int64 {
	shares := make([]int64, len(people))
	if len(people) == 0 {
		return shares
	}

	var totalWeight int64
	for _, person := range people {
		totalWeight += person.Weight
	}
	if totalWeight <= 0 {
		return shares
	}

	remainders := make([]int64, len(people))
	var allocated int64
	for index, person := range people {
		weightedCents := cents * person.Weight
		shares[index] = weightedCents / totalWeight
		remainders[index] = weightedCents % totalWeight
		allocated += shares[index]
	}

	remaining := cents - allocated
	order := make([]int, len(people))
	for index := range order {
		order[index] = index
	}
	sort.Slice(order, func(i, j int) bool {
		left, right := order[i], order[j]
		if remainders[left] != remainders[right] {
			return remainders[left] > remainders[right]
		}
		if people[left].Name != people[right].Name {
			return people[left].Name < people[right].Name
		}
		return left < right
	})

	for index := int64(0); index < remaining && index < int64(len(order)); index++ {
		shares[order[index]]++
	}
	return shares
}
