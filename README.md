## Usage

``` shell
go run main.go [Database file path] [Migration file directory(optional: default=./migration)]
```

e.g.

``` shell
go run main.go ./test.db ./migration/
```

## Specification

- File path should be relative
- Migration will be committed independently for each migration-directory path
- Migration file name should be below

``` shell
yyyyMMdd-HHmmss-xxxxxxxxx.sql
```
