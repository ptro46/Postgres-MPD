package main

import (
	"bufio"
	"database/sql" // package SQL
	"encoding/json"
	"errors"
	"fmt"
	_ "github.com/lib/pq" // driver Postgres
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

const (
	STATUS_OK       = 0
	STATUS_WARNING  = 1
	STATUS_CRITICAL = 2
	STATUS_UNKNOW   = 3
)

func readFile(configFilename string) (string, error) {
	contentOfFile, err := ioutil.ReadFile(configFilename)
	if err != nil {
		return "", err
	}
	return string(contentOfFile), nil
}

func readTableList(db *sql.DB) []string {
	result := make([]string, 0, 0)

	sqlTableList := "SELECT c.relname as Name FROM pg_catalog.pg_class c LEFT JOIN pg_catalog.pg_namespace n ON n.oid = c.relnamespace WHERE c.relkind IN ('r','') AND n.nspname <> 'pg_catalog' AND n.nspname <> 'information_schema' AND n.nspname !~ '^pg_toast' AND pg_catalog.pg_table_is_visible(c.oid) ORDER BY 1"
	rows, err := db.Query(sqlTableList)
	if err != nil {
		return result
	}
	defer rows.Close()

	for rows.Next() {
		var tableName string
		err = rows.Scan(&tableName)
		if err != nil {
			return make([]string, 0, 0)
		}
		result = append(result, tableName)
	}
	return result
}

func getTableNameOID(db *sql.DB, tableName string) (*Table, error) {
	sqlTableOID := "SELECT c.oid,c.relname FROM pg_catalog.pg_class c LEFT JOIN pg_catalog.pg_namespace n ON n.oid = c.relnamespace WHERE c.relname=$1 AND pg_catalog.pg_table_is_visible(c.oid) ORDER BY 2"
	rows, err := db.Query(sqlTableOID, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		var oid string
		var name string
		err = rows.Scan(&oid, &name)
		if err != nil {
			return nil, err
		}
		return NewTable(oid, name), nil
	}
	return nil, errors.New("No Results")
}

func getForeignKeysList(db *sql.DB, table *Table) error {
	result := make([]*ForeignKey, 0, 0)
	sqlColumns := "SELECT conname, pg_catalog.pg_get_constraintdef(r.oid, true) as condef FROM pg_catalog.pg_constraint r WHERE r.conrelid=$1 AND r.contype = 'f' ORDER BY 1"
	rows, err := db.Query(sqlColumns, table.Oid)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var conname string
		var condef string
		err = rows.Scan(&conname, &condef)
		if err != nil {
			return err
		}
		fk := NewForeignKey(conname, condef)
		result = append(result, fk)
	}
	table.foreignKeys = result
	return nil
}

func getColumnsList(db *sql.DB, table *Table) error {
	result := make([]*Column, 0, 0)
	sqlColumns := "SELECT a.attname, pg_catalog.format_type(a.atttypid, a.atttypmod), a.attnotnull FROM pg_catalog.pg_attribute a WHERE a.attrelid=$1 AND a.attnum > 0 AND NOT a.attisdropped ORDER BY a.attnum"
	rows, err := db.Query(sqlColumns, table.Oid)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var columnName string
		var formatType string
		var isnullable bool
		err = rows.Scan(&columnName, &formatType, &isnullable)
		if err != nil {
			return err
		}
		column := NewColumn(columnName, formatType, isnullable)
		result = append(result, column)
	}
	table.columns = result
	return nil
}

func getPrimaryKeyConstraint(db *sql.DB, table *Table) error {
	sqlPrimaryKey := "SELECT c2.relname, pg_catalog.pg_get_constraintdef(con.oid, true) FROM pg_catalog.pg_class c, pg_catalog.pg_class c2, pg_catalog.pg_index i LEFT JOIN pg_catalog.pg_constraint con ON (conrelid = i.indrelid AND conindid = i.indexrelid AND contype IN ('p','u','x')) WHERE c.oid=$1 AND c.oid = i.indrelid AND i.indexrelid = c2.oid ORDER BY i.indisprimary DESC, i.indisunique DESC, c2.relname"
	rows, err := db.Query(sqlPrimaryKey, table.Oid)
	if err != nil {
		return err
	}
	defer rows.Close()

	if rows.Next() {
		var relname string
		var constraint string
		err = rows.Scan(&relname, &constraint)
		if err != nil {
			return err
		}
		table.PrimaryKeyName = relname
		table.PrimaryKeyConstraint = constraint
		return nil
	}
	return errors.New("No PrimaryKey")
}

func descTable(logWriter *bufio.Writer, db *sql.DB, tableName string) (*Table, error) {
	table, err := getTableNameOID(db, tableName)
	if err == nil {
		fmt.Fprintf(logWriter, "%s::%s ", table.Oid, table.Name)
		err := getColumnsList(db, table)
		if err != nil {
			return table, err
		}
		fmt.Fprintf(logWriter, " columns::%d ", len(table.columns))

		err = getPrimaryKeyConstraint(db, table)
		if err != nil {
			fmt.Fprintf(logWriter, " NoPrimaryKey ")
		} else {
			fmt.Fprintf(logWriter, " %s::%s ", table.PrimaryKeyName, table.PrimaryKeyConstraint)
		}

		err = getForeignKeysList(db, table)
		if err != nil {
			fmt.Fprintf(logWriter, " NoForeignKeys\n")
		} else {
			fmt.Fprintf(logWriter, " foreignKeys::%d\n", len(table.foreignKeys))
		}

	}
	return table, nil
}

func loadMPDConfig(configFileName string) ([]*Cluster, []*Hidden, *OutputConfig, error) {
	clusters := make([]*Cluster, 0, 0)
	hiddens := make([]*Hidden, 0, 0)

	configContent, err := readFile(configFileName)
	if err == nil {
		var parsed map[string]interface{}
		errUnmarshal := json.Unmarshal([]byte(configContent), &parsed)
		if errUnmarshal == nil {
			outputConfigInterface := parsed["outputConfig"].(map[string]interface{})
			outputName := outputConfigInterface["outputName"].(string)
			dotOutputFormat := outputConfigInterface["dotOutputFormat"].(string)
			dotBinary := outputConfigInterface["dotBinary"].(string)
			includePrimaryKeys := outputConfigInterface["includePrimaryKeys"].(bool)
			includeForeignKeys := outputConfigInterface["includeForeignKeys"].(bool)
			includeDataColumns := outputConfigInterface["includeDataColumns"].(bool)
			outputConfig := NewOutputConfig(outputName, dotOutputFormat, dotBinary, includePrimaryKeys, includeForeignKeys, includeDataColumns)

			arrayOfClusters := parsed["clusters"].([]interface{})
			for _, oneCluster := range arrayOfClusters {
				oneClusterMap := oneCluster.(map[string]interface{})
				clusterName := oneClusterMap["name"].(string)
				clusterColor := oneClusterMap["background"].(string)
				clusterShow := oneClusterMap["show"].(bool)
				cluster := NewCluster(clusterName, clusterColor)
				cluster.Show = clusterShow

				content := make([]string, 0, 0)
				arrayOfTablesNames := oneClusterMap["content"].([]interface{})
				for _, tableNameInterface := range arrayOfTablesNames {
					tableName := tableNameInterface.(string)
					content = append(content, tableName)
				}
				cluster.content = content
				clusters = append(clusters, cluster)
			}

			arrayOfHiddens := parsed["hidden"].([]interface{})
			for _, oneHidden := range arrayOfHiddens {
				oneHiddenMap := oneHidden.(map[string]interface{})
				tableName := oneHiddenMap["tableName"].(string)
				hidden := NewHidden(tableName)

				columns := make([]string, 0, 0)
				arrayOfColumnsNames := oneHiddenMap["columns"].([]interface{})
				for _, columnNameInterface := range arrayOfColumnsNames {
					columnName := columnNameInterface.(string)
					columns = append(columns, columnName)
				}
				hidden.columns = columns
				hiddens = append(hiddens, hidden)
			}

			return clusters, hiddens, outputConfig, nil

		} else {
			return nil, nil, nil, errUnmarshal
		}
	} else {
		return nil, nil, nil, err
	}
}

func findByName(tables []*Table, tableName string) *Table {
	for _, table := range tables {
		if table.Name == tableName {
			return table
		}
	}
	return nil
}

func trimSuffix(s, suffix string) string {
	if strings.HasSuffix(s, suffix) {
		s = s[:len(s)-len(suffix)]
	}
	return s
}

func main() {
	//testRegex()
	message := "UNKNOW - "
	status := STATUS_UNKNOW

	workingDirectory := "./"
	configFileName := "mpd.config"
	if len(os.Args) == 3 {
		workingDirectory = os.Args[1]
		configFileName = os.Args[2]
	}
	workingDirectory = trimSuffix(workingDirectory, "/")
	postgresConfigFullFileName := "./" + "postgres-mpd.config"
	configFullFileName := workingDirectory + "/" + configFileName
	logFullFileName := workingDirectory + "/postgres-mpd.log"
	logFileHandle, err := os.Create(logFullFileName)
	if err != nil {
		fmt.Println(err)
		message = fmt.Sprintf("CRITICAL - can not create log file %s", logFullFileName)
		status = STATUS_CRITICAL
		fmt.Printf("%s\n", message)
		os.Exit(status)
	}
	defer logFileHandle.Close()
	logWriter := bufio.NewWriter(logFileHandle) // *bufio.Writer

	dbinfo, err := readFile(postgresConfigFullFileName)
	if err == nil {
		postgresMPDConfig := new(PostgresMPDConfig)
		errUnmarshal := json.Unmarshal([]byte(dbinfo), postgresMPDConfig)
		if errUnmarshal == nil {
			db, err := sql.Open("postgres", postgresMPDConfig.dbPostgresConnectString())
			if err == nil {
				err = db.Ping()
				if err == nil {
					fmt.Fprintln(logWriter, "Connected to "+postgresMPDConfig.Db)
					fmt.Println("Connected to " + postgresMPDConfig.Db)
					defer db.Close()

					clusters, hiddensTables, outputConfig, err := loadMPDConfig(configFullFileName)
					if err != nil {
						fmt.Fprintln(logWriter, err)
						message = fmt.Sprintf("CRITICAL - can not load config file %s", configFullFileName)
						status = STATUS_CRITICAL
						fmt.Fprintf(logWriter, "%s\n", message)
						fmt.Printf("%s\n", message)
						os.Exit(status)
					}
					if err == nil {
						fmt.Fprintln(logWriter, "Describe tables")
						fmt.Println("Describe tables")
						tableList := readTableList(db)
						tables := make([]*Table, 0, 0)
						for _, tableName := range tableList {
							table, err := descTable(logWriter, db, tableName)
							if err != nil {
								fmt.Fprintln(logWriter, err)
								fmt.Println(err)
							}
							if nil != table {
								tables = append(tables, table)
							}
						}

						fmt.Fprintln(logWriter, "Build Hidden Skel")
						fmt.Println("Build Hidden Skel")
						hiddenFullFileName := workingDirectory + "/" + postgresMPDConfig.Db + "-hidden.json"
						hiddenFile, err := os.Create(hiddenFullFileName)
						if err == nil {
							defer hiddenFile.Close()
							hiddenWriter := bufio.NewWriter(hiddenFile) // *bufio.Writer
							hiddenWriter.WriteString("	\"hidden\":[\n")
							for _, table := range tables {
								table.generateHidden(hiddenWriter)
							}
							hiddenWriter.WriteString("	]\n")
							hiddenWriter.Flush()
						}

						fmt.Println("Build Clusters")
						fmt.Fprintln(logWriter, "Build Clusters")
						for _, cluster := range clusters {
							fmt.Fprintln(logWriter, "    Build cluster "+cluster.Name)
							clusterTables := make([]*Table, 0, 0)
							for _, tableName := range cluster.content {
								table := findByName(tables, tableName)
								if nil != table {
									clusterTables = append(clusterTables, table)
								} else {
									fmt.Fprintln(logWriter, "        Can not find table "+tableName)
								}
							}
							cluster.tables = clusterTables
							fmt.Fprintf(logWriter, "        %d tables\n", len(cluster.tables))
						}

						for _, cluster := range clusters {
							cluster.generatePrimaryKeyConstraint()
							cluster.generateForeignKeysConstraint(logWriter)
						}

						fmt.Println("Generate Clusters and Tables DOT")
						fmt.Fprintln(logWriter, "Generate Clusters and Tables DOT")
						dotFullFileName := workingDirectory + "/" + outputConfig.OutputName + ".dot"
						dotFile, err := os.Create(dotFullFileName)
						if err == nil {
							defer dotFile.Close()
							dotWriter := bufio.NewWriter(dotFile) // *bufio.Writer
							dotWriter.WriteString("digraph \"" + postgresMPDConfig.Db + "\" {\n")
							for _, cluster := range clusters {
								fmt.Fprintln(logWriter, "    Generate cluster "+cluster.Name)
								err := cluster.generateDot(logWriter, dotWriter, hiddensTables, outputConfig)
								if nil != err {
									message = fmt.Sprintf("CRITICAL - can not generate DOT file %s.dot %s", postgresMPDConfig.Db, err.Error())
									status = STATUS_CRITICAL

								}
							}

							fmt.Println("Generate tables relations ")
							fmt.Fprintln(logWriter, "    Generate relations ")
							if status == STATUS_UNKNOW {
								for _, table := range tables {
									table.generateDotRelations(dotWriter, tables)
								}
							}
							dotWriter.WriteString("}\n")
							dotWriter.Flush()

							dotOutputFullFileName := workingDirectory + "/" + outputConfig.OutputName

							/*
								cmd := fmt.Sprintf("%s", outputConfig.DotBinary)
								arg1 := fmt.Sprintf("-T%s", outputConfig.DotOutputFormat)
								arg2 := fmt.Sprintf("-o%s.%s", dotOutputFullFileName, outputConfig.DotOutputFormat)
								arg3 := fmt.Sprintf("%s", dotFullFileName)
								fmt.Fprintln(logWriter, cmd+" "+arg1+" "+arg2+" "+arg3)
								fmt.Println(cmd + " " + arg1 + " " + arg2 + " " + arg3)
								execCmd := exec.Command(cmd, arg1, arg2, arg3)
								err := execCmd.Run()
							*/
							cmd := fmt.Sprintf("%s -T%s -o%s.%s %s", outputConfig.DotBinary, outputConfig.DotOutputFormat, dotOutputFullFileName, outputConfig.DotOutputFormat, dotFullFileName)
							fmt.Fprintln(logWriter, cmd)
							fmt.Println(cmd)
							_, err := exec.Command("bash", "-c", cmd).Output()
							if err != nil {
								message = fmt.Sprintf("CRITICAL - can not generate output file %s.%s %s", dotOutputFullFileName, outputConfig.DotOutputFormat, err.Error())
								status = STATUS_CRITICAL

							} else {
								message = fmt.Sprintf("OK")
								status = STATUS_OK

							}
							logWriter.Flush()
						} else {
							message = fmt.Sprintf("CRITICAL - can not create file %s.dot %s", postgresMPDConfig.Db, err.Error())
							status = STATUS_CRITICAL
						}

					} else {
						message = fmt.Sprintf("CRITICAL - can not load config file %s %s", configFullFileName, err.Error())
						status = STATUS_CRITICAL
					}
				} else {
					message = fmt.Sprintf("CRITICAL - can not connect to postgres database %s", err.Error())
					status = STATUS_CRITICAL
				}
			} else {
				message = fmt.Sprintf("CRITICAL - can not connect to postgres database %s", err.Error())
				status = STATUS_CRITICAL
			}
		} else {
			message = fmt.Sprintf("CRITICAL - can not parse json config file %s", errUnmarshal.Error())
			status = STATUS_CRITICAL
		}
	} else {
		message = fmt.Sprintf("CRITICAL - can not read postgres config file : %s", postgresConfigFullFileName)
		status = STATUS_CRITICAL
	}
	fmt.Fprintf(logWriter, "%s\n", message)
	fmt.Printf("%s\n", message)
	os.Exit(status)
}
