package dbs

type SQLPreparer interface {
	// PrepareOnce 可重用的Prepare
	PrepareOnce(query string) (*Stmt, bool, error)

	// Prepare 不可重用的Prepare
	Prepare(query string) (*Stmt, error)
}
