package dbs

type TablePartition struct {
	Name            string
	Method          string
	Expression      string
	Description     string
	OrdinalPosition int
	NodeGroup       string
	Rows            int64
}
