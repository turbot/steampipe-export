package main

import (
	"fmt"
	"log"
	"strconv"

	filter2 "github.com/turbot/steampipe-plugin-sdk/v5/filter"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/sperr"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

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

	case "compare":
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

		// validate this qual
		// - the column exists in the table
		// - the column is a key column
		// - the operator is supported
		if err := validateQual(column, operator, columnMap, keyColumns); err != nil {
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
			Operator:  &proto.Qual_StringValue{operator},
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

		// validate this qual
		// - the column exists in the table
		// - the colummn is a key column
		// - the operator is supported
		if err := validateQual(column, operator, columnMap, keyColumns); err != nil {
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
			Operator:  &proto.Qual_StringValue{operator},
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
			if slices.Contains(keyColumn.Operators, operator) {
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
