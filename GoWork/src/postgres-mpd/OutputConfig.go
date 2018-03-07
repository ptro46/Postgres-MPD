package main

type OutputConfig struct {
	OutputName         string `json:"outputName,omitempty"`
	DotOutputFormat    string `dotOutputFormat:"password,omitempty"`
	DotBinary          string `json:"dotBinary,omitempty"`
	IncludePrimaryKeys bool   `json:"includePrimaryKeys,omitempty"`
	IncludeForeignKeys bool   `json:"includeForeignKeys,omitempty"`
	IncludeDataColumns bool   `json:"includeDataColumns,omitempty"`
}

func NewOutputConfig(outputName string, outputFormat string, dotBinary string, incPK bool, incFK bool, incDATA bool) *OutputConfig {
	return &OutputConfig{
		OutputName:         outputName,
		DotOutputFormat:    outputFormat,
		DotBinary:          dotBinary,
		IncludePrimaryKeys: incPK,
		IncludeForeignKeys: incFK,
		IncludeDataColumns: incDATA}
}
