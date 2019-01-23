package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"

	_ "github.com/mattn/go-sqlite3"
)

var databaseName string
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

	db, err := sql.Open("sqlite3", databaseName)
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
		panic(err)
	}

	transaction.Commit()
	fmt.Println("commit")

}

func main() {
	databaseName = "./test.db"
	migrateFileDir = "./migration"

	readFileList()
}
