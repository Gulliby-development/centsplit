package ledger

import (
	"fmt"
	"strings"

	"github.com/Gulliby-development/centsplit/internal/money"
)

// Person is a participant in a shared expense. Weight defaults to 1.
type Person struct {
	Name   string
	Weight int64
	Exact  int64
}

// Record is either a paid expense or a direct payment between people.
type Record struct {
	Kind   string
	From   string
	To     string
	Cents  int64
	People []Person
}

// ParsePeople parses NAME and NAME*WEIGHT tokens, rejecting duplicate names.
func ParsePeople(tokens []string) ([]Person, error) {
	people := make([]Person, 0, len(tokens))
	seen := make(map[string]struct{}, len(tokens))
	for _, token := range tokens {
		person, err := parsePerson(token)
		if err != nil {
			return nil, err
		}
		if _, exists := seen[person.Name]; exists {
			return nil, fmt.Errorf("duplicate person %q", person.Name)
		}
		seen[person.Name] = struct{}{}
		people = append(people, person)
	}
	return people, nil
}

// ParseFile parses expense records, preserving source line numbers in errors.
func ParseFile(display, text string) ([]Record, error) {
	text = strings.TrimPrefix(text, "\uFEFF")
	lines := strings.Split(text, "\n")
	records := make([]Record, 0)
	recordCount := 0

	for index, line := range lines {
		hasLineFeed := index < len(lines)-1
		if hasLineFeed && strings.HasSuffix(line, "\r") {
			line = strings.TrimSuffix(line, "\r")
		}
		line = strings.Trim(line, " \t")
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		lineNumber := index + 1
		if recordCount == 1000 {
			return nil, fileError(display, lineNumber, "more than 1000 records")
		}
		recordCount++

		record, err := parseRecordLine(line)
		if err != nil {
			return nil, fileError(display, lineNumber, err.Error())
		}
		records = append(records, record)
	}
	return records, nil
}

func parseRecordLine(line string) (Record, error) {
	words := splitWords(line)
	for index, word := range words {
		if strings.HasPrefix(word, "#") {
			words = words[:index]
			break
		}
	}
	record := Record{}

	if !validName(words[0]) {
		return Record{}, invalidName(words[0])
	}
	record.From = words[0]
	if len(words) < 2 {
		return Record{}, fmt.Errorf("missing verb")
	}

	switch words[1] {
	case "paid":
		record.Kind = "paid"
		if len(words) < 3 {
			return Record{}, fmt.Errorf("missing amount")
		}
		cents, ok := money.ParseAmount(words[2])
		if !ok {
			return Record{}, invalidAmount(words[2])
		}
		record.Cents = cents
		if len(words) < 4 {
			return Record{}, fmt.Errorf("missing \"for\"")
		}
		if words[3] != "for" {
			return Record{}, fmt.Errorf("expected \"for\", found %q", words[3])
		}
		if len(words) < 5 {
			return Record{}, fmt.Errorf("no people after \"for\"")
		}
		people, err := ParsePeople(words[4:])
		if err != nil {
			return Record{}, err
		}
		if err := CheckExact(record.Cents, people); err != nil {
			return Record{}, err
		}
		record.People = people
		return record, nil
	case "sent":
		record.Kind = "sent"
		if len(words) < 3 {
			return Record{}, fmt.Errorf("missing recipient")
		}
		if !validName(words[2]) {
			return Record{}, invalidName(words[2])
		}
		record.To = words[2]
		if record.From == record.To {
			return Record{}, fmt.Errorf("cannot send to the same person")
		}
		if len(words) < 4 {
			return Record{}, fmt.Errorf("missing amount")
		}
		cents, ok := money.ParseAmount(words[3])
		if !ok {
			return Record{}, invalidAmount(words[3])
		}
		record.Cents = cents
		if len(words) > 4 {
			return Record{}, fmt.Errorf("unexpected %q", words[4])
		}
		return record, nil
	default:
		return Record{}, fmt.Errorf("unknown verb %q", words[1])
	}
}

func parsePerson(token string) (Person, error) {
	if equals := strings.IndexByte(token, '='); equals >= 0 {
		name := token[:equals]
		if !validName(name) {
			return Person{}, invalidName(name)
		}
		cents, ok := money.ParseAmount(token[equals+1:])
		if !ok {
			return Person{}, invalidAmount(token)
		}
		return Person{Name: name, Weight: 1, Exact: cents}, nil
	}
	name := token
	weight := int64(1)
	if separator := strings.IndexByte(token, '*'); separator >= 0 {
		name = token[:separator]
		if !validName(name) {
			return Person{}, invalidName(name)
		}
		weightText := token[separator+1:]
		if !validWeight(weightText) {
			return Person{}, fmt.Errorf("invalid weight %q", token)
		}
		weight = 0
		for i := 0; i < len(weightText); i++ {
			weight = weight*10 + int64(weightText[i]-'0')
		}
	} else if !validName(name) {
		return Person{}, invalidName(name)
	}
	return Person{Name: name, Weight: weight}, nil
}

func validWeight(text string) bool {
	if len(text) < 1 || len(text) > 2 || text[0] == '0' {
		return false
	}
	value := int64(0)
	for i := 0; i < len(text); i++ {
		if text[i] < '0' || text[i] > '9' {
			return false
		}
		value = value*10 + int64(text[i]-'0')
	}
	return value >= 1 && value <= 99
}

func validName(name string) bool {
	if len(name) < 1 || len(name) > 20 {
		return false
	}
	for i := 0; i < len(name); i++ {
		character := name[i]
		if (character >= 'A' && character <= 'Z') ||
			(character >= 'a' && character <= 'z') ||
			(character >= '0' && character <= '9') ||
			character == '_' || character == '-' {
			continue
		}
		return false
	}
	return true
}

func invalidName(name string) error {
	return fmt.Errorf("invalid name %q", name)
}

func invalidAmount(text string) error {
	return fmt.Errorf("invalid amount %q", text)
}

func fileError(display string, line int, message string) error {
	return fmt.Errorf("%s:%d: %s", display, line, message)
}

func splitWords(line string) []string {
	words := make([]string, 0)
	for start := 0; start < len(line); {
		for start < len(line) && (line[start] == ' ' || line[start] == '\t') {
			start++
		}
		if start == len(line) {
			break
		}
		end := start
		for end < len(line) && line[end] != ' ' && line[end] != '\t' {
			end++
		}
		words = append(words, line[start:end])
		start = end
	}
	return words
}
