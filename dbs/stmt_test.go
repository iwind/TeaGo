package dbs

import (
	"github.com/iwind/TeaGo/logs"
	"testing"
)

func TestStmt_FindOnes(t *testing.T) {
	db, err := Default()
	if err != nil {
		t.Fatal(err)
	}
	stmt, err := db.Prepare("SELECT * FROM users")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = stmt.Close()
	}()

	ones, columnNames, err := stmt.FindOnes()
	if err != nil {
		t.Fatal(err)
	}

	t.Log("columnNames:", columnNames)
	logs.PrintAsJSON(ones, t)
}

func TestStmt_FindOnes_Empty(t *testing.T) {
	db, err := Default()
	if err != nil {
		t.Fatal(err)
	}
	stmt, err := db.Prepare("SELECT * FROM users WHERE id=100000")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = stmt.Close()
	}()

	ones, columnNames, err := stmt.FindOnes()
	if err != nil {
		t.Fatal(err)
	}

	t.Log("columnNames:", columnNames)
	logs.PrintAsJSON(ones, t)
}

func TestStmt_FindOnes_Error(t *testing.T) {
	db, err := Default()
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Prepare("SELECT * FROM users WHERE id1=123")
	if err != nil {
		t.Log("expected error:", err)
		return
	}
}

func TestStmt_FindOnes_FUNCTION_Error(t *testing.T) {
	db, err := Default()
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Prepare("SELECT * FROM users WHERE LENGTH()")
	if err != nil {
		t.Log("expected error 1:", err)
		return
	}
}

func TestStmt_FindOne(t *testing.T) {
	db, err := Default()
	if err != nil {
		t.Fatal(err)
	}
	stmt, err := db.Prepare("SELECT * FROM users")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = stmt.Close()
	}()

	one, err := stmt.FindOne()
	if err != nil {
		t.Fatal(err)
	}

	logs.PrintAsJSON(one, t)
}

func TestStmt_FindOne_Limit(t *testing.T) {
	db, err := Default()
	if err != nil {
		t.Fatal(err)
	}
	stmt, err := db.Prepare("SELECT * FROM users LIMIT 1")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = stmt.Close()
	}()

	one, err := stmt.FindOne()
	if err != nil {
		t.Fatal(err)
	}

	logs.PrintAsJSON(one, t)
}

func TestStmt_FindOne_Empty(t *testing.T) {
	db, err := Default()
	if err != nil {
		t.Fatal(err)
	}
	stmt, err := db.Prepare("SELECT * FROM users WHERE id=11111111")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = stmt.Close()
	}()

	one, err := stmt.FindOne()
	if err != nil {
		t.Fatal(err)
	}

	logs.PrintAsJSON(one, t)
}

func TestStmt_FindCol(t *testing.T) {
	db, err := Default()
	if err != nil {
		t.Fatal(err)
	}
	stmt, err := db.Prepare("SELECT id, name FROM users")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = stmt.Close()
	}()

	{
		col, err := stmt.FindCol(0)
		if err != nil {
			t.Fatal(err)
		}
		t.Log("col0:", col)
	}

	{
		col, err := stmt.FindCol(1)
		if err != nil {
			t.Fatal(err)
		}
		t.Log("col1:", col)
	}
}

func TestStmt_FindCol_Empty(t *testing.T) {
	db, err := Default()
	if err != nil {
		t.Fatal(err)
	}
	stmt, err := db.Prepare("SELECT id, name FROM users WHERE id=111111111")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = stmt.Close()
	}()

	{
		col, err := stmt.FindCol(0)
		if err != nil {
			t.Fatal(err)
		}
		t.Log("col0:", col)
	}

	{
		col, err := stmt.FindCol(1)
		if err != nil {
			t.Fatal(err)
		}
		t.Log("col1:", col)
	}
}

func TestStmt_FindCol_Overflow(t *testing.T) {
	db, err := Default()
	if err != nil {
		t.Fatal(err)
	}
	stmt, err := db.Prepare("SELECT id, name FROM users")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = stmt.Close()
	}()

	{
		col, err := stmt.FindCol(-1)
		if err != nil {
			t.Log("expected error:", err)
		}
		t.Log("col0:", col)
	}

	{
		col, err := stmt.FindCol(2)
		if err != nil {
			t.Log("expected error:", err)
		}
		t.Log("col0:", col)
	}
}

func BenchmarkStmt_FindOnes(b *testing.B) {
	db, err := Default()
	if err != nil {
		b.Fatal(err)
	}
	stmt, err := db.Prepare("SELECT id, name FROM users LIMIT 1")
	if err != nil {
		b.Fatal(err)
	}
	defer func() {
		_ = stmt.Close()
	}()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _, _ = stmt.FindOnes()
	}
}

func BenchmarkStmt_FindOne(b *testing.B) {
	db, err := Default()
	if err != nil {
		b.Fatal(err)
	}
	stmt, err := db.Prepare("SELECT id, name FROM users LIMIT 1")
	if err != nil {
		b.Fatal(err)
	}
	defer func() {
		_ = stmt.Close()
	}()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = stmt.FindOne()
	}
}
