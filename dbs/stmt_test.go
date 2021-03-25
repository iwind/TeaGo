package dbs

import (
	"strconv"
	"testing"
	"time"
)

func TestDB_Prepare(t *testing.T) {
	db, err := NewInstanceFromConfig(&DBConfig{
		Driver: "mysql",
		Dsn:    "root:123456@tcp(127.0.0.1:3306)/db_edge?charset=utf8mb4&timeout=30s",
		Prefix: "edge",
	})
	if err != nil {
		t.Fatal(err)
	}
	/**_, err = db.Exec("SET GLOBAL max_prepared_stmt_count=65535")
	if err != nil {
		t.Fatal(err)
	}**/

	for i := 0; i < 20000; i++ {
		stmt, err := db.Prepare("SELECT " + strconv.Itoa(i))
		if err != nil {
			t.Fatal(err)
		}
		_ = stmt.Close()
	}
}

func TestDB_PrepareOnce(t *testing.T) {
	db, err := NewInstanceFromConfig(&DBConfig{
		Driver: "mysql",
		Dsn:    "root:123456@tcp(127.0.0.1:3306)/db_edge?charset=utf8mb4&timeout=30s",
		Prefix: "edge",
	})
	if err != nil {
		t.Fatal(err)
	}
	/**_, err = db.Exec("SET GLOBAL max_prepared_stmt_count=65535")
	if err != nil {
		t.Fatal(err)
	}**/

	for i := 0; i < 20000; i++ {
		stmt, err := db.PrepareOnce("SELECT " + strconv.Itoa(i))
		if err != nil {
			t.Fatal(err)
		}
		stmt.accessAt = time.Now().Unix() - 3600 - 1
	}
}
