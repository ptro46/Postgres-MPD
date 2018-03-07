package main

import (
	"bufio"
	"fmt"
)

type Cluster struct {
	Name       string
	Background string
	Show       bool
	content    []string
	tables     []*Table
}

func NewCluster(name_ string, background_ string) *Cluster {
	return &Cluster{Name: name_, Background: background_}
}

func (cluster *Cluster) generatePrimaryKeyConstraint() {
	for _, table := range cluster.tables {
		table.generatePrimaryKeyConstraint()
	}
}

func (cluster *Cluster) generateForeignKeysConstraint(logWriter *bufio.Writer) {
	for _, table := range cluster.tables {
		table.generateForeignKeysConstraint(logWriter)
	}
}

func (cluster *Cluster) generateDot(logWriter *bufio.Writer, dotWriter *bufio.Writer, hiddenTableColumns []*Hidden, outputConfig *OutputConfig) error {
	if cluster.Show == true {
		dotWriter.WriteString("  subgraph \"cluster_" + cluster.Name + "\" {\n")
		dotWriter.WriteString("    node [shape=plaintext];\n")
		dotWriter.WriteString(fmt.Sprintf("    label=\"%s\";\n", cluster.Name))
		dotWriter.WriteString("    style=filled;\n")
		dotWriter.WriteString(fmt.Sprintf("    color=\"%s\";\n", cluster.Background))
		for _, table := range cluster.tables {
			table.generateDot(logWriter, dotWriter, hiddenTableColumns, outputConfig)
		}
		dotWriter.WriteString("  }\n")
	}
	return nil
}
