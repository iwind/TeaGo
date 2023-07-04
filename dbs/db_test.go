package dbs

import (
	"encoding/json"
	"github.com/iwind/TeaGo/logs"
	"log"
	"testing"
	"time"
)

func TestDBName(t *testing.T) {
	db, err := Instance("db2")
	if err != nil {
		log.Fatal(err)
	}

	t.Log("Name:", db.Name())
}

func TestDBFindTable(t *testing.T) {
	db, err := Instance("dev")
	if err != nil {
		log.Fatal(err)
	}

	table, err := db.FindTable("pp_users")
	if err != nil {
		log.Fatal(err)
	}
	jsonBytes, _ := json.MarshalIndent(table, "", "   ")
	t.Log(string(jsonBytes))
}

func TestDB_FindFullTable(t *testing.T) {
	db, err := Instance("dev")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	table, err := db.FindFullTable("pp_shopPlayLogs")
	if err != nil {
		log.Fatal(err)
	}
	jsonBytes, _ := json.MarshalIndent(table.Partitions, "", "   ")
	t.Log(string(jsonBytes))

	indexes, _ := json.MarshalIndent(table.Indexes, "", "   ")
	t.Log(string(indexes))
}

func TestDB_FindFunctions(t *testing.T) {
	db, err := Instance("dev")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	functions, err := db.FindFunctions()
	if err != nil {
		t.Fatal(err)
	}

	data, _ := json.MarshalIndent(functions, "", "   ")
	t.Log(string(data))
}

func TestDBTableNames(t *testing.T) {
	db, err := Instance("db1")
	if err != nil {
		log.Fatal(err)
	}

	log.Println(db.TableNames())
}

func TestDB_FindPreparedOnes(t *testing.T) {
	db, err := NewInstanceFromConfig(&DBConfig{
		Driver: "mysql",
		Dsn:    "root:123456@tcp(127.0.0.1:3306)/db_edge?charset=utf8mb4&timeout=30s",
		Prefix: "edge",
	})
	if err != nil {
		t.Fatal(err)
	}

	var count = 10000
	var before = time.Now()
	for i := 0; i < count; i++ {
		_, _, err := db.FindPreparedOnes("SELECT 1")
		if err != nil {
			t.Fatal(err)
		}
	}
	t.Log("FindPreparedOnes():", time.Since(before).Seconds()*1000, "ms")

	before = time.Now()
	for i := 0; i < count; i++ {
		_, _, err := db.FindOnes("SELECT 1")
		if err != nil {
			t.Fatal(err)
		}
	}
	t.Log("FindOnes:", time.Since(before).Seconds()*1000, "ms")
}

func TestDB_FindOnes(t *testing.T) {
	db, err := Default()
	if err != nil {
		t.Fatal(err)
	}

	ones, columnNames, err := db.FindOnes("SELECT id, name FROM users LIMIT 2")
	if err != nil {
		t.Fatal(err)
	}

	t.Log(columnNames)
	logs.PrintAsJSON(ones, t)
}

func TestDB_FindOne(t *testing.T) {
	db, err := Default()
	if err != nil {
		t.Fatal(err)
	}

	one, err := db.FindOne("SELECT id, name FROM users WHERE id=?", 1)
	if err != nil {
		t.Fatal(err)
	}

	logs.PrintAsJSON(one, t)
}

func TestDB_FindCol(t *testing.T) {
	db, err := Default()
	if err != nil {
		t.Fatal(err)
	}

	colValue, err := db.FindCol(1, "SELECT id, name FROM users WHERE id=?", 1)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("col:", colValue)
}

func TestDB_FindCol_Empty(t *testing.T) {
	db, err := Default()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = db.Close()
	}()

	colValue, err := db.FindCol(1, "SELECT id, name FROM users WHERE id=?", 2)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("col:", colValue)
}

func TestDB_MultipleStatements(t *testing.T) {
	db, err := Default()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = db.Close()
	}()

	one, err := db.FindOne("UPDATE users SET state=1 WHERE id=1; SELECT state FROM users WHERE id=1")
	if err != nil {
		t.Fatal(err)
	}
	logs.PrintAsJSON(one, t)
}

func BenchmarkDB_FindOne(b *testing.B) {
	db, err := Default()
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		func() {
			_, err = db.FindOne("SELECT id, name FROM users LIMIT 1")
			if err != nil {
				b.Fatal(err)
			}
		}()
	}
}
