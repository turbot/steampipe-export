package main

import (
	"context"
	encoding_csv "encoding/csv"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/hashicorp/go-hclog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/turbot/steampipe-plugin-sdk/v5/anywhere"
	filter2 "github.com/turbot/steampipe-plugin-sdk/v5/filter"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/logging"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/sperr"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var pluginServer *grpc.PluginServer
var pluginAlias = ""
var connection = pluginAlias

type displayRowFunc func(row *proto.ExecuteResponse, columns []string)

func main() {
	setupLogger(pluginAlias)
	rootCmd := &cobra.Command{
		Use:   "steampipe_export TABLE_NAME [flags]",
		Short: "Steampipe data export tool",
		Run:   executeCommand,
		Args:  cobra.ExactArgs(1),
	}

	// Define flags
	rootCmd.PersistentFlags().String("config", "", "Config file data")
	rootCmd.PersistentFlags().String("where", "", "where clause data")
	rootCmd.PersistentFlags().StringSlice("select", nil, "Column data to display")
	rootCmd.PersistentFlags().Int("limit", 0, "Limit data")

	viper.BindPFlags(rootCmd.PersistentFlags())

	pluginServer = plugin.Server(&plugin.ServeOpts{})

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}

func executeCommand(cmd *cobra.Command, args []string) {
	// TODO template

	table := args[0]
	if err := setConnectionConfig(); err != nil {
		// TODO display error
		fmt.Println(err)
		os.Exit((1))
	}

	schema, err := getSchema(table)
	if err != nil {
		// TODO display error
		fmt.Println(err)
		os.Exit((1))
	}

	columns, err := getColumns(schema)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var qual map[string]*proto.Quals
	if viper.GetString("where") != "" {
		whereFlag := viper.GetString("where")
		qual, err = filterStringToQuals(whereFlag, schema)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	executeQuery(table, connection, columns, qual, displayCSVRow)
}

func getColumns(schema *proto.TableSchema) ([]string, error) {
	var columns = viper.GetStringSlice("select")
	if len(columns) != 0 {
		tableColumn := schema.GetColumnNames()
		for _, item := range columns {
			if !slices.Contains(tableColumn, item) {
				return nil, fmt.Errorf("column %s does not exist", item)
			}
		}
	}
	if len(columns) == 0 {
		columns = schema.GetColumnNames()
	}
	sort.Strings(columns)
	return columns, nil
}

func getSchema(table string) (*proto.TableSchema, error) {
	req := &proto.GetSchemaRequest{
		Connection: connection,
	}
	pluginSchema, err := pluginServer.GetSchema(req)
	if err != nil {
		return nil, err
	}
	return pluginSchema.Schema.Schema[table], nil
}

func setConnectionConfig() error {
	pluginName := NewSteampipeImageRef(pluginAlias).DisplayImageRef()

	connectionConfig := &proto.ConnectionConfig{
		Connection:      connection,
		Plugin:          pluginName,
		PluginShortName: pluginAlias,
		Config:          viper.GetString("config"),
		PluginInstance:  pluginName,
	}

	configs := []*proto.ConnectionConfig{connectionConfig}
	req := &proto.SetAllConnectionConfigsRequest{
		Configs: configs,
	}

	_, err := pluginServer.SetAllConnectionConfigs(req)

	if err != nil {
		return err
	}
	return nil
}

func executeQuery(tableName string, conectionName string, columns []string, qual map[string]*proto.Quals, displayRow displayRowFunc) {
	// construct execute request

	var qualMap = map[string]*proto.Quals{}

	if qual != nil {
		qualMap = qual
	}

	var limit int64 = -1

	if viper.GetInt("limit") != 0 {
		limit = int64(viper.GetInt("limit"))
	}

	queryContext := proto.NewQueryContext(columns, qualMap, limit)
	req := &proto.ExecuteRequest{
		Table:                 tableName,
		QueryContext:          queryContext,
		CallId:                grpc.BuildCallId(),
		Connection:            conectionName,
		TraceContext:          nil,
		ExecuteConnectionData: make(map[string]*proto.ExecuteConnectionData),
	}
	req.ExecuteConnectionData = map[string]*proto.ExecuteConnectionData{
		req.Connection: {
			Limit:        req.QueryContext.Limit,
			CacheEnabled: false,
		},
	}
	ctx := context.Background()
	stream := anywhere.NewLocalPluginStream(ctx)
	pluginServer.CallExecuteAsync(req, stream)
	for {

		response, err := stream.Recv()
		if err != nil {
			fmt.Printf("[ERROR] Error receiving data from the channel: %v", err)
			break
		}
		if response == nil {
			break
		}
		displayRow(response, columns)
	}
}

var rowCount = 0

func displayCSVRow(displayRow *proto.ExecuteResponse, columns []string) {
	row := displayRow.Row
	selectColumns := viper.GetStringSlice("select")

	// Process each column and store values in a map
	res := make(map[string]string, len(row.Columns))
	for columnName, column := range row.Columns {
		var val interface{}
		if bytes := column.GetJsonValue(); bytes != nil {
			val = string(bytes)
		} else if timestamp := column.GetTimestampValue(); timestamp != nil {
			val = ptypes.TimestampString(timestamp)
		} else {
			column.ProtoReflect().Range(func(descriptor protoreflect.FieldDescriptor, v protoreflect.Value) bool {
				if descriptor.JSONName() == "nullValue" {
					val = nil
				} else {
					val = v.Interface()
				}
				return false
			})
		}
		res[columnName] = fmt.Sprintf("%v", val)
	}

	// Prepare CSV writer
	writer := encoding_csv.NewWriter(os.Stdout)
	defer writer.Flush()

	// Write headers
	if rowCount == 0 {
		if len(selectColumns) > 0 {
			// Write headers based on selectColumns
			writer.Write(selectColumns)
		} else {
			// Write all headers
			writer.Write(columns)
		}
		writer.Flush()

		if err := writer.Error(); err != nil {
			fmt.Println(err)
		}
	}

	rowCount++

	// Generate row data
	var colVals []string
	if len(selectColumns) > 0 {
		colVals = make([]string, len(selectColumns))
		for i, columnName := range selectColumns {
			colVals[i], _ = res[columnName] // Using _ to ignore whether columnName is present in res
		}
	} else {
		colVals = make([]string, len(columns))
		for i, columnName := range columns {
			colVals[i], _ = res[columnName]
		}
	}

	// Write the row data
	writer.Write(colVals)
	writer.Flush()

	// Handle potential errors from the writer
	if err := writer.Error(); err != nil {
		fmt.Println(err)
	}
}

func filterStringToQuals(raw string, tableSchema *proto.TableSchema) (map[string]*proto.Quals, error) {
	columnMap := tableSchema.GetColumnMap()
	keyColumns := tableSchema.GetAllKeyColumns()

	parsed, err := filter2.Parse("", []byte(raw))
	if err != nil {
		log.Printf("err %v", err)
		return nil, sperr.New("failed to parse 'where' property: %s", err.Error())
	}

	// convert table schema into a column map

	filter := parsed.(filter2.ComparisonNode)
	log.Println(filter)
	var qual *proto.Qual
	var column string

	switch filter.Type {

	case "compare", "like":
		codeNodes, ok := filter.Values.([]filter2.CodeNode)
		if !ok {
			return nil, fmt.Errorf("failed to parse filter")
		}
		if len(codeNodes) != 2 {
			return nil, fmt.Errorf("failed to parse filter")
		}

		column = codeNodes[0].Value
		value := codeNodes[1].Value
		operator := filter.Operator.Value

		// map the operator
		mappedOperator := mapOperator(operator)

		// validate this qual
		// - the column exists in the table
		// - the column is a key column
		// - the operator is supported
		if err := validateQual(column, mappedOperator, columnMap, keyColumns); err != nil {
			return nil, err
		}

		// convert the value string into a qual
		columnType := columnMap[column].Type
		qualValue, err := stringToQualValue(value, columnType)
		if err != nil {
			return nil, err
		}

		qual = &proto.Qual{
			FieldName: column,
			Operator:  &proto.Qual_StringValue{mappedOperator},
			Value:     qualValue,
		}

	case "in":
		if filter.Operator.Value == "not in" {
			return nil, fmt.Errorf("failed to convert 'where' arg to qual - 'not in' is not supported")
		}
		codeNodes, ok := filter.Values.([]filter2.CodeNode)
		if !ok || len(codeNodes) < 2 {
			return nil, fmt.Errorf("failed to parse filter")
		}
		column = codeNodes[0].Value
		operator := "="

		// map the operator
		mappedOperator := mapOperator(operator)

		// validate this qual
		// - the column exists in the table
		// - the colummn is a key column
		// - the operator is supported
		if err := validateQual(column, mappedOperator, columnMap, keyColumns); err != nil {
			return nil, err
		}

		// Build look up of values
		values := make(map[string]struct{}, len(codeNodes)-1)
		for _, c := range codeNodes[1:] {
			values[c.Value] = struct{}{}
		}

		// Convert these raw values into a qual
		columnType := columnMap[column].Type
		qualValue, err := stringToQualListValue(maps.Keys(values), columnType)
		if err != nil {
			return nil, err
		}

		// Create a Qual slice for the field and add the Qual to it
		qual = &proto.Qual{
			FieldName: column,
			Operator:  &proto.Qual_StringValue{mappedOperator},
			Value:     qualValue,
		}

	default:
		return nil, fmt.Errorf("failed to convert 'where' arg to qual")

	}

	if qual == nil {
		// unexpected
		return nil, fmt.Errorf("failed to convert 'where' arg to qual")
	}

	qualmap := make(map[string]*proto.Quals)
	qualmap[column] = &proto.Quals{Quals: []*proto.Qual{qual}}

	return qualmap, nil
}

// validate this qual
// - the column exists in the table
// - the colummn is a key column
// - the operator is supported
func validateQual(column, operator string, columnMap map[string]*proto.ColumnDefinition, quals []*proto.KeyColumn) error {
	// does the column exists in the table
	_, ok := columnMap[column]
	if !ok {
		return fmt.Errorf("column %s does not exist", column)
	}

	unsupportedOperator := false
	// is the column is a key column
	for _, keyColumn := range quals {
		// is this key column for the target column
		if keyColumn.Name == column {
			// check the operator is supported
			if isOperatorSupported(keyColumn.Operators, operator) {
				// ok this qual is valid
				return nil
			} else {
				unsupportedOperator = true
			}
		}
	}
	if unsupportedOperator {
		return fmt.Errorf("key column for '%s' does not support operator '%s'", column, operator)
	}
	return fmt.Errorf("there is no key column defined for column '%s'", column)
}

func stringToQualValue(valueString string, columnType proto.ColumnType) (*proto.QualValue, error) {
	result := &proto.QualValue{}
	switch columnType {
	case proto.ColumnType_BOOL:
		b, err := strconv.ParseBool(valueString)
		if err != nil {
			return nil, err
		}
		result.Value = &proto.QualValue_BoolValue{BoolValue: b}
	case proto.ColumnType_INT:
		i, err := strconv.ParseInt(valueString, 10, 64)
		if err != nil {
			return nil, err
		}
		result.Value = &proto.QualValue_Int64Value{Int64Value: i}
	case proto.ColumnType_DOUBLE:
		f, err := strconv.ParseFloat(valueString, 64)
		if err != nil {
			return nil, err
		}
		result.Value = &proto.QualValue_DoubleValue{DoubleValue: f}
	case proto.ColumnType_STRING:
		result.Value = &proto.QualValue_StringValue{StringValue: valueString}
	case proto.ColumnType_JSON:
		result.Value = &proto.QualValue_JsonbValue{JsonbValue: valueString}
	case proto.ColumnType_IPADDR:
		// todo parse
	case proto.ColumnType_CIDR:
		// todo parse
	case proto.ColumnType_INET:
		// todo parse

	case proto.ColumnType_DATETIME, proto.ColumnType_TIMESTAMP:
		//t, err := time.Parse("Mon Jan 2 15:04:05 MST 2006", valueString)
		//if err != nil{
		//	return nil, err
		//}
		//result.Value = &proto.QualValue_TimestampValue{TimestampValue: t}
		// todo parse
	case proto.ColumnType_LTREE:
		result.Value = &proto.QualValue_LtreeValue{LtreeValue: valueString}
	}

	if result.Value == nil {
		return nil, fmt.Errorf("faile to convert value string")
	}
	return result, nil
}

func stringToQualListValue(values []string, columnType proto.ColumnType) (*proto.QualValue, error) {
	res := &proto.QualValue{
		Value: &proto.QualValue_ListValue{
			ListValue: &proto.QualValueList{
				Values: make([]*proto.QualValue, len(values)),
			},
		},
	}
	for i, v := range values {
		qv, err := stringToQualValue(v, columnType)

		if err != nil {
			return nil, err
		}
		res.Value.(*proto.QualValue_ListValue).ListValue.Values[i] = qv
	}
	return res, nil
}

func setupLogger(plugin string) {
	level := logging.LogLevel()
	hcLevel := hclog.LevelFromString(level)

	options := &hclog.LoggerOptions{
		// make the name unique so that logs from this instance can be filtered
		Name:       fmt.Sprintf("[%s]", plugin),
		Level:      hcLevel,
		Output:     os.Stderr,
		TimeFn:     func() time.Time { return time.Now().UTC() },
		TimeFormat: "2006-01-02 15:04:05.000 UTC",
	}
	logger := logging.NewLogger(options)
	log.SetOutput(logger.StandardWriter(&hclog.StandardLoggerOptions{InferLevels: true}))
	log.SetPrefix("")
	log.SetFlags(0)
}

// mapOperator translates equivalent operator representations to a standard form.
func mapOperator(operator string) string {
	operatorMappings := map[string]string{
		"like":      "~~",
		"not like":  "!~~",
		"ilike":     "~~*",
		"not ilike": "!~~*",
	}

	// Check if the operator is in the mapping, if so, return the mapped value.
	if mappedOperator, ok := operatorMappings[operator]; ok {
		return mappedOperator
	}

	// If no mapping is found, return the original operator.
	return operator
}

func isOperatorSupported(keyColumns []string, mappedOperator string) bool {
	// Check if the mapped operator is supported.
	return slices.Contains(keyColumns, mappedOperator)
}
