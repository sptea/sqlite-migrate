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

func readFileList() {
	// fileList automatically sorted by filename
	fileList, err := ioutil.ReadDir(migrateFileDir)
	if err != nil {
		panic(err)
	}

	for _, file := range fileList {
		var fileName = file.Name()
		reg := regexp.MustCompile("^*.sql$")
		if !file.IsDir() && reg.MatchString(fileName) {
			fmt.Println("import: " + filepath.Join(migrateFileDir, fileName))
			executeTargetSQL(filepath.Join(migrateFileDir, fileName))
		}
	}
}

func executeTargetSQL(path string) {
	file, err := os.Open(path)
	if err != nil {
		fmt.Println("error")
	}
	defer file.Close()

	buffer, err := ioutil.ReadAll(file)

	db, err := sql.Open("sqlite3", databasePath)
	if err != nil {
		panic(err)
	}

	transaction, err := db.Begin()
	if err != nil {
		panic(err)
	}

	fmt.Println(string(buffer))
	_, err = transaction.Exec(string(buffer))
	if err != nil {
		transaction.Rollback()
		fmt.Println("rollback")
		os.Exit(0)
	}

	transaction.Commit()
	fmt.Println("commit")

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

	readFileList()
}
