package dbs

type SQLPreparer interface {
	// 可重用的Prepare
	PrepareOnce(query string) (*Stmt, error)

	// 不可重用的Prepare
	Prepare(query string) (*Stmt, error)
}
