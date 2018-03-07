package main

import (
	"bufio"
	"fmt"
	"regexp"
	"strings"
)

type Table struct {
	Oid                  string
	Name                 string
	IsVisible            bool
	columns              []*Column
	PrimaryKeyName       string
	PrimaryKeyConstraint string
	foreignKeys          []*ForeignKey
}

func NewTable(oid_ string, name_ string) *Table {
	return &Table{Oid: oid_, Name: name_, IsVisible: false}
}

func findHiddenByTableName(hiddenTableColumns []*Hidden, tableName string) *Hidden {
	for _, hidden := range hiddenTableColumns {
		if hidden.TableName == tableName {
			return hidden
		}
	}
	return nil
}

func isHiddenColumn(columns []string, columnName string) bool {
	for _, column := range columns {
		if column == columnName {
			return true
		}
	}
	return false
}

func (table *Table) generatePrimaryKeyConstraint() {
	s := strings.Replace(table.PrimaryKeyConstraint, "(", "'", -1)
	s = strings.Replace(s, ")", "'", -1)
	re := regexp.MustCompile("PRIMARY KEY '([a-zA-Z0-9_-]+)'")
	matchArray := re.FindSubmatch([]byte(s))
	if len(matchArray) == 2 {
		primaryKeyColumn := string(matchArray[1])
		for _, column := range table.columns {
			if column.Name == primaryKeyColumn {
				column.IsPrimary = true
			}
		}
	}
}

func (table *Table) generateForeignKeysConstraint(logWriter *bufio.Writer) {
	for _, foreignKey := range table.foreignKeys {
		err := foreignKey.parse()
		if err == nil {
			for _, column := range table.columns {
				if column.Name == foreignKey.ColumnName {
					column.IsForeign = true
				}
			}
		} else {
			fmt.Fprintln(logWriter, err)
		}
	}
}

func (table *Table) generateDot(logWriter *bufio.Writer, dotWriter *bufio.Writer, hiddenTableColumns []*Hidden, outputConfig *OutputConfig) error {
	table.IsVisible = true
	sTableDotFragment := fmt.Sprintf("    \"%s\" [fontname=\"Helvetica\" color=\"violetred4\" peripheries=\"1\"  fontsize=\"10\" label=<<TABLE BORDER=\"0\" bgcolor=\"grey80\"><TR><TD COLSPAN=\"2\" bgcolor=\"grey52\">%s</TD></TR>", table.Name, table.Name)
	dotWriter.WriteString(sTableDotFragment)
	hidden := findHiddenByTableName(hiddenTableColumns, table.Name)
	for _, column := range table.columns {
		if hidden == nil {
			if column.IsPrimary == true && outputConfig.IncludePrimaryKeys == true {
				column.generateDot(dotWriter, hiddenTableColumns, outputConfig)

			} else if column.IsForeign == true && outputConfig.IncludeForeignKeys == true {
				column.generateDot(dotWriter, hiddenTableColumns, outputConfig)

			} else if column.IsPrimary == false && column.IsForeign == false && outputConfig.IncludeDataColumns == true {
				column.generateDot(dotWriter, hiddenTableColumns, outputConfig)

			} else {
				fmt.Fprintf(logWriter, "Bypass %s::%s ", table.Name, column.Name)
			}
		} else {
			if false == isHiddenColumn(hidden.columns, column.Name) {
				if column.IsPrimary == true && outputConfig.IncludePrimaryKeys == true {
					column.generateDot(dotWriter, hiddenTableColumns, outputConfig)

				} else if column.IsForeign == true && outputConfig.IncludeForeignKeys == true {
					column.generateDot(dotWriter, hiddenTableColumns, outputConfig)

				} else if column.IsPrimary == false && column.IsForeign == false && outputConfig.IncludeDataColumns == true {
					column.generateDot(dotWriter, hiddenTableColumns, outputConfig)

				} else {
					fmt.Fprintf(logWriter, "Bypass %s::%s ", table.Name, column.Name)
				}
			} else {
				fmt.Fprintf(logWriter, "Bypass %s::%s ", table.Name, column.Name)
			}
		}
	}
	dotWriter.WriteString("</TABLE>>];\n")
	return nil
}

func (table *Table) generateDotRelations(dotWriter *bufio.Writer, tables []*Table) error {
	dotWriter.WriteString("\n")
	for _, foreignKey := range table.foreignKeys {
		tableForeign := findByName(tables, foreignKey.RefTable)
		if table.IsVisible == true && tableForeign != nil && tableForeign.IsVisible == true {
			sRelFragment := fmt.Sprintf("  \"%s\" -> \"%s\" [color=\"blue\" fontsize=\"6\" labelcolor=\"blue\" fontcolor=\"blue\" arrowhead=\"normal\"   label=\"%s\\n%s = %s\"]\n", table.Name, foreignKey.RefTable, foreignKey.Name, foreignKey.ColumnName, foreignKey.RefColumn)
			dotWriter.WriteString(sRelFragment)
		}
	}
	dotWriter.WriteString("\n")
	return nil
}

func (table *Table) generateHidden(hiddenWriter *bufio.Writer) {
	hiddenWriter.WriteString("		{\n")
	hiddenWriter.WriteString("			\"tableName\" : \"" + table.Name + "\",\n")
	hiddenWriter.WriteString("			\"columns\" : [")
	for colIndex, column := range table.columns {
		if colIndex < len(table.columns)-1 {
			hiddenWriter.WriteString("\"" + column.Name + "\",")
		} else {
			hiddenWriter.WriteString("\"" + column.Name + "\"")
		}
	}
	hiddenWriter.WriteString("]\n")
	hiddenWriter.WriteString("		},\n")
}
