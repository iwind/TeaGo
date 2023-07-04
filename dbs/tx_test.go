// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package dbs

import (
	"github.com/iwind/TeaGo/types"
	"testing"
)

func TestTx_id(t *testing.T) {
	for i := 0; i < 10; i++ {
		var tx = NewTx(nil, nil)
		t.Log(tx.id)
	}
}

func TestTx_PrepareOnce(t *testing.T) {
	db, err := Default()
	if err != nil {
		t.Fatal(err)
	}

	err = db.RunTx(func(tx *Tx) error {
		t.Log(len(tx.db.stmtManager.stmtMap), "stmts")

		stmt, _, err := tx.PrepareOnce("UPDATE lockers SET version=version+1")
		if err != nil {
			return err
		}
		_, err = stmt.Exec()

		t.Log(len(tx.db.stmtManager.stmtMap), "stmts")

		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ok")

	t.Log(len(db.stmtManager.stmtMap), "stmts")
}

func TestTx_FindOnes(t *testing.T) {
	db, err := Default()
	if err != nil {
		t.Fatal(err)
	}

	err = db.RunTx(func(tx *Tx) error {
		ones, columns, err := tx.FindOnes("SELECT * FROM users WHERE id=?", 1)
		if err != nil {
			return err
		}

		t.Log(ones)
		t.Log(columns)

		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func BenchmarkTx_PrepareOnce(b *testing.B) {
	db, err := Default()
	if err != nil {
		b.Fatal(err)
	}

	_, _ = db.FindCol(0, "SELECT VERSION()")

	// prepare many statements
	for i := 0; i < 1_000; i++ {
		_, _, err = db.PrepareOnce("SELECT " + types.String(i))
		if err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			err = db.RunTx(func(tx *Tx) error {
				stmt, _, err := tx.PrepareOnce("UPDATE lockers SET version=version+1 WHERE `key`=?")
				if err != nil {
					return err
				}
				_, err = stmt.Exec("test_key")
				return err
			})
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkTx_Exec(b *testing.B) {
	db, err := Default()
	if err != nil {
		b.Fatal(err)
	}

	_, _ = db.FindCol(0, "SELECT VERSION()")

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			err = db.RunTx(func(tx *Tx) error {
				_, err = tx.Exec("UPDATE lockers SET version=version+1 WHERE `key`=?", "test_key")
				return err
			})
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
