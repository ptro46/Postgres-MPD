package main

type Hidden struct {
	TableName string
	columns   []string
}

func NewHidden(tablename_ string) *Hidden {
	return &Hidden{TableName: tablename_}
}
