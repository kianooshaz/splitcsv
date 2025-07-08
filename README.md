# splitcsv

A simple command-line tool to split large CSV files into smaller files with a specified number of records per file.

## Installation

You need [Go](https://golang.org/dl/) installed (version 1.18 or newer recommended).

Install the binary using `go install`:

```sh
go install github.com/kianooshaz/splitcsv@latest
```

This will install the `splitcsv` executable in your `$GOPATH/bin` directory.

## Usage

```
splitcsv -input <input.csv> -out <output_prefix> -limit <max_records_per_file>
```

- `-input` (required): Path to the input CSV file.
- `-out` (optional): Prefix for the output files (default: `output`).
- `-limit` (optional): Maximum number of records per output file (default: `10000`).

### Example

Split a file named `bigdata.csv` into files with at most 5000 records each:

```
splitcsv -input bigdata.csv -out part -limit 5000
```

This will generate files like `part_1.csv`, `part_2.csv`, etc., each containing up to 5000 records (plus the header).

## License

See [LICENSE](LICENSE).