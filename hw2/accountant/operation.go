package accountant

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"
)

const (
	// describes operation's statuses
	valid OperationStatus = iota
	invalid
	skip

	income      = "income"
	outcome     = "outcome"
	emptyString = ""

	// if you want to see operation processing
	printErrors = true
)

var (
	// available Operation types (can be expanded)
	typeValues = map[string][]string{
		income:  {"income", "Income", "+"},
		outcome: {"outcome", "Outcome", "-"},
	}

	// ErrSkipOperation describes violation behavior for operations to be skipped
	ErrSkipOperation = errors.New("this operation should be skipped")
	// ErrRejectOperation describes violation behavior for operations to be rejected
	ErrRejectOperation = errors.New("this operation should be rejected")
)

// OperationStatus describes will operation be skipped, rejected or passed
type OperationStatus int

// EmbeddedJSON describes unmarshalled JSON into struct with empty interfaces
// as value of the fields
type EmbeddedJSON struct {
	Company   string    `json:"company"`
	Operation InnerJSON `json:"operation"`
	InnerJSON
}

// InnerJSON describes inner JSON field
type InnerJSON struct {
	Type         string      `json:"type"`
	Value        interface{} `json:"value"`
	ID           interface{} `json:"id"`
	CreationTime time.Time   `json:"created_at"`
}

// Operation describes structure of entire JSON, and it's status of condition
// (valid, invalid, rejected)
type Operation struct {
	Company string
	Type    string
	Value   int64
	ID      interface{}
	Time    string
	Status  OperationStatus
}

// String provides string representation of Operation
func (operation Operation) String() string {
	return fmt.Sprintf("company:\t%s\n"+
		"type:\t%s\n"+
		"value:\t%d\n"+
		"id:\t%s\n"+
		"created_at:\t%s\n",
		operation.Company, operation.Type, operation.Value, operation.ID, operation.Time)
}

// UnmarshalJSON implements Unmarshaler interface and represents custom Unmarshal.
func (operation *Operation) UnmarshalJSON(data []byte) error {
	// unmarshall json to map
	var fields EmbeddedJSON
	if err := json.Unmarshal(data, &fields); err != nil {
		operation.Status = skip
	}

	// operations with incorrect company, id or time will be skipped
	operation.updateOperation(fields)
	return nil
}

// updateOperation enter fields (field by field) from dirtyJSON to Operation instance.
// And updates Operation instance's status.
func (operation *Operation) updateOperation(fields EmbeddedJSON) {
	var err error
	if operation.Company, err = fields.CompanyProcessed(); err != nil {
		switchErrors(operation, err)
		return
	}
	if operation.Time, err = fields.TimeProcessed(); err != nil {
		switchErrors(operation, err)
		return
	}
	if operation.ID, err = fields.IDProcessed(); err != nil {
		switchErrors(operation, err)
		return
	}
	if operation.Type, err = fields.TypeProcessed(); err != nil {
		switchErrors(operation, err)
		return
	}
	if operation.Value, err = fields.ValueProcessed(); err != nil {
		switchErrors(operation, err)
		return
	}
}

// switchErrors update Operation instance's status by given error.
func switchErrors(operation *Operation, err error) {
	if printErrors {
		fmt.Printf("%s\n\n", err)
	}
	switch {
	case errors.Is(err, ErrSkipOperation):
		operation.Status = skip
	case errors.Is(err, ErrRejectOperation):
		operation.Status = invalid
	}
}

// CompanyProcessed returns string implementation of company from EmbeddedJSON
func (dj EmbeddedJSON) CompanyProcessed() (string, error) {
	if dj.Company == emptyString {
		return emptyString, fmt.Errorf("%scompany tag content is empty: %w",
			dj, ErrSkipOperation)
	}
	return dj.Company, nil
}

// TypeProcessed returns string implementation of type from EmbeddedJSON
func (dj EmbeddedJSON) TypeProcessed() (string, error) {
	var typeTag string
	// check if json is broken
	if dj.Type == emptyString && dj.Operation.Type == emptyString {
		return emptyString, fmt.Errorf("%sjson hasn't type tag: %w",
			dj, ErrSkipOperation)
	}
	// pick not null value
	if dj.Type != emptyString {
		typeTag = dj.Type
	} else {
		typeTag = dj.Operation.Type
	}

	// check for outcome value
	for _, typeValue := range typeValues[outcome] {
		if typeValue == typeTag {
			return outcome, nil
		}
	}
	// check for income value
	for _, typeValue := range typeValues[income] {
		if typeValue == typeTag {
			return income, nil
		}
	}
	return emptyString, fmt.Errorf("%stype tag content is illigal: %w",
		dj, ErrRejectOperation)
}

// ValueProcessed returns int64 implementation of value from EmbeddedJSON
func (dj EmbeddedJSON) ValueProcessed() (int64, error) {
	var valueTag interface{}
	// check if json is broken
	if dj.Value == nil && dj.Operation.Value == nil {
		return 0, fmt.Errorf("%sjson hasn't value tag: %w",
			dj, ErrRejectOperation)
	}
	// pick not null value
	if dj.Value != nil {
		valueTag = dj.Value
	} else {
		valueTag = dj.Operation.Value
	}

	// try to assert
	switch value := valueTag.(type) {
	case float64:
		// separation of the int part
		return separateFloat(dj, value, "value tag content has fractional: ")
	case string:
		// parsing
		floatValue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return 0, fmt.Errorf("%svalue tag content isn't a float: %w",
				dj, ErrRejectOperation)
		}
		// separation of the int part
		return separateFloat(dj, floatValue, "value tag content has fractional: ")
	default:
		return 0, fmt.Errorf("%svalue tag content isn't a digit: %w",
			dj, ErrRejectOperation)
	}
}

// IDProcessed returns string implementation of id from EmbeddedJSON
func (dj EmbeddedJSON) IDProcessed() (interface{}, error) {
	var idTag interface{}
	// check if json is broken
	if dj.ID == nil && dj.Operation.ID == nil {
		return emptyString, fmt.Errorf("%sjson hasn't id tag: %w",
			dj, ErrSkipOperation)
	}
	// pick not null value
	if dj.ID != nil {
		idTag = dj.ID
	} else {
		idTag = dj.Operation.ID
	}

	// try to assert
	switch id := idTag.(type) {
	case float64:
		// separation of int part
		intPart, frac := math.Modf(id)
		if frac == 0 {
			return int64(intPart), nil
		}
		return "", fmt.Errorf("%sid tag content has fractional: %w",
			dj, ErrRejectOperation)
	case string:
		if id == emptyString {
			return emptyString, fmt.Errorf("%stime tag is empty: %w",
				dj, ErrSkipOperation)
		}
		return id, nil
	default:
		return emptyString, fmt.Errorf("%scan't cast id tag to string %w",
			dj, ErrSkipOperation)
	}
}

// TimeProcessed returns time.Time implementation of created_at from EmbeddedJSON
func (dj EmbeddedJSON) TimeProcessed() (string, error) {
	// check if both time values are empty
	if dj.Operation.CreationTime.IsZero() &&
		dj.CreationTime.IsZero() {
		return time.Now().Format(time.RFC3339),
			fmt.Errorf("%sjson hasn't time tag: %w", dj, ErrSkipOperation)
	}
	// check if inner json time is empty
	if dj.CreationTime.IsZero() {
		return dj.Operation.CreationTime.Format(time.RFC3339), nil
	}
	return dj.CreationTime.Format(time.RFC3339), nil
}

// String represents EmbeddedJSON as string through string builder.
func (dj EmbeddedJSON) String() string {
	return fmt.Sprintf("CompanyProcessed: %s\n"+
		"TypeProcessed: %s\n"+
		"ValueProcessed: %v\n"+
		"IDProcessed: %v\n"+
		"TimeProcessed: %s\n%s",
		dj.Company, dj.Type, dj.Value, dj.ID,
		dj.CreationTime, dj.Operation)
}

// String represents EmbeddedJSON as string through string builder.
func (inner InnerJSON) String() string {
	return fmt.Sprintf("Operation:\n"+
		"\tTypeProcessed: %v\n"+
		"\tValueProcessed: %v\n"+
		"\tIDProcessed: %v\n"+
		"\tTimeProcessed: %s\n",
		inner.Type, inner.Value, inner.ID, inner.CreationTime)
}

// separateFloat separate float to int part and check fractional for zero-value
func separateFloat(fields EmbeddedJSON, fl float64, errorText string) (int64, error) {
	intPart, frac := math.Modf(fl)
	if frac == 0 {
		return int64(intPart), nil
	}
	return 0, fmt.Errorf("%s%s: %w",
		fields, errorText, ErrRejectOperation)
}

type Operations []Operation

// Len is the number of elements in the collection.
func (operations Operations) Len() int { return len(operations) }

// Less reports whether the element with
// index i should sort before the element with index j.
func (operations Operations) Less(i, j int) bool {
	if operations[i].Status == skip {
		return false
	}
	return operations[i].Time < operations[j].Time
}

// Swap swaps the elements with indexes i and j.
func (operations Operations) Swap(i, j int) {
	operations[i], operations[j] = operations[j], operations[i]
}
