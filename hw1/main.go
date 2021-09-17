package main

import (
	"fmt"
	"os"
	"strconv"
	"unicode/utf8"
)

// ColorNumber describes color numbers (0, [31; 37])
type ColorNumber int

// SymbolOptions replaces structure and describes sandglass' settings,
// so it contains three "fields": symbol, color, size
// It's terrible crutch, sorry :(
type SymbolOptions map[string]string

// SymbolUpdater describes group of functions, which changes sandglasses' settings.
// It is necessary for closure mechanism and stable signature of the main function.
// If update is impossible than runner will use previous settings.
type SymbolUpdater func(options SymbolOptions) error

// There are color constants, which makes boundary conditions more obvious.
const (
	colorReset ColorNumber = 0
	colorRed   ColorNumber = 31
	colorWhite ColorNumber = 37
)

func main() {
	// run console client
	runCommandLineClient()
}

// Takes color number and returns it utf-8 value and error if it is impossible.
func (cn ColorNumber) getUTF() (string, error) {
	if (cn < colorRed || cn > colorWhite) && cn != colorReset {
		return fmt.Sprintf("\033[%dm", colorReset),
			fmt.Errorf("incorrect color number <%d>", cn)
	}
	return fmt.Sprintf("\033[%dm", cn), nil
}

// Concatenates color and symbol for convenient printing.
func (so SymbolOptions) String() string {
	// stops text coloring
	defaultColor, _ := colorReset.getUTF()
	return so["color"] + so["symbol"] + defaultColor
}

// Main function, which apply all user's settings from updaters to sandglass settings.
// Run printing function.
func runUpdaters(updaters ...SymbolUpdater) {
	// set up default settings
	color, _ := colorReset.getUTF()
	symbolOptions := SymbolOptions(map[string]string{
		"color":  color,
		"symbol": "X",
		"size":   "8"})

	for _, updater := range updaters {
		if err := updater(symbolOptions); err != nil {
			fmt.Println(err)
		}
	}

	// run printing function
	finalSymbol := symbolOptions.String()
	columnSize, _ := strconv.Atoi(symbolOptions["size"])
	printSandglass(finalSymbol, columnSize)
}

// Print sandglass
func printSandglass(finalSymbol string, columnSize int) {
	// returns minimal int
	min := func(a, b int) int {
		if a > b {
			return b
		}
		return a
	}
	// returns maximal int
	max := func(a, b int) int {
		if a < b {
			return b
		}
		return a
	}

	rowSize := columnSize/2 + 1
	for i := 0; i <= columnSize; i++ {
		// median
		if i == rowSize && columnSize%2 == 1 ||
			i == rowSize-1 && columnSize%2 == 0 {
			continue
		}
		// other rows
		for j := 0; j < columnSize; j++ {
			if j == min(i, columnSize-i) || j == max(i, columnSize-i)-1 ||
				i == 0 || i == columnSize {
				fmt.Print(finalSymbol)
			} else {
				fmt.Print(" ")
			}
		}
		fmt.Println()
	}
	fmt.Println()
}

// One of setting functions. Updates symbol if it is possible.
func changeSymbol(symbol string) SymbolUpdater {
	return func(symbolOptions SymbolOptions) error {
		if utf8.RuneCountInString(symbol) != 1 && symbol != " " {
			return fmt.Errorf("incorrect symbol <%s>", symbol)
		}
		symbolOptions["symbol"] = symbol
		return nil
	}
}

// One of setting functions. Updates color if it is possible.
func changeColor(colorNumberString string) SymbolUpdater {
	return func(symbolOptions SymbolOptions) error {
		colorNumberInt, err := strconv.Atoi(colorNumberString)
		if err != nil {
			return err
		}
		colorNumber := ColorNumber(colorNumberInt)
		symbolOptions["color"], err = colorNumber.getUTF()
		return err
	}
}

// One of setting functions. Updates size if it is possible.
func changeSize(stringSize string) SymbolUpdater {
	return func(symbolOptions SymbolOptions) error {
		size, err := strconv.Atoi(stringSize)
		if err != nil {
			return err
		}
		if size <= 0 || size > 1000 {
			return fmt.Errorf("incorrect size number <%d>", size)
		}
		symbolOptions["size"] = fmt.Sprint(size)
		return nil
	}
}

// Runs command line client
func runCommandLineClient() {
	// sequence of setting functions (updaters)
	var commandNumber int
	commands := make([]SymbolUpdater, 0, 10)
	fmt.Println("Welcome to sandglass app!\n" +
		"Follow instructions...")
	// loops possibility to enter commands until "Exit"
	for {
		fmt.Println("Enter command number:\n" +
			"1. Change size\n" +
			"2. Change symbol\n" +
			"3. Change color\n" +
			"4. Run\n" +
			"5. Exit")
		_, err := fmt.Scanf("%d", &commandNumber)
		if err != nil || commandNumber < 1 || commandNumber > 5 {
			fmt.Println("incorrect command number")
			continue
		}
		// run command
		switchCommandNumber(&commands, commandNumber)
	}
}

// Runs command by it's number.
func switchCommandNumber(commands *[]SymbolUpdater, commandNumber int) {
	switch commandNumber {
	case 1:
		// append size updater
		fmt.Println("Enter bottom side size (int):")
		*commands = append(*commands, changeSize(enterNewOption()))
	case 2:
		// append symbol updater
		fmt.Println("Enter utf-8 symbol (just one):")
		*commands = append(*commands, changeSymbol(enterNewOption()))
	case 3:
		// append color updater
		fmt.Println("Choose color and enter it's number (int):")
		fmt.Println("0\t default\n" +
			"31\tRed\n" +
			"32\tGreen\n" +
			"33\tYellow\n" +
			"34\tBlue\n" +
			"35\tPurple\n" +
			"36\tCyan\n" +
			"37\tcolorWhite")
		*commands = append(*commands, changeColor(enterNewOption()))
	case 4:
		// execute main function and drop slice of updaters
		runUpdaters(*commands...)
		*commands = make([]SymbolUpdater, 0, 10)
	case 5:
		// exit
		fmt.Println("Goodbye!")
		os.Exit(0)
	}
}

// Enters new setting for updaters.
func enterNewOption() string {
	var option string
	// loops if input isn't ok
	for {
		fmt.Println("Enter value:")

		if _, err := fmt.Scanln(&option); err == nil {
			break
		}
		fmt.Println("Something went wrong, try again...")
	}
	return option
}
