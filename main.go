package main

import (
	"database/sql"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"

	_ "github.com/mattn/go-sqlite3"
)

var databasePath string
var migrateFileDir string

var db *sql.DB
var tx *sql.Tx

var migratedVersion string
var reg = regexp.MustCompile(sqlFileFormat)

const initialVersion = "00000000-000000"
const datetimeFormat = "[0-9]{4}(0[1-9]|1[0-2])(0[1-9]|[12][0-9]|3[01])-([01][0-9]|2[0-3])([0-5][0-9])([0-5][0-9])"
const sqlFileFormat = "^" + datetimeFormat + "-.*.sql$"

// Get migration version from DB
// If there is no target table, create migration_version table
func initMigratedVersion() {
	var existCount int

	if err := db.QueryRow(
		`select count(*) from sqlite_master where type='table' and name='migration_version'`,
	).Scan(&existCount); err != nil {
		panic(err)
	}

	// Table not exists
	if existCount == 0 {
		_, err := db.Exec(`create table migration_version (key text, version text)`)
		if err != nil {
			fmt.Println("\nCouldnt create version table")
			panic(err)
		}
	}

	if err := db.QueryRow(
		`select count(*) from migration_version where key = ?`,
		migrateFileDir,
	).Scan(&existCount); err != nil {
		panic(err)
	}

	// Record not exists
	if existCount == 0 {
		if _, err := db.Exec(
			`insert into migration_version (key, version) VALUES (?, ?) `,
			migrateFileDir,
			initialVersion,
		); err != nil {
			panic(err)
		}
	}

	if err := db.QueryRow(
		`select version from migration_version where key = ?`,
		migrateFileDir,
	).Scan(&migratedVersion); err != nil {
		panic(err)
	}

	fmt.Println("Old version: " + migratedVersion)
}

// Register migratedVersion to DB
func registerMigratedVersion() {
	// Use tx
	if _, err := tx.Exec(
		`update migration_version set version = ? where key = ?`,
		migratedVersion,
		migrateFileDir,
	); err != nil {
		tx.Rollback()
		fmt.Println("\nCouldnt update version table\nrollback")
		panic(err)
	}
}

// Execute sql in target file (passed by argument)
func executeTargetFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
		fmt.Println("\nError when opening sqlFile")
		return err
	}
	defer file.Close()

	buffer, err := ioutil.ReadAll(file)

	fmt.Println(string(buffer))
	_, err = tx.Exec(string(buffer))
	if err != nil {
		tx.Rollback()
		fmt.Println(err)
		fmt.Println("\nFaild to Execute query")
		return err
	}

	return nil
}

// Read sql files in target directory (path is in global variable) and execute them
func readAndExecuteFiles() {
	// fileList automatically sorted by filename
	fileList, err := ioutil.ReadDir(migrateFileDir)
	if err != nil {
		panic(err)
	}

	for _, file := range fileList {
		var fileName = file.Name()

		// Execute when file is assigned format
		if !file.IsDir() && reg.MatchString(fileName) {
			fmt.Println("\n******" + fileName + "******")

			var newVersion = fileName[0:14] // yyyyMMdd-HHmmss
			// Skip if the file is older version
			if newVersion <= migratedVersion {
				fmt.Println("Past version: skipped")
				continue
			}

			var targetFilePath = filepath.Join(migrateFileDir, fileName)

			if err := executeTargetFile(targetFilePath); err != nil {
				fmt.Println("\nrollback")
				tx.Rollback()
				os.Exit(0)
			}

			migratedVersion = newVersion
		}
	}
}

func main() {
	migrateFileDir = "./migration"

	flag.Parse()
	var argDbPath = flag.Arg(0)
	var argMigrateDir = flag.Arg(1)

	if argDbPath == "" {
		fmt.Println("DB path must be passed to argument")
		os.Exit(0)
	}

	dbDir := filepath.Dir(argDbPath)
	fileInfo, err := os.Stat(dbDir)
	if err != nil || !fileInfo.IsDir() {
		fmt.Println("Database directory does not exist (" + dbDir + ")")
		os.Exit(0)
	}
	databasePath = argDbPath

	if argMigrateDir != "" {
		fileInfo, err := os.Stat(argMigrateDir)
		if err != nil || !fileInfo.IsDir() {
			fmt.Println("Migration file directory does not exist (" + argMigrateDir + ")")
			os.Exit(0)
		}

		migrateFileDir = argMigrateDir
	}

	db, err = sql.Open("sqlite3", databasePath)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	initMigratedVersion()

	// Transaction will committed after executing all files
	tx, err = db.Begin()
	if err != nil {
		panic(err)
	}

	readAndExecuteFiles()
	registerMigratedVersion()

	tx.Commit()
	fmt.Println("\ncommit")

	fmt.Println("\nNew version: " + migratedVersion)
}
