package ledger

import "sort"

// Shares divides cents among people by weighted largest remainder. The result
// stays in participant order; ties for a remaining cent are resolved by name.
func Shares(cents int64, people []Person) []int64 {
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
