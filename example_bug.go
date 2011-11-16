package seven5

import "fmt"
import "os"
import "exp/sql"
import _ "github.com/mattn/go-sqlite3" //force linkage,without naming the library

func show_bug() {
	db, _ := sql.Open("sqlite3", "./foo.sqlite")

	//this is a ddl operation, so LastInsertId and RowsAffected should not work
	_, _ = db.Exec("create table foo (id integer)")

	// this is a data operation that is doing an insert
	result, err := db.Exec("insert into foo values (123)")
	if err != nil {
		fmt.Fprintf(os.Stderr, "insert failed:%s\n", err)
	}

	lastId, err := result.LastInsertId()
	if err != nil {
		fmt.Fprintf(os.Stderr, "last id blew up:%s\n", err)
	} else {
		fmt.Fprintf(os.Stderr, "last id was:%d\n", lastId)
	}
}
