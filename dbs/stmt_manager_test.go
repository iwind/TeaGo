// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package dbs

import (
	"runtime"
	"strconv"
	"testing"
	"time"
)

func TestStmtManager_PrepareOnce(t *testing.T) {
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

	for i := 0; i < 100_000; i++ {
		stmt, cached, err := db.PrepareOnce("SELECT " + strconv.Itoa(i))
		if err != nil {
			t.Log("cached:", cached)
			t.Fatal(err)
		}
		if !cached {
			_ = stmt.Close()
		}
		_ = stmt
	}

	t.Log(db.StmtManager().Len())
}

func TestStmtManager_Prepare(t *testing.T) {
	db, err := NewInstanceFromConfig(&DBConfig{
		Driver: "mysql",
		Dsn:    "root:123456@tcp(127.0.0.1:3306)/db_edge?charset=utf8mb4&timeout=30s",
		Prefix: "edge",
	})
	if err != nil {
		t.Fatal(err)
	}
	/**_, err = db.Exec("SET GLOBAL max_prepared_stmt_count=6000")
	if err != nil {
		t.Fatal(err)
	}**/

	for i := 0; i < 20000; i++ {
		stmt, err := db.Prepare("SELECT " + strconv.Itoa(i))
		if err != nil {
			t.Fatal(err)
		}
		_ = stmt.Close()
		_ = stmt
	}
}

func TestStmtManager_Prepare_Same(t *testing.T) {
	db, err := NewInstanceFromConfig(&DBConfig{
		Driver: "mysql",
		Dsn:    "root:123456@tcp(127.0.0.1:3306)/db_edge?charset=utf8mb4&timeout=30s",
		Prefix: "edge",
	})
	if err != nil {
		t.Fatal(err)
	}
	/**_, err = db.Exec("SET GLOBAL max_prepared_stmt_count=6000")
	if err != nil {
		t.Fatal(err)
	}**/

	for i := 0; i < 20000; i++ {
		stmt, cached, err := db.PrepareOnce("SELECT 1")
		if err != nil {
			t.Fatal(err)
		}
		if !cached {
			_ = stmt.Close()
		}
		_ = stmt
	}

	t.Log(db.StmtManager().Len())
}

func TestStmtManager_Tx_PrepareOnce(t *testing.T) {
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

	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 20000; i++ {
		stmt, cached, err := tx.PrepareOnce("SELECT " + strconv.Itoa(i))
		if err != nil {
			t.Fatal(err)
		}
		if !cached {
			_ = stmt.Close()
		}
	}
	t.Log("stmt map:", tx.db.stmtManager.Len(), "sub map:", len(tx.db.stmtManager.subMap))

	_ = tx.Commit()

	t.Log("stmt map:", tx.db.stmtManager.Len(), "sub map:", len(tx.db.stmtManager.subMap))
}

func TestStmtManager_DB_Close(t *testing.T) {
	db, err := NewInstanceFromConfig(&DBConfig{
		Driver: "mysql",
		Dsn:    "root:123456@tcp(127.0.0.1:3306)/db_edge?charset=utf8mb4&timeout=30s",
		Prefix: "edge",
	})
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 20000; i++ {
		stmt, cached, err := db.PrepareOnce("SELECT " + strconv.Itoa(i))
		if err != nil {
			t.Fatal(err)
		}
		if !cached {
			_ = stmt.Close()
		}
	}
	t.Log("before close: stmt map:", db.stmtManager.Len(), "sub map:", len(db.stmtManager.subMap))
	_ = db.Close()
	t.Log("after close: stmt map:", db.stmtManager.Len(), "sub map:", len(db.stmtManager.subMap))
}

func TestStmtManager_GC(t *testing.T) {
	db, err := NewInstanceFromConfig(&DBConfig{
		Driver: "mysql",
		Dsn:    "root:123456@tcp(127.0.0.1:3306)/db_edge?charset=utf8mb4&timeout=30s",
		Prefix: "edge",
	})
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 10; i++ {
		stmt, cached, err := db.PrepareOnce("SELECT " + strconv.Itoa(i))
		if err != nil {
			t.Fatal(err)
		}
		if !cached {
			_ = stmt.Close()
		}
	}

	runtime.GC()

	time.Sleep(1 * time.Second)
}

func TestUnixTime(t *testing.T) {
	t.Log(unixTime())
	time.Sleep(1 * time.Second)
	t.Log(unixTime())
	time.Sleep(1 * time.Second)
	t.Log(unixTime())
	time.Sleep(1 * time.Second)
	t.Log(unixTime())
}
