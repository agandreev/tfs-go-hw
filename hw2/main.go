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

// ArgsReader describes function, which reads arguments from individual source
type ArgsReader func() string

func main() {
	if err := runBalanceCounter(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// readPath processes reading path-argument from different sources by priority
func readPath() string {
	var filePath string
	// uncomment for convenient way to add env var
	// os.Setenv(envFile, "env file path")
	// slice of sources could be expanded
	argsReaders := []ArgsReader{loadFlags(), loadEnv()}
	for _, argsReader := range argsReaders {
		filePath = argsReader()
		if filePath != "" {
			return filePath
		}
	}
	return filePath
}

// runBalanceCounter runs app logic step by step
func runBalanceCounter() error {
	// check file and read if it's possible
	data, err := processDataReading()
	if err != nil {
		return err
	}

	// unmarshalling
	var operations accountant.Operations
	err = json.Unmarshal(data, &operations)
	if err != nil {
		return err
	}
	sort.Sort(operations)

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

// processDataReading check existence of file (not directory) by path
func processDataReading() ([]byte, error) {
	if filePath := readPath(); filePath != "" {
		return readFile(filePath)
	}
	return readInput()
}

// readFile reads file
func readFile(filePath string) ([]byte, error) {
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

// readInput reads stdin
func readInput() ([]byte, error) {
	// reading
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return nil, fmt.Errorf("problems in stdin reading: %w", err)
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
	return func() string {
		var filePath string
		flag.StringVar(&filePath, "file", "", "path to json")
		flag.Parse()
		return filePath
	}
}

// loadEnv returns ArgsReader function,
// which reads const environment variable envFile
func loadEnv() ArgsReader {
	return func() string {
		filePath := os.Getenv(envFile)
		return filePath
	}
}
