# cpe2db

download cpe and import to database

## Install

```shell
go install github.com/akkuman/cpe2db
```

## Usage

```shell
$ ./cpe2db -h
NAME:
   cpe2db - download cpe and import to database

USAGE:
   cpe2db [global options]

GLOBAL OPTIONS:
   --output-file value, -o value  database file path
   --type value, -t value         database type (duckdb, sqlite3) (default: "duckdb")
   --help, -h                     show help
```