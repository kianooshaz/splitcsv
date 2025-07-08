package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func main() {
	inputPath := flag.String("input", "", "Path to the input CSV file")
	outputPrefix := flag.String("out", "output", "Prefix for the output files")
	maxRecords := flag.Int("limit", 10000, "Maximum number of records per output file")
	flag.Parse()

	if *inputPath == "" {
		fmt.Println("Error: -input is required")
		flag.Usage()
		os.Exit(1)
	}

	if err := splitCSV(*inputPath, *outputPrefix, *maxRecords); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	fmt.Println("Splitting completed successfully.")
}

func splitCSV(inputPath, outputPrefix string, maxRecords int) error {
	file, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open input CSV: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	header, err := reader.Read()
	if err != nil {
		return fmt.Errorf("failed to read header: %w", err)
	}

	var (
		part        = 1
		recordCount = 0
		writer      *csv.Writer
		outFile     *os.File
	)

	createNewFile := func() error {
		if writer != nil {
			writer.Flush()
			outFile.Close()
		}
		outFileName := fmt.Sprintf("%s_%d.csv", outputPrefix, part)
		outFilePath := filepath.Join(".", outFileName)

		outFile, err = os.Create(outFilePath)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		writer = csv.NewWriter(outFile)
		if err := writer.Write(header); err != nil {
			return fmt.Errorf("failed to write header: %w", err)
		}
		recordCount = 0
		part++
		return nil
	}

	if err := createNewFile(); err != nil {
		return err
	}

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading record: %w", err)
		}

		// Skip empty record (e.g. empty newline at EOF)
		isEmpty := true
		for _, field := range record {
			if field != "" {
				isEmpty = false
				break
			}
		}
		if isEmpty {
			continue
		}

		if recordCount >= maxRecords {
			if err := createNewFile(); err != nil {
				return err
			}
		}

		if err := writer.Write(record); err != nil {
			return fmt.Errorf("error writing record: %w", err)
		}
		recordCount++
	}

	if writer != nil {
		writer.Flush()
		outFile.Close()
	}
	return nil
}
