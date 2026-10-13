package money

import "strconv"

// ParseAmount parses a nonzero decimal amount with at most nine whole digits
// and one or two fractional digits, returning its value in cents.
func ParseAmount(text string) (cents int64, ok bool) {
	if len(text) == 0 {
		return 0, false
	}

	dot := -1
	for i := 0; i < len(text); i++ {
		if text[i] == '.' {
			if dot >= 0 {
				return 0, false
			}
			dot = i
			continue
		}
		if text[i] < '0' || text[i] > '9' {
			return 0, false
		}
	}

	wholeEnd := len(text)
	fractionStart := len(text)
	if dot >= 0 {
		wholeEnd = dot
		fractionStart = dot + 1
		fractionLength := len(text) - fractionStart
		if fractionLength < 1 || fractionLength > 2 {
			return 0, false
		}
	}
	if wholeEnd < 1 || wholeEnd > 9 {
		return 0, false
	}

	var whole int64
	for i := 0; i < wholeEnd; i++ {
		whole = whole*10 + int64(text[i]-'0')
	}
	cents = whole * 100

	if dot >= 0 {
		fraction := int64(text[fractionStart] - '0')
		if fractionStart+1 < len(text) {
			fraction = fraction*10 + int64(text[fractionStart+1]-'0')
		} else {
			fraction *= 10
		}
		cents += fraction
	}

	return cents, cents != 0
}

// Format renders cents as a decimal amount with exactly two fractional digits.
// When sign is true, positive nonzero values receive a leading plus sign.
func Format(cents int64, sign bool) string {
	negative := cents < 0
	var magnitude uint64
	if negative {
		// Avoid overflowing when cents is the minimum int64 value.
		magnitude = uint64(-(cents + 1)) + 1
	} else {
		magnitude = uint64(cents)
	}

	formatted := strconv.FormatUint(magnitude/100, 10) + "."
	remainder := magnitude % 100
	if remainder < 10 {
		formatted += "0"
	}
	formatted += strconv.FormatUint(remainder, 10)

	if negative {
		return "-" + formatted
	}
	if sign && cents > 0 {
		return "+" + formatted
	}
	return formatted
}
