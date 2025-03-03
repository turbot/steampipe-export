package main

import (
	"log"
	"testing"

	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
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

func TestStringToQualValue(t *testing.T) {
	tests := []struct {
		name        string
		valueString string
		columnType  proto.ColumnType
		want        *proto.QualValue
		wantErr     bool
	}{
		{
			name:        "Valid boolean true",
			valueString: "true",
			columnType:  proto.ColumnType_BOOL,
			want: &proto.QualValue{
				Value: &proto.QualValue_BoolValue{BoolValue: true},
			},
		},
		{
			name:        "Invalid boolean",
			valueString: "notabool",
			columnType:  proto.ColumnType_BOOL,
			wantErr:     true,
		},
		{
			name:        "Valid integer",
			valueString: "123",
			columnType:  proto.ColumnType_INT,
			want: &proto.QualValue{
				Value: &proto.QualValue_Int64Value{Int64Value: 123},
			},
		},
		{
			name:        "Invalid integer",
			valueString: "not_a_number",
			columnType:  proto.ColumnType_INT,
			wantErr:     true,
		},
		{
			name:        "Valid double",
			valueString: "123.45",
			columnType:  proto.ColumnType_DOUBLE,
			want: &proto.QualValue{
				Value: &proto.QualValue_DoubleValue{DoubleValue: 123.45},
			},
		},
		{
			name:        "Invalid double",
			valueString: "not_a_double",
			columnType:  proto.ColumnType_DOUBLE,
			wantErr:     true,
		},
		{
			name:        "Valid string",
			valueString: "test string",
			columnType:  proto.ColumnType_STRING,
			want: &proto.QualValue{
				Value: &proto.QualValue_StringValue{StringValue: "test string"},
			},
		},
		{
			name:        "Valid JSON",
			valueString: `{"key": "value"}`,
			columnType:  proto.ColumnType_JSON,
			want: &proto.QualValue{
				Value: &proto.QualValue_JsonbValue{JsonbValue: `{"key": "value"}`},
			},
		},
		{
			name:        "Valid timestamp RFC3339",
			valueString: "2024-03-19T10:30:00Z",
			columnType:  proto.ColumnType_TIMESTAMP,
		},
		{
			name:        "Valid timestamp RFC3339 with nanoseconds",
			valueString: "2024-03-19T10:30:00.123456789Z",
			columnType:  proto.ColumnType_TIMESTAMP,
		},
		{
			name:        "Valid timestamp RFC3339 with timezone offset",
			valueString: "2024-03-19T10:30:00+02:00",
			columnType:  proto.ColumnType_TIMESTAMP,
		},
		{
			name:        "Valid timestamp ISO8601 basic",
			valueString: "2024-03-19T10:30:05",
			columnType:  proto.ColumnType_TIMESTAMP,
		},
		{
			name:        "Valid timestamp with space separator",
			valueString: "2024-03-19 10:30:05",
			columnType:  proto.ColumnType_TIMESTAMP,
		},
		{
			name:        "Valid timestamp date only",
			valueString: "2024-03-19",
			columnType:  proto.ColumnType_TIMESTAMP,
		},
		{
			name:        "Valid timestamp RFC1123",
			valueString: "Tue, 19 Mar 2024 10:30:00 GMT",
			columnType:  proto.ColumnType_TIMESTAMP,
		},
		{
			name:        "Valid timestamp RFC1123Z",
			valueString: "Tue, 19 Mar 2024 10:30:00 +0000",
			columnType:  proto.ColumnType_TIMESTAMP,
		},
		{
			name:        "Valid timestamp RFC822",
			valueString: "19 Mar 24 10:30 GMT",
			columnType:  proto.ColumnType_TIMESTAMP,
		},
		{
			name:        "Valid timestamp RFC822Z",
			valueString: "19 Mar 24 10:30 +0000",
			columnType:  proto.ColumnType_TIMESTAMP,
		},
		{
			name:        "Valid Unix timestamp seconds",
			valueString: "1710842400", // 2024-03-19 10:00:00 UTC
			columnType:  proto.ColumnType_TIMESTAMP,
		},
		{
			name:        "Invalid timestamp format",
			valueString: "not_a_timestamp",
			columnType:  proto.ColumnType_TIMESTAMP,
			wantErr:     true,
		},
		{
			name:        "Invalid timestamp partial date",
			valueString: "2024-03",
			columnType:  proto.ColumnType_TIMESTAMP,
			wantErr:     true,
		},
		{
			name:        "Invalid timestamp bad month",
			valueString: "2024-13-19T10:30:00Z",
			columnType:  proto.ColumnType_TIMESTAMP,
			wantErr:     true,
		},
		{
			name:        "Invalid timestamp bad day",
			valueString: "2024-03-32T10:30:00Z",
			columnType:  proto.ColumnType_TIMESTAMP,
			wantErr:     true,
		},
		{
			name:        "Invalid timestamp bad hour",
			valueString: "2024-03-19T25:30:00Z",
			columnType:  proto.ColumnType_TIMESTAMP,
			wantErr:     true,
		},
		// Same tests for DATETIME type
		{
			name:        "Valid datetime RFC3339",
			valueString: "2024-03-19T10:30:00Z",
			columnType:  proto.ColumnType_DATETIME,
		},
		{
			name:        "Valid datetime RFC3339 with nanoseconds",
			valueString: "2024-03-19T10:30:00.123456789Z",
			columnType:  proto.ColumnType_DATETIME,
		},
		{
			name:        "Valid datetime RFC3339 with timezone offset",
			valueString: "2024-03-19T10:30:00+02:00",
			columnType:  proto.ColumnType_DATETIME,
		},
		{
			name:        "Valid datetime ISO8601 basic",
			valueString: "2024-03-19T10:30:05",
			columnType:  proto.ColumnType_DATETIME,
		},
		{
			name:        "Valid datetime with space separator",
			valueString: "2024-03-19 10:30:05",
			columnType:  proto.ColumnType_DATETIME,
		},
		{
			name:        "Valid datetime date only",
			valueString: "2024-03-19",
			columnType:  proto.ColumnType_DATETIME,
		},
		{
			name:        "Valid datetime RFC1123",
			valueString: "Tue, 19 Mar 2024 10:30:00 GMT",
			columnType:  proto.ColumnType_DATETIME,
		},
		{
			name:        "Valid datetime RFC1123Z",
			valueString: "Tue, 19 Mar 2024 10:30:00 +0000",
			columnType:  proto.ColumnType_DATETIME,
		},
		{
			name:        "Valid datetime RFC822",
			valueString: "19 Mar 24 10:30 GMT",
			columnType:  proto.ColumnType_DATETIME,
		},
		{
			name:        "Valid datetime RFC822Z",
			valueString: "19 Mar 24 10:30 +0000",
			columnType:  proto.ColumnType_DATETIME,
		},
		{
			name:        "Valid Unix datetime seconds",
			valueString: "1710842400", // 2024-03-19 10:00:00 UTC
			columnType:  proto.ColumnType_DATETIME,
		},
		{
			name:        "Invalid datetime format",
			valueString: "not_a_datetime",
			columnType:  proto.ColumnType_DATETIME,
			wantErr:     true,
		},
		{
			name:        "Invalid datetime partial date",
			valueString: "2024-03",
			columnType:  proto.ColumnType_DATETIME,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := stringToQualValue(tt.valueString, tt.columnType)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if got != nil {
					t.Fatal("expected nil result, got", got)
				}
				return
			}

			if err != nil {
				t.Fatal(err)
			}

			if tt.columnType == proto.ColumnType_TIMESTAMP {
				if got == nil {
					t.Fatal("got nil, want timestamp value")
				}
				if got.GetTimestampValue() == nil {
					t.Fatal("got nil timestamp value")
				}
			} else if tt.want != nil {
				// For non-timestamp values, compare the actual values
				switch v := got.GetValue().(type) {
				case *proto.QualValue_BoolValue:
					if want := tt.want.GetBoolValue(); v.BoolValue != want {
						t.Fatal("got", v.BoolValue, "want", want)
					}
				case *proto.QualValue_Int64Value:
					if want := tt.want.GetInt64Value(); v.Int64Value != want {
						t.Fatal("got", v.Int64Value, "want", want)
					}
				case *proto.QualValue_DoubleValue:
					if want := tt.want.GetDoubleValue(); v.DoubleValue != want {
						t.Fatal("got", v.DoubleValue, "want", want)
					}
				case *proto.QualValue_StringValue:
					if want := tt.want.GetStringValue(); v.StringValue != want {
						t.Fatal("got", v.StringValue, "want", want)
					}
				case *proto.QualValue_JsonbValue:
					if want := tt.want.GetJsonbValue(); v.JsonbValue != want {
						t.Fatal("got", v.JsonbValue, "want", want)
					}
				default:
					t.Fatal("got unexpected type", v)
				}
			}
		})
	}
}
