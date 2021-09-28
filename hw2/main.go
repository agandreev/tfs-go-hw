package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/agandreev/tfs-go-hw/hw2/accountant"
)

const (
	// env var name
	envFile = "ENV FILE"
	outFile = "out.json"
)

var (
	// path to json file
	filePath string
)

// ArgsReader describes function, which reads arguments from individual source
type ArgsReader func()

// init processes reading argument from different sources by priority
func init() {
	// uncomment for convenient way to add env var
	// os.Setenv(envFile, "env file path")
	// slice of sources could be expanded
	argsReaders := []ArgsReader{loadFlags(), loadEnv(), loadInput()}
	for _, argsReader := range argsReaders {
		argsReader()
		if filePath != "" {
			break
		}
	}
}

func main() {
	if err := runBalanceCounter(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// runBalanceCounter runs app logic step by step
func runBalanceCounter() error {
	// check file and read if it's possible
	data, err := processFileReading()
	if err != nil {
		return err
	}

	// unmarshalling
	var operations []accountant.Operation
	err = json.Unmarshal(data, &operations)
	if err != nil {
		return err
	}
	// sort operations for convenient processing and representations
	sort.SliceStable(operations, func(i, j int) bool {
		return operations[i].ID < operations[j].ID
	})

	// add operations to gross book
	gb := accountant.GrossBook{}
	for _, operation := range operations {
		gb.AddOperation(operation)
	}

	// marshal balances from gross book
	data, err = json.MarshalIndent(gb.SortedBalances(), "", "\t")
	if err != nil {
		return err
	}

	// write bytes to JSON file
	err = processFileWriting(data)
	if err != nil {
		return err
	}
	return nil
}

// processFileReading check existence of file (not directory) by path
func processFileReading() ([]byte, error) {
	// file opening
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("problems in file opening: %w", err)
	}
	defer file.Close()

	// checking for directory
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("problems in file path: %w", err)
	}
	if fileInfo.IsDir() {
		return nil, fmt.Errorf("given file is directory")
	}

	// reading
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("problems in file reading: %w", err)
	}
	return data, nil
}

// processFileWriting writes data to outFile
func processFileWriting(data []byte) error {
	file, err := os.Create(outFile)
	if err != nil {
		return fmt.Errorf("problems in file creating: %w", err)
	}
	defer file.Close()
	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("problems in file writing: %w", err)
	}
	return nil
}

// loadEnv returns ArgsReader function,
// which reads command line arg "--file"
func loadFlags() ArgsReader {
	return func() {
		flag.StringVar(&filePath, "file", "", "path to json")
		flag.Parse()
	}
}

// loadEnv returns ArgsReader function,
// which reads const environment variable envFile
func loadEnv() ArgsReader {
	return func() {
		filePath = os.Getenv(envFile)
	}
}

// loadInput returns ArgsReader function, which reads user's input
func loadInput() ArgsReader {
	return func() {
		// infinite loop makes user to enter non-empty string
		for {
			fmt.Println("Enter path to json file:")
			_, err := fmt.Scanf("%s", &filePath)
			if err != nil || len(filePath) == 0 {
				fmt.Println("Incorrect string entered...")
				continue
			}
			break
		}
	}
}
