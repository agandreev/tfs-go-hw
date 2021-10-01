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

// DirtyJSON describes unmarshalled JSON into struct with empty interfaces
// as value of the fields
type DirtyJSON struct {
	InnerOperation  InnerOperations `json:"operation"`
	CompanyTag      string          `json:"company"`
	TypeTag         string          `json:"type"`
	ValueTag        interface{}     `json:"value"`
	IDTag           interface{}     `json:"id"`
	CreationTimeTag time.Time       `json:"created_at"`
}

// InnerOperations describes inner JSON field
type InnerOperations struct {
	TypeTag         string      `json:"type"`
	ValueTag        interface{} `json:"value"`
	IDTag           interface{} `json:"id"`
	CreationTimeTag time.Time   `json:"created_at"`
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
	var fields DirtyJSON
	if err := json.Unmarshal(data, &fields); err != nil {
		operation.Status = skip
	}

	// operations with incorrect company, id or time will be skipped
	operation.updateOperation(fields)
	return nil
}

// updateOperation enter fields (field by field) from dirtyJSON to Operation instance.
// And updates Operation instance's status.
func (operation *Operation) updateOperation(fields DirtyJSON) {
	var err error
	if operation.Company, err = fields.Company(); err != nil {
		switchErrors(operation, err)
		return
	}
	if operation.Time, err = fields.Time(); err != nil {
		switchErrors(operation, err)
		return
	}
	if operation.ID, err = fields.ID(); err != nil {
		switchErrors(operation, err)
		return
	}
	if operation.Type, err = fields.Type(); err != nil {
		switchErrors(operation, err)
		return
	}
	if operation.Value, err = fields.Value(); err != nil {
		switchErrors(operation, err)
		return
	}
	operation.harmonizeValueWithType()
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

// Company returns string implementation of company from DirtyJSON
func (dj DirtyJSON) Company() (string, error) {
	if dj.CompanyTag == emptyString {
		return emptyString, fmt.Errorf("%scompany tag content is empty: %w",
			dj, ErrSkipOperation)
	}
	return dj.CompanyTag, nil
}

// Type returns string implementation of type from DirtyJSON
func (dj DirtyJSON) Type() (string, error) {
	var typeTag string
	// check if json is broken
	if dj.TypeTag == emptyString && dj.InnerOperation.TypeTag == emptyString {
		return emptyString, fmt.Errorf("%sjson hasn't type tag: %w",
			dj, ErrSkipOperation)
	}
	// pick not null value
	if dj.TypeTag != emptyString {
		typeTag = dj.TypeTag
	} else {
		typeTag = dj.InnerOperation.TypeTag
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

// Value returns int64 implementation of value from DirtyJSON
func (dj DirtyJSON) Value() (int64, error) {
	var valueTag interface{}
	// check if json is broken
	if dj.ValueTag == nil && dj.InnerOperation.ValueTag == nil {
		return 0, fmt.Errorf("%sjson hasn't value tag: %w",
			dj, ErrRejectOperation)
	}
	// pick not null value
	if dj.ValueTag != nil {
		valueTag = dj.ValueTag
	} else {
		valueTag = dj.InnerOperation.ValueTag
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

// ID returns string implementation of id from DirtyJSON
func (dj DirtyJSON) ID() (interface{}, error) {
	var idTag interface{}
	// check if json is broken
	if dj.IDTag == nil && dj.InnerOperation.IDTag == nil {
		return emptyString, fmt.Errorf("%sjson hasn't id tag: %w",
			dj, ErrSkipOperation)
	}
	// pick not null value
	if dj.IDTag != nil {
		idTag = dj.IDTag
	} else {
		idTag = dj.InnerOperation.IDTag
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

// Time returns time.Time implementation of created_at from DirtyJSON
func (dj DirtyJSON) Time() (string, error) {
	// check if both time values are empty
	if dj.InnerOperation.CreationTimeTag.IsZero() &&
		dj.CreationTimeTag.IsZero() {
		return time.Now().Format(time.RFC3339),
			fmt.Errorf("%sjson hasn't time tag: %w", dj, ErrSkipOperation)
	}
	// check if inner json time is empty
	if dj.CreationTimeTag.IsZero() {
		return dj.InnerOperation.CreationTimeTag.Format(time.RFC3339), nil
	}
	return dj.CreationTimeTag.Format(time.RFC3339), nil
}

// String represents DirtyJSON as string through string builder.
func (dj DirtyJSON) String() string {
	return fmt.Sprintf("Company: %s\n"+
		"Type: %s\n"+
		"Value: %v\n"+
		"ID: %v\n"+
		"Time: %s\n%s",
		dj.CompanyTag, dj.TypeTag, dj.ValueTag, dj.IDTag,
		dj.CreationTimeTag, dj.InnerOperation)
}

// String represents DirtyJSON as string through string builder.
func (inner InnerOperations) String() string {
	return fmt.Sprintf("Operation:\n"+
		"\tType: %v\n"+
		"\tValue: %v\n"+
		"\tID: %v\n"+
		"\tTime: %s\n",
		inner.TypeTag, inner.ValueTag, inner.IDTag, inner.CreationTimeTag)
}

// separateFloat separate float to int part and check fractional for zero-value
func separateFloat(fields DirtyJSON, fl float64, errorText string) (int64, error) {
	intPart, frac := math.Modf(fl)
	if frac == 0 {
		return int64(intPart), nil
	}
	return 0, fmt.Errorf("%s%s: %w",
		fields, errorText, ErrRejectOperation)
}

// harmonizeValueWithType correlates the operation type with the value sign
func (operation *Operation) harmonizeValueWithType() {
	// todo: add overflow checking
	if (operation.Type == income && operation.Value < 0) ||
		(operation.Type == outcome && operation.Value > 0) {
		operation.Value *= -1
	}
}
