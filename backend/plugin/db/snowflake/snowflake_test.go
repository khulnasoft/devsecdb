// Package snowflake is the plugin for Snowflake driver.
package snowflake

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/khulnasoft/devsecdb/backend/common"
	"github.com/khulnasoft/devsecdb/backend/plugin/db"
)

func TestBuildSnowflakeDSN(t *testing.T) {
	testCases := []struct {
		input db.ConnectionConfig
		want  string
	}{
		{
			input: db.ConnectionConfig{
				Host:                 "nb47110.ap-southeast-1",
				Port:                 "443",
				Username:             "devsecdb",
				Password:             "pwd",
				MaximumSQLResultSize: common.DefaultMaximumSQLResultSize,
			},
			want: "devsecdb:pwd@nb47110.ap-southeast-1.snowflakecomputing.com:443?database=%22%22&ocspFailOpen=true&region=ap-southeast-1&validateDefaultParameters=true",
		},
		{
			input: db.ConnectionConfig{
				Host:                 "nb47110.ap-southeast-1",
				Port:                 "443",
				Username:             "devsecdb",
				Password:             "pwd",
				Database:             "SAMPLE_DB",
				MaximumSQLResultSize: common.DefaultMaximumSQLResultSize,
			},
			want: "devsecdb:pwd@nb47110.ap-southeast-1.snowflakecomputing.com:443?database=%22SAMPLE_DB%22&ocspFailOpen=true&region=ap-southeast-1&validateDefaultParameters=true",
		},
		{
			input: db.ConnectionConfig{
				Host:                 "nb47110.ap-southeast-1@10.0.0.1",
				Port:                 "4182",
				Username:             "devsecdb",
				Password:             "pwd",
				Database:             "",
				MaximumSQLResultSize: common.DefaultMaximumSQLResultSize,
			},
			want: "devsecdb:pwd@10.0.0.1:443?account=nb47110&database=%22%22&ocspFailOpen=true&region=ap-southeast-1&validateDefaultParameters=true",
		},
		{
			input: db.ConnectionConfig{
				Host:                 "nb47110.ap-southeast-1@10.0.0.1",
				Port:                 "4182",
				Username:             "devsecdb",
				Password:             "pwd",
				Database:             "SAMPLE_DB",
				MaximumSQLResultSize: common.DefaultMaximumSQLResultSize,
			},
			want: "devsecdb:pwd@10.0.0.1:443?account=nb47110&database=%22SAMPLE_DB%22&ocspFailOpen=true&region=ap-southeast-1&validateDefaultParameters=true",
		},
	}
	for _, testCase := range testCases {
		got, _ /* redacted */, err := buildSnowflakeDSN(testCase.input)
		require.NoError(t, err)
		require.Equal(t, testCase.want, got)
	}
}
