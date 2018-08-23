package dbs

type Table struct {
	Name      string
	Schema    string
	Fields    []*Field
	Comment   string
	Collation string
	Engine    string
	Code      string // DDL SQL

	Partitions []*TablePartition // 分区
	Indexes    []*TableIndex     // 索引

	MappingName string // 映射的模型名
}

func (this *Table) FindFieldWithName(name string) *Field {
	for _, field := range this.Fields {
		if field.Name == name {
			return field
		}
	}
	return nil
}

func (this *Table) FindPartitionWithName(name string) *TablePartition {
	for _, partition := range this.Partitions {
		if partition.Name == name {
			return partition
		}
	}
	return nil
}

func (this *Table) FindIndexWithName(name string) *TableIndex {
	for _, index := range this.Indexes {
		if index.Name == name {
			return index
		}
	}
	return nil
}
