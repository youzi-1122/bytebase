package ast

// RenameIndexStmt is the struct for the rename index statement.
type RenameIndexStmt struct {
	node

	// For PostgreSQL, we only use TableDef.Schema.
	// If this rename index statement doesn't contain schema name, Table will be nil.
	Table     *TableDef
	IndexName string
	NewName   string
}
