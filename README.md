# CSV Splitter

A robust and flexible command-line tool to split large CSV files into smaller chunks while preserving headers and maintaining data integrity.

## Features

- **Flexible Configuration**: Multiple command-line options for customization
- **Performance Optimized**: Efficient memory usage and I/O operations
- **Error Handling**: Comprehensive error reporting with line numbers
- **Multiple Delimiters**: Support for different CSV delimiter characters
- **Verbose Output**: Optional detailed progress information
- **Empty Record Handling**: Configurable skipping of empty records
- **Output Directory Control**: Specify custom output directories

## Installation

You need [Go](https://golang.org/dl/) installed (version 1.18 or newer recommended).

### Build from Source

```bash
git clone https://github.com/kianooshaz/splitcsv.git
cd splitcsv
go build -o csvplit
```

### Install via go install

```bash
go install github.com/kianooshaz/splitcsv@latest
```

## Usage

### Basic Usage

```bash
./csvplit -input data.csv -limit 5000
```

### Advanced Usage

```bash
./csvplit -i data.csv -o chunk -dir ./output -l 1000 -delimiter ";" -v
```

### Command Line Options

| Flag | Shorthand | Default | Description |
|------|-----------|---------|-------------|
| `-input` | `-i` | *required* | Path to the input CSV file |
| `-out` | `-o` | `output` | Prefix for the output files |
| `-limit` | `-l` | `10000` | Maximum number of records per output file |
| `-dir` | | `.` | Output directory for split files |
| `-delimiter` | | `,` | CSV delimiter character |
| `-buffer` | | `65536` | Buffer size for file I/O in bytes |
| `-skip-empty` | | `true` | Skip empty records |
| `-verbose` | `-v` | `false` | Enable verbose output |
| `-help` | `-h` | | Show help message |

### Examples

**Split a large CSV file into chunks of 5000 records:**

```bash
./csvplit -input bigdata.csv -limit 5000
```

**Use a custom output prefix and directory:**

```bash
./csvplit -i data.csv -o part -dir ./chunks
```

**Process a semicolon-delimited file with verbose output:**

```bash
./csvplit -i data.csv -delimiter ";" -v
```

**Split with custom buffer size for better performance:**

```bash
./csvplit -i largefile.csv -buffer 131072 -l 10000
```

## Output

The tool creates numbered output files with the format: `{prefix}_{number}.csv`

For example, with `-out part`, you'll get:

- `part_1.csv`
- `part_2.csv`
- `part_3.csv`
- etc.

Each output file includes:

- The original CSV header as the first line
- Up to the specified number of data records
- Proper CSV formatting with the same delimiter as the input

## Error Handling

The tool provides detailed error messages including:

- File access issues
- CSV parsing errors with line numbers
- Configuration validation errors
- I/O operation failures

## Performance Considerations

- **Memory Efficient**: Processes files in streaming fashion
- **Configurable Buffering**: Adjust buffer size for optimal I/O performance
- **Large File Support**: Can handle files larger than available RAM

## Requirements

- Go 1.18 or newer
- Read access to input CSV file
- Write access to output directory

## License

See [LICENSE](LICENSE).
