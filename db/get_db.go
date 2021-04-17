package db

import "database/sql"

var DataBase *sql.DB

func UploadDataBase() {
	var err error
	DataBase, err = sql.Open("mysql", "root:@/test")
	if err != nil {
		panic(err.Error())
	}
}
