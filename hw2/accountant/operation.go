package accountant

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

const (
	// describes operation's statuses
	valid OperationStatus = iota
	invalid
	skip

	// describes JSON's fields
	companyField   = "company"
	typeField      = "type"
	valueField     = "value"
	idField        = "id"
	timeField      = "created_at"
	operationField = "operation"

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

// DirtyJSON describes unmarshalled JSON into map of string key (field name)
// and empty interface as value of the field
type DirtyJSON map[string]interface{}

// Operation describes structure of entire JSON and it's status of condition
// (valid, invalid, rejected)
type Operation struct {
	Company string
	Type    string
	Value   int64
	ID      string
	Time    time.Time
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
		return err
	}

	// check received map for nestedMaps and concat fields
	if nestedMap, ok := fields[operationField]; ok {
		for field, value := range nestedMap.(map[string]interface{}) {
			// check for duplicated fields from nested map
			if _, ok := fields[field]; ok {
				return fmt.Errorf("%sthe same fields in one structure", fields)
			}
			fields[field] = value
		}
		delete(fields, operationField)
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
	companyTag, ok := dj[companyField]
	// check for tag existence
	if !ok {
		return emptyString, fmt.Errorf("%sjson hasn't company tag: %w",
			dj, ErrSkipOperation)
	}

	// try to assert
	switch company := companyTag.(type) {
	case string:
		if company == emptyString {
			return emptyString, fmt.Errorf("%scompany tag content is empty: %w",
				dj, ErrSkipOperation)
		}
		return company, nil
	default:
		return emptyString, fmt.Errorf("%scan't cast company tag content to string: %w",
			dj, ErrSkipOperation)
	}
}

// Type returns string implementation of type from DirtyJSON
func (dj DirtyJSON) Type() (string, error) {
	typeTag, ok := dj[typeField]
	// check for tag existence
	if !ok {
		return emptyString, fmt.Errorf("%sjson hasn't type tag: %w",
			dj, ErrRejectOperation)
	}

	// try to assert
	switch typeSwitcher := typeTag.(type) {
	case string:
		// check for outcome value
		for _, typeValue := range typeValues[outcome] {
			if typeValue == typeSwitcher {
				return outcome, nil
			}
		}
		// check for income value
		for _, typeValue := range typeValues[income] {
			if typeValue == typeSwitcher {
				return income, nil
			}
		}
		return emptyString, fmt.Errorf("%stype tag content is illigal: %w",
			dj, ErrRejectOperation)
	default:
		return emptyString, fmt.Errorf("%scan't cast type tag content to string: %w",
			dj, ErrRejectOperation)
	}
}

// Value returns int64 implementation of value from DirtyJSON
func (dj DirtyJSON) Value() (int64, error) {
	valueTag, ok := dj[valueField]
	// check for tag existence
	if !ok {
		return 0, fmt.Errorf("%sjson hasn't value tag: %w",
			dj, ErrRejectOperation)
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
func (dj DirtyJSON) ID() (string, error) {
	idTag, ok := dj[idField]
	// check for tag existence
	if !ok {
		return emptyString, fmt.Errorf("%sjson hasn't id tag: %w",
			dj, ErrSkipOperation)
	}

	// try to assert
	switch id := idTag.(type) {
	case float64:
		// separation of int part
		intPart, frac := math.Modf(id)
		if frac == 0 {
			return fmt.Sprintf("%d", int64(intPart)), nil
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
func (dj DirtyJSON) Time() (time.Time, error) {
	timeTag, ok := dj[timeField]
	if !ok {
		return time.Now(), fmt.Errorf("%sjson hasn't time tag: %w",
			dj, ErrSkipOperation)
	}

	// try to assert
	switch timeValue := timeTag.(type) {
	case string:
		timeLayouted, err := time.Parse(time.RFC3339, timeValue)
		if err != nil {
			return time.Now(), fmt.Errorf("%stime tag content is broken format: %w",
				dj, ErrSkipOperation)
		}
		return timeLayouted, nil
	default:
		return time.Now(), fmt.Errorf("%scan't cast time tag content to string: %w",
			dj, ErrSkipOperation)
	}
}

// String represents DirtyJSON as string through string builder.
func (dj DirtyJSON) String() string {
	sb := strings.Builder{}
	for k, v := range dj {
		sb.WriteString(fmt.Sprintf("%s\t%s\n", k, v))
	}
	return sb.String()
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
