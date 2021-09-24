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

	emptyString = ""
	// if you want to see operation processing
	printErrors = true
)

var (
	// available Operation types (can be expanded)
	typeValues = [...]string{"income", "outcome", "+", "-"}

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
	updateOperation(operation, fields)
	return nil
}

// updateOperation enter fields (field by field) from dirtyJSON to Operation instance.
// And updates Operation instance's status.
func updateOperation(operation *Operation, fields DirtyJSON) {
	var err error
	if operation.Company, err = processCompanyTag(fields); err != nil {
		switchErrors(operation, err)
		return
	}
	if operation.Time, err = processTimeTag(fields); err != nil {
		switchErrors(operation, err)
		return
	}
	if operation.ID, err = processIDTag(fields); err != nil {
		switchErrors(operation, err)
		return
	}
	if operation.Type, err = processTypeTag(fields); err != nil {
		switchErrors(operation, err)
		return
	}
	if operation.Value, err = processValueTag(fields); err != nil {
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

// processCompanyTag returns string implementation of company from DirtyJSON
func processCompanyTag(fields DirtyJSON) (string, error) {
	companyTag, ok := fields[companyField]
	// check for tag existence
	if !ok {
		return emptyString, fmt.Errorf("%sjson hasn't company tag: %w",
			fields, ErrSkipOperation)
	}

	// try to assert
	switch company := companyTag.(type) {
	case string:
		if company == emptyString {
			return emptyString, fmt.Errorf("%scompany tag content is empty: %w",
				fields, ErrSkipOperation)
		}
		return company, nil
	default:
		return emptyString, fmt.Errorf("%scan't cast company tag content to string: %w",
			fields, ErrSkipOperation)
	}
}

// processTypeTag returns string implementation of type from DirtyJSON
func processTypeTag(fields DirtyJSON) (string, error) {
	typeTag, ok := fields[typeField]
	// check for tag existence
	if !ok {
		return emptyString, fmt.Errorf("%sjson hasn't type tag: %w",
			fields, ErrRejectOperation)
	}

	// try to assert
	switch typeSwitcher := typeTag.(type) {
	case string:
		// check for income, outcome, +, -
		for _, typeValue := range typeValues {
			if typeValue == typeSwitcher {
				return typeSwitcher, nil
			}
		}
		return emptyString, fmt.Errorf("%stype tag content is illigal: %w",
			fields, ErrRejectOperation)
	default:
		return emptyString, fmt.Errorf("%scan't cast type tag content to string: %w",
			fields, ErrRejectOperation)
	}
}

// processValueTag returns int64 implementation of value from DirtyJSON
func processValueTag(fields DirtyJSON) (int64, error) {
	valueTag, ok := fields[valueField]
	// check for tag existence
	if !ok {
		return 0, fmt.Errorf("%sjson hasn't value tag: %w",
			fields, ErrRejectOperation)
	}

	// try to assert
	switch value := valueTag.(type) {
	case float64:
		// separation of the int part
		return separateFloat(fields, value, "value tag content has fractional: ")
	case string:
		// parsing
		floatValue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return 0, fmt.Errorf("%svalue tag content isn't a float: %w",
				fields, ErrRejectOperation)
		}
		// separation of the int part
		return separateFloat(fields, floatValue, "value tag content has fractional: ")
	default:
		return 0, fmt.Errorf("%svalue tag content isn't a digit: %w",
			fields, ErrRejectOperation)
	}
}

// processIDTag returns string implementation of id from DirtyJSON
func processIDTag(fields DirtyJSON) (string, error) {
	idTag, ok := fields[idField]
	// check for tag existence
	if !ok {
		return emptyString, fmt.Errorf("%sjson hasn't id tag: %w",
			fields, ErrSkipOperation)
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
			fields, ErrRejectOperation)
	case string:
		if id == emptyString {
			return emptyString, fmt.Errorf("%stime tag is empty: %w",
				fields, ErrSkipOperation)
		}
		return id, nil
	default:
		return emptyString, fmt.Errorf("%scan't cast id tag to string %w",
			fields, ErrSkipOperation)
	}
}

// processTimeTag returns time.Time implementation of created_at from DirtyJSON
func processTimeTag(fields DirtyJSON) (time.Time, error) {
	timeTag, ok := fields[timeField]
	if !ok {
		return time.Now(), fmt.Errorf("%sjson hasn't time tag: %w",
			fields, ErrSkipOperation)
	}

	// try to assert
	switch timeValue := timeTag.(type) {
	case string:
		timeLayouted, err := time.Parse(time.RFC3339, timeValue)
		if err != nil {
			return time.Now(), fmt.Errorf("%stime tag content is broken format: %w",
				fields, ErrSkipOperation)
		}
		return timeLayouted, nil
	default:
		return time.Now(), fmt.Errorf("%scan't cast time tag content to string: %w",
			fields, ErrSkipOperation)
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
