package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Config holds the configuration for CSV splitting
type Config struct {
	InputPath    string
	OutputPrefix string
	OutputDir    string
	MaxRecords   int
	BufferSize   int
	SkipEmpty    bool
	Delimiter    rune
	Verbose      bool
}

// CSVSplitter handles the CSV splitting operation
type CSVSplitter struct {
	config     Config
	partNumber int
	writer     *csv.Writer
	outFile    *os.File
}

func main() {
	config := parseFlags()

	if err := validateConfig(config); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		flag.Usage()
		os.Exit(1)
	}

	splitter := NewCSVSplitter(config)
	if err := splitter.Split(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if config.Verbose {
		fmt.Printf("Splitting completed successfully. Created %d files.\n", splitter.partNumber-1)
	}
}

// parseFlags parses command-line flags and returns a Config
func parseFlags() Config {
	config := Config{}

	flag.StringVar(&config.InputPath, "input", "", "Path to the input CSV file (required)")
	flag.StringVar(&config.InputPath, "i", "", "Path to the input CSV file (shorthand)")
	flag.StringVar(&config.OutputPrefix, "out", "output", "Prefix for the output files")
	flag.StringVar(&config.OutputPrefix, "o", "output", "Prefix for the output files (shorthand)")
	flag.StringVar(&config.OutputDir, "dir", ".", "Output directory for split files")
	flag.IntVar(&config.MaxRecords, "limit", 10000, "Maximum number of records per output file")
	flag.IntVar(&config.MaxRecords, "l", 10000, "Maximum number of records per output file (shorthand)")
	flag.IntVar(&config.BufferSize, "buffer", 64*1024, "Buffer size for file I/O in bytes")
	flag.BoolVar(&config.SkipEmpty, "skip-empty", true, "Skip empty records")
	flag.BoolVar(&config.Verbose, "verbose", false, "Enable verbose output")
	flag.BoolVar(&config.Verbose, "v", false, "Enable verbose output (shorthand)")

	delimiterStr := flag.String("delimiter", ",", "CSV delimiter character")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Split large CSV files into smaller chunks while preserving headers.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s -input data.csv -limit 5000\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -i data.csv -o chunk -dir ./output -l 1000 -v\n", os.Args[0])
	}

	flag.Parse()

	// Parse delimiter
	if len(*delimiterStr) == 1 {
		config.Delimiter = rune((*delimiterStr)[0])
	} else {
		config.Delimiter = ','
	}

	return config
}

// validateConfig validates the configuration
func validateConfig(config Config) error {
	if config.InputPath == "" {
		return fmt.Errorf("input file path is required")
	}

	if config.MaxRecords <= 0 {
		return fmt.Errorf("limit must be greater than 0")
	}

	if config.BufferSize <= 0 {
		return fmt.Errorf("buffer size must be greater than 0")
	}

	// Check if input file exists and is readable
	if _, err := os.Stat(config.InputPath); os.IsNotExist(err) {
		return fmt.Errorf("input file does not exist: %s", config.InputPath)
	}

	// Ensure output directory exists
	if err := os.MkdirAll(config.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	return nil
}

// NewCSVSplitter creates a new CSV splitter with the given configuration
func NewCSVSplitter(config Config) *CSVSplitter {
	return &CSVSplitter{
		config:     config,
		partNumber: 1,
	}
}

// Split performs the CSV splitting operation
func (s *CSVSplitter) Split() error {
	file, err := s.openInputFile()
	if err != nil {
		return err
	}
	defer file.Close()

	reader := s.createReader(file)
	header, err := s.readHeader(reader)
	if err != nil {
		return err
	}

	if s.config.Verbose {
		fmt.Printf("Starting to split CSV file: %s\n", s.config.InputPath)
		fmt.Printf("Max records per file: %d\n", s.config.MaxRecords)
	}

	recordCount := 0
	totalRecords := 0

	// Create first output file
	if err := s.createNewFile(header); err != nil {
		return err
	}
	defer s.closeCurrentFile()

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading record at line %d: %w", totalRecords+2, err)
		}

		totalRecords++

		// Skip empty records if configured
		if s.config.SkipEmpty && s.isEmptyRecord(record) {
			continue
		}

		// Check if we need to create a new file
		if recordCount >= s.config.MaxRecords {
			if err := s.createNewFile(header); err != nil {
				return err
			}
			recordCount = 0
		}

		// Write record to current file
		if err := s.writer.Write(record); err != nil {
			return fmt.Errorf("error writing record at line %d: %w", totalRecords+1, err)
		}
		recordCount++
	}

	if s.config.Verbose {
		fmt.Printf("Processed %d total records\n", totalRecords)
	}

	return nil
}

// openInputFile opens the input CSV file with buffering
func (s *CSVSplitter) openInputFile() (*os.File, error) {
	file, err := os.Open(s.config.InputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open input CSV file '%s': %w", s.config.InputPath, err)
	}
	return file, nil
}

// createReader creates a CSV reader with the configured options
func (s *CSVSplitter) createReader(file *os.File) *csv.Reader {
	reader := csv.NewReader(file)
	reader.Comma = s.config.Delimiter
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true
	return reader
}

// readHeader reads and validates the CSV header
func (s *CSVSplitter) readHeader(reader *csv.Reader) ([]string, error) {
	header, err := reader.Read()
	if err != nil {
		if err == io.EOF {
			return nil, fmt.Errorf("input file is empty")
		}
		return nil, fmt.Errorf("failed to read header: %w", err)
	}

	if len(header) == 0 {
		return nil, fmt.Errorf("header is empty")
	}

	return header, nil
}

// isEmptyRecord checks if a record contains only empty fields
func (s *CSVSplitter) isEmptyRecord(record []string) bool {
	for _, field := range record {
		if field != "" {
			return false
		}
	}
	return true
}

// createNewFile creates a new output file and initializes the writer
func (s *CSVSplitter) createNewFile(header []string) error {
	// Close previous file if it exists
	s.closeCurrentFile()

	// Generate output filename
	filename := fmt.Sprintf("%s_%d.csv", s.config.OutputPrefix, s.partNumber)
	filepath := filepath.Join(s.config.OutputDir, filename)

	// Create the output file
	outFile, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create output file '%s': %w", filepath, err)
	}

	// Create CSV writer
	s.outFile = outFile
	s.writer = csv.NewWriter(outFile)
	s.writer.Comma = s.config.Delimiter

	// Write header to new file
	if err := s.writer.Write(header); err != nil {
		s.closeCurrentFile()
		return fmt.Errorf("failed to write header to file '%s': %w", filepath, err)
	}

	if s.config.Verbose {
		fmt.Printf("Created output file: %s\n", filepath)
	}

	s.partNumber++
	return nil
}

// closeCurrentFile flushes and closes the current output file
func (s *CSVSplitter) closeCurrentFile() {
	if s.writer != nil {
		s.writer.Flush()
		s.writer = nil
	}
	if s.outFile != nil {
		s.outFile.Close()
		s.outFile = nil
	}
}
