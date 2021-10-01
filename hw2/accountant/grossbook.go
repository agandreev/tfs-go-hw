package accountant

import (
	"fmt"
	"sort"
)

// GrossBook represents map of string keys (company name) and Balance values.
type GrossBook map[string]*Balance

// InvalidOperationsSlice describes dynamic slice of invalid operation's ids
type InvalidOperationsSlice []interface{}

// Len is the number of elements in the collection.
func (operations InvalidOperationsSlice) Len() int { return len(operations) }

// Less reports whether the element with
// index i should sort before the element with index j.
func (operations InvalidOperationsSlice) Less(i, j int) bool {
	return castIDToString(operations[i]) < castIDToString(operations[j])
}

// Swap swaps the elements with indexes i and j.
func (operations InvalidOperationsSlice) Swap(i, j int) {
	operations[i], operations[j] = operations[j], operations[i]
}

// Balance represents final balance for specific company.
type Balance struct {
	Company              string                 `json:"company"`
	ValidOperationsCount int64                  `json:"valid_operations_count"`
	Balance              int64                  `json:"balance"`
	InvalidOperations    InvalidOperationsSlice `json:"invalid_operations"`
}

// AddOperation add new key and Balance struct to GrossBook. It is sensitive for
// Operation status (valid, invalid, skip).
func (gb GrossBook) AddOperation(operation Operation) {
	// if operation should be skipped
	if operation.Status == skip {
		return
	}
	// create new Balance if it doesn't exist
	_, ok := gb[operation.Company]
	if !ok {
		stock := &Balance{
			Company:           operation.Company,
			InvalidOperations: make([]interface{}, 0),
		}
		gb[operation.Company] = stock
	}
	// take stock again for updated case
	stock := gb[operation.Company]
	// if operation should be rejected
	if operation.Status == invalid {
		stock.InvalidOperations = append(stock.InvalidOperations, operation.ID)
		return
	}
	// common way
	if operation.Status == valid {
		stock.ValidOperationsCount++
		stock.Balance += operation.Value
	}
}

// SortedBalances represent GrossBook as sorted Balance slice for convenient processing.
func (gb GrossBook) SortedBalances() []*Balance {
	// fill array by GrossBook values
	operations := make([]*Balance, 0, len(gb))
	for _, v := range gb {
		sort.Sort(v.InvalidOperations)
		operations = append(operations, v)
	}
	// sorting
	sort.SliceStable(operations, func(i, j int) bool {
		return operations[i].Company < operations[j].Company
	})
	return operations
}

func castIDToString(id interface{}) string {
	switch idSwitcher := id.(type) {
	case int64:
		return fmt.Sprintf("%d", idSwitcher)
	case string:
		return idSwitcher
	}
	fmt.Println(id)
	panic("unexpected value in castIDToString function")
}
