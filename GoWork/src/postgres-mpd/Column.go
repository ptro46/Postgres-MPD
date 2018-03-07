package main

import (
	"bufio"
	"fmt"
)

type Column struct {
	Name       string
	Type       string
	IsNullable bool
	IsPrimary  bool
	IsForeign  bool
}

func NewColumn(name_ string, type_ string, isnullable_ bool) *Column {
	return &Column{Name: name_, Type: type_, IsNullable: isnullable_, IsPrimary: false, IsForeign: false}
}

func (column *Column) generateDot(dotWriter *bufio.Writer, hiddenTableColumns []*Hidden, outputConfig *OutputConfig) error {
	sColumnDotFragment := fmt.Sprintf("<TR><TD align=\"left\">%s</TD><TD align=\"right\">%s</TD></TR>", column.Name, column.Type)

	dotWriter.WriteString(sColumnDotFragment)

	return nil
}
