package main

import (
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"log"
	"testing"
)

func TestFilterStringToQual(t *testing.T) {
	tableSchema := &proto.TableSchema{
		Columns: []*proto.ColumnDefinition{
			{
				Name: "connection",
				Type: proto.ColumnType_STRING,
			},
		},
		ListCallKeyColumnList: []*proto.KeyColumn{{
			Name:      "connection",
			Operators: []string{"="},
		}},
	}

	testCases := []struct {
		filter   string
		expected []*proto.Qual
		err      string
	}{
		//comparisons
		{
			filter: "connection = 'foo'",
		},
		{
			filter: "connection != 'foo'",
			err:    "ERROR",
		},
		{
			filter: "connection <> 'foo'",
			err:    "ERROR",
		},
		// in
		{
			filter: "connection in ('foo','bar')",
		},
		{
			filter: "connection not in ('foo','bar')",
		},
		//like
		{
			filter: "connection like 'fo_'",
			err:    "ERROR",
		},
		{
			filter: "connection like 'fo_'",
			err:    "ERROR",
		},
		{
			filter: "connection like '_o_'",
			err:    "ERROR",
		},
		{
			filter: "connection like '_o_'",
			err:    "ERROR",
		},
		{
			filter: "connection like 'f%'",
			err:    "ERROR",
		},
		{
			filter: "connection like '%ob%'",
			err:    "ERROR",
		},
		{
			filter: "connection like 'fo_'",
			err:    "ERROR",
		},

		//ilike
		{
			filter: "connection ilike 'FO_'",
			err:    "ERROR",
		},
		// not  like
		{
			filter: "connection not like 'fo_'",
			err:    "ERROR",
		},
		{
			filter: "connection not like '_o_'",
			err:    "ERROR",
		},
		{
			filter: "connection not like 'f%'",
			err:    "ERROR",
		},
		{
			filter: "connection not like '%ob%'",
			err:    "ERROR",
		},
		{
			filter: "connection not like '_oo%'",
			err:    "ERROR",
		},
		{
			filter: "connection not like 'fo_'",
			err:    "ERROR",
		},
		{
			filter: "connection not like 'FO_'",
			err:    "ERROR",
		},
		// not ilike
		{
			filter: "connection not ilike 'FO_'",
			err:    "ERROR",
		},
		// complex queries
		{
			filter: "connection not in ('foo','bar') or connection='hello'",
			err:    "ERROR"},
		{
			filter: "connection in ('foo','bar') and connection='foo'",
			err:    "ERROR"},
		{
			filter: "connection in ('foo','bar') and connection='other'",
			err:    "ERROR"},
		{
			filter: "connection in ('a','b') or connection='foo'",
			err:    "ERROR"},
		{
			filter: "connection in ('a','b') or connection='c'",
			err:    "ERROR"},

		// not supported
		{
			// 'is not' not supported
			filter: "connection is null",
			err:    "ERROR",
		},
		{
			// 'is' not supported
			filter: "connection is not null",
			err:    "ERROR",
		},
		{
			// '<' is not supported
			filter: "connection < 'bar'",
			err:    "ERROR",
		},
		{
			// '<=' is not supported
			filter: "connection <= 'bar'",
			err:    "ERROR",
		},
		{
			// '>' is not supported
			filter: "connection > 'bar'",
			err:    "ERROR",
		},
		{
			// '>=' is not supported
			filter: "connection >= 'bar'",
			err:    "ERROR",
		},
	}
	for _, testCase := range testCases {
		quals, err := filterStringToQuals(testCase.filter, tableSchema)
		if testCase.err != "" {
			if err == nil /*|| err.Error() != testCase.err */ {
				t.Errorf("parseWhere(%v) err: %v, want %s", testCase.filter, err, testCase.err)
			}
			continue
		}
		if err != nil {
			t.Fatal(err)
		}

		log.Println(quals)

	}
}
