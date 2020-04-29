package data

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

var dbType string = "sqlite3"

type MysqlDatabase struct {
	Db *sql.DB
}

func NewMysqlDatabase() *MysqlDatabase {
	ret := MysqlDatabase{}

	return &ret
}

func (this *MysqlDatabase) Open(path string) error {
	var err error
	this.Db, err = sql.Open(dbType, path)
	if err != nil {
		return fmt.Errorf("[MysqlDatabase][Open] Cannot open alias\n%s", err.Error())
	}
	return nil
}

func (this *MysqlDatabase) Close() {
	this.Db.Close()
}

func (this *MysqlDatabase) Begin() (*sql.Tx, error) {
	return this.Db.Begin()
}

func (this *MysqlDatabase) Query(query string, args ...interface{}) (*sql.Rows, error) {
	rows, err := this.Db.Query(query, args...)
	return rows, err
}

func (this *MysqlDatabase) Prepare(query string) (*sql.Stmt, error) {
	statement, err := this.Db.Prepare(query)
	if err != nil {
		return nil, fmt.Errorf("[MysqlDatabase][Prepare] Problem preparing statement\n%s", err.Error())
	}
	return statement, err
}

func (this *MysqlDatabase) Execute(query string, args ...interface{}) (sql.Result, error) {
	statement, err := this.Db.Prepare(query)
	if err != nil {
		return nil, fmt.Errorf("[MysqlDatabase][Execute] Problem preparing statement\n%s", err.Error())
	}
	result, err := statement.Exec(args...)
	if err != nil {
		return nil, fmt.Errorf("[MysqlDatabase][Execute] Problem executing statement\n%s", err.Error())
	}
	return result, err
}
