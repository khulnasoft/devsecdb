//go:build mysql

package tests

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	// Import pg driver.
	// init() in pgx will register it's pgx driver.
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"gopkg.in/yaml.v3"

	"github.com/khulnasoft/devsecdb/backend/common"
	api "github.com/khulnasoft/devsecdb/backend/legacyapi"
	"github.com/khulnasoft/devsecdb/backend/resources/mysql"
	"github.com/khulnasoft/devsecdb/backend/resources/postgres"
	"github.com/khulnasoft/devsecdb/backend/tests/fake"
	v1pb "github.com/khulnasoft/devsecdb/proto/generated-go/v1"
)

var (
	noSQLReviewPolicy = []*v1pb.PlanCheckRun_Result{
		{
			Status: v1pb.PlanCheckRun_Result_SUCCESS,
			Title:  "OK",
		},
	}
)

type test struct {
	Statement string
	Result    []*v1pb.PlanCheckRun_Result
	Run       bool
}

func TestSQLReviewForPostgreSQL(t *testing.T) {
	const (
		record = false
	)
	var (
		filepath   = filepath.Join("test-data", "sql_review_pg.yaml")
		statements = []string{
			`CREATE TABLE "user"(
				id INT,
				name VARCHAR(255) NOT NULL,
				room_id INT NOT NULL,
				creator_id INT NOT NULL,
				created_ts TIMESTAMP NOT NULL,
				updater_id INT NOT NULL,
				updated_ts TIMESTAMP NOT NULL,
				CONSTRAINT pk_user_id PRIMARY KEY (id),
				CONSTRAINT uk_user_id_name UNIQUE (id, name)
				);
				COMMENT ON TABLE "user" IS 'comment';`,
			`CREATE TABLE "userTable"(
				id INT,
				name VARCHAR(255),
				"roomId" INT,
				CONSTRAINT uk1 UNIQUE (id, name),
				CONSTRAINT fk1 FOREIGN KEY ("roomId") REFERENCES room(id)
				);
				COMMENT ON TABLE "userTable" IS 'comment';`,
		}
		databaseName = "testsqlreview"
	)

	t.Parallel()
	a := require.New(t)
	ctx := context.Background()
	ctl := &controller{}
	dataDir := t.TempDir()
	tests, err := readTestData(filepath)
	a.NoError(err)
	ctx, err = ctl.StartServerWithExternalPg(ctx, &config{
		dataDir:            dataDir,
		vcsProviderCreator: fake.NewGitLab,
	})
	a.NoError(err)
	defer ctl.Close(ctx)

	// Create a PostgreSQL instance.
	pgPort := getTestPort()
	stopInstance := postgres.SetupTestInstance(pgBinDir, t.TempDir(), pgPort)
	defer stopInstance()

	pgDB, err := sql.Open("pgx", fmt.Sprintf("host=/tmp port=%d user=root database=postgres", pgPort))
	a.NoError(err)
	defer pgDB.Close()

	err = pgDB.Ping()
	a.NoError(err)

	_, err = pgDB.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %v", databaseName))
	a.NoError(err)
	_, err = pgDB.Exec("CREATE USER devsecdb WITH ENCRYPTED PASSWORD 'devsecdb'")
	a.NoError(err)
	_, err = pgDB.Exec("ALTER USER devsecdb WITH SUPERUSER")
	a.NoError(err)

	reviewConfig, err := prodTemplateReviewConfigForPostgreSQL()
	a.NoError(err)

	createdConfig, err := ctl.reviewConfigServiceClient.CreateReviewConfig(ctx, &v1pb.CreateReviewConfigRequest{
		ReviewConfig: reviewConfig,
	})
	a.NoError(err)
	a.NotNil(createdConfig.Name)

	policy, err := ctl.orgPolicyServiceClient.CreatePolicy(ctx, &v1pb.CreatePolicyRequest{
		Parent: "environments/prod",
		Policy: &v1pb.Policy{
			Type: v1pb.PolicyType_TAG,
			Policy: &v1pb.Policy_TagPolicy{
				TagPolicy: &v1pb.TagPolicy{
					Tags: map[string]string{
						string(api.ReservedTagReviewConfig): createdConfig.Name,
					},
				},
			},
		},
	})
	a.NoError(err)
	a.NotNil(policy.Name)

	instance, err := ctl.instanceServiceClient.CreateInstance(ctx, &v1pb.CreateInstanceRequest{
		InstanceId: generateRandomString("instance", 10),
		Instance: &v1pb.Instance{
			Title:       "pgInstance",
			Engine:      v1pb.Engine_POSTGRES,
			Environment: "environments/prod",
			Activation:  true,
			DataSources: []*v1pb.DataSource{{Type: v1pb.DataSourceType_ADMIN, Host: "/tmp", Port: strconv.Itoa(pgPort), Username: "devsecdb", Password: "devsecdb", Id: "admin"}},
		},
	})
	a.NoError(err)

	err = ctl.createDatabaseV2(ctx, ctl.project, instance, nil /* environment */, databaseName, "devsecdb", nil)
	a.NoError(err)

	database, err := ctl.databaseServiceClient.GetDatabase(ctx, &v1pb.GetDatabaseRequest{
		Name: fmt.Sprintf("%s/databases/%s", instance.Name, databaseName),
	})
	a.NoError(err)

	for i, t := range tests {
		result := createIssueAndReturnSQLReviewResult(ctx, a, ctl, ctl.project, database, t.Statement, t.Run)
		if record {
			tests[i].Result = result
		} else {
			equalReviewResultProtos(a, t.Result, result, t.Statement)
		}
	}

	if record {
		err := writeTestData(filepath, tests)
		a.NoError(err)
	}

	// disable the SQL review policy
	policy.Enforce = false
	_, err = ctl.orgPolicyServiceClient.UpdatePolicy(ctx, &v1pb.UpdatePolicyRequest{
		Policy: policy,
		UpdateMask: &fieldmaskpb.FieldMask{
			Paths: []string{"enforce"},
		},
	})
	a.NoError(err)

	result := createIssueAndReturnSQLReviewResult(ctx, a, ctl, ctl.project, database, statements[0], false)
	equalReviewResultProtos(a, noSQLReviewPolicy, result, "")

	// delete the SQL review policy
	_, err = ctl.orgPolicyServiceClient.DeletePolicy(ctx, &v1pb.DeletePolicyRequest{
		Name: policy.Name,
	})
	a.NoError(err)

	result = createIssueAndReturnSQLReviewResult(ctx, a, ctl, ctl.project, database, statements[0], false)
	equalReviewResultProtos(a, noSQLReviewPolicy, result, "")
}

func TestSQLReviewForMySQL(t *testing.T) {
	const (
		record = false
	)
	var (
		filepath     = filepath.Join("test-data", "sql_review_mysql.yaml")
		databaseName = "testsqlreview"
		statements   = []string{
			"CREATE TABLE user(" +
				"id INT PRIMARY KEY COMMENT 'comment'," +
				"name VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'comment'," +
				"room_id INT NOT NULL DEFAULT 0 COMMENT 'comment'," +
				"creator_id INT NOT NULL DEFAULT 0 COMMENT 'comment'," +
				"created_ts TIMESTAMP NOT NULL DEFAULT NOW() COMMENT 'comment'," +
				"updater_id INT NOT NULL DEFAULT 0 COMMENT 'comment'," +
				"updated_ts TIMESTAMP NOT NULL DEFAULT NOW() ON UPDATE NOW() COMMENT 'comment'," +
				"INDEX idx_user_name(name)," +
				"UNIQUE KEY uk_user_id_name(id, name)" +
				") ENGINE = INNODB COMMENT 'comment'",
			"CREATE TABLE userTable(" +
				"id INT NOT NULL," +
				"name VARCHAR(255) CHARSET ascii," +
				"roomId INT," +
				"time_created TIMESTAMP NOT NULL DEFAULT NOW() ON UPDATE NOW() COMMENT 'comment'," +
				"time_updated TIMESTAMP NOT NULL DEFAULT NOW() ON UPDATE NOW() COMMENT 'comment'," +
				"content BLOB NOT NULL COMMENT 'comment'," +
				"json_content JSON NOT NULL COMMENT 'comment'," +
				"INDEX idx1(name)," +
				"UNIQUE KEY uk1(id, name)," +
				"FOREIGN KEY fk1(roomId) REFERENCES room(id)," +
				"INDEX idx_userTable_content(content)" +
				") ENGINE = CSV COLLATE latin1_bin",
		}
		valueTable = `(SELECT 1 AS id, 'a' AS name WHERE 1=1 UNION ALL
			SELECT 2 AS id, 'b' AS name WHERE 1=1 UNION ALL
			SELECT 3 AS id, 'c' AS name WHERE 1=1 UNION ALL
			SELECT 4 AS id, 'd' AS name WHERE 1=1 UNION ALL
			SELECT 5 AS id, 'e' AS name WHERE 1=1 UNION ALL
			SELECT 6 AS id, 'f' AS name WHERE 1=1 UNION ALL
			SELECT 7 AS id, 'g' AS name WHERE 1=1 UNION ALL
			SELECT 8 AS id, 'h' AS name WHERE 1=1 UNION ALL
			SELECT 9 AS id, 'i' AS name WHERE 1=1 UNION ALL
			SELECT 10 AS id, 'j' AS name WHERE 1=1) value_table`
		wantQueryResult = &v1pb.QueryResult{
			ColumnNames:     []string{"count(*)"},
			ColumnTypeNames: []string{"BIGINT"},
			Masked:          []bool{false},
			Sensitive:       []bool{false},
			Rows: []*v1pb.QueryRow{
				{
					Values: []*v1pb.RowValue{
						{Kind: &v1pb.RowValue_Int64Value{Int64Value: 4}},
					},
				},
			},
			Statement: "SELECT count(*) FROM test WHERE 1=1;",
		}
	)

	t.Parallel()
	a := require.New(t)
	ctx := context.Background()
	ctl := &controller{}
	tests, err := readTestData(filepath)
	a.NoError(err)
	dataDir := t.TempDir()
	ctx, err = ctl.StartServerWithExternalPg(ctx, &config{
		dataDir:            dataDir,
		vcsProviderCreator: fake.NewGitLab,
	})
	a.NoError(err)
	defer ctl.Close(ctx)

	// Create a MySQL instance.
	mysqlPort := getTestPort()
	stopInstance := mysql.SetupTestInstance(t, mysqlPort, mysqlBinDir)
	defer stopInstance()

	mysqlDB, err := sql.Open("mysql", fmt.Sprintf("root@tcp(127.0.0.1:%d)/mysql", mysqlPort))
	a.NoError(err)
	defer mysqlDB.Close()

	_, err = mysqlDB.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %v", databaseName))
	a.NoError(err)

	_, err = mysqlDB.Exec("DROP USER IF EXISTS devsecdb")
	a.NoError(err)
	_, err = mysqlDB.Exec("CREATE USER 'devsecdb' IDENTIFIED WITH mysql_native_password BY 'devsecdb'")
	a.NoError(err)
	_, err = mysqlDB.Exec("GRANT ALTER, ALTER ROUTINE, CREATE, CREATE ROUTINE, CREATE VIEW, DELETE, DROP, EVENT, EXECUTE, INDEX, INSERT, PROCESS, REFERENCES, SELECT, SHOW DATABASES, SHOW VIEW, TRIGGER, UPDATE, USAGE, REPLICATION CLIENT, REPLICATION SLAVE, LOCK TABLES, RELOAD ON *.* to devsecdb")
	a.NoError(err)

	reviewConfig, err := prodTemplateReviewConfigForMySQL()
	a.NoError(err)

	createdConfig, err := ctl.reviewConfigServiceClient.CreateReviewConfig(ctx, &v1pb.CreateReviewConfigRequest{
		ReviewConfig: reviewConfig,
	})
	a.NoError(err)
	a.NotNil(createdConfig.Name)

	policy, err := ctl.orgPolicyServiceClient.CreatePolicy(ctx, &v1pb.CreatePolicyRequest{
		Parent: "environments/prod",
		Policy: &v1pb.Policy{
			Type: v1pb.PolicyType_TAG,
			Policy: &v1pb.Policy_TagPolicy{
				TagPolicy: &v1pb.TagPolicy{
					Tags: map[string]string{
						string(api.ReservedTagReviewConfig): createdConfig.Name,
					},
				},
			},
		},
	})
	a.NoError(err)
	a.NotNil(policy.Name)

	instance, err := ctl.instanceServiceClient.CreateInstance(ctx, &v1pb.CreateInstanceRequest{
		InstanceId: generateRandomString("instance", 10),
		Instance: &v1pb.Instance{
			Title:       "mysqlInstance",
			Engine:      v1pb.Engine_MYSQL,
			Environment: "environments/prod",
			Activation:  true,
			DataSources: []*v1pb.DataSource{{Type: v1pb.DataSourceType_ADMIN, Host: "127.0.0.1", Port: strconv.Itoa(mysqlPort), Username: "devsecdb", Password: "devsecdb", Id: "admin"}},
		},
	})
	a.NoError(err)

	err = ctl.createDatabaseV2(ctx, ctl.project, instance, nil /* environment */, databaseName, "", nil)
	a.NoError(err)

	database, err := ctl.databaseServiceClient.GetDatabase(ctx, &v1pb.GetDatabaseRequest{
		Name: fmt.Sprintf("%s/databases/%s", instance.Name, databaseName),
	})
	a.NoError(err)

	for i, t := range tests {
		result := createIssueAndReturnSQLReviewResult(ctx, a, ctl, ctl.project, database, t.Statement, t.Run)
		if record {
			tests[i].Result = result
		} else {
			equalReviewResultProtos(a, t.Result, result, tests[i].Statement)
		}
	}

	if record {
		err := writeTestData(filepath, tests)
		a.NoError(err)
	}

	// test for dry-run-dml
	initialStmts := []string{
		"CREATE TABLE test(" +
			"id INT PRIMARY KEY COMMENT 'comment'," +
			"name VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'comment'," +
			"room_id INT NOT NULL DEFAULT 0 COMMENT 'comment'," +
			"creator_id INT NOT NULL DEFAULT 0 COMMENT 'comment'," +
			"created_ts TIMESTAMP NOT NULL DEFAULT NOW() COMMENT 'comment'," +
			"updater_id INT NOT NULL DEFAULT 0 COMMENT 'comment'," +
			"updated_ts TIMESTAMP NOT NULL DEFAULT NOW() ON UPDATE NOW() COMMENT 'comment'," +
			"INDEX idx_test_name(name)," +
			"UNIQUE KEY uk_test_id_name(id, name)" +
			") ENGINE = INNODB COMMENT 'comment';",
		`INSERT INTO test(id, name) VALUES (1, 'a'), (2, 'b'), (3, 'c'), (4, 'd');`,
	}
	for _, stmt := range initialStmts {
		createIssueAndReturnSQLReviewResult(ctx, a, ctl, ctl.project, database, stmt, true /* wait */)
	}
	countSQL := "SELECT count(*) FROM test WHERE 1=1;"
	dmlSQL := "INSERT INTO test SELECT * FROM " + valueTable
	originQueryResp, err := ctl.sqlServiceClient.Query(ctx, &v1pb.QueryRequest{
		Name:         database.Name,
		Statement:    countSQL,
		DataSourceId: "admin",
	})
	a.NoError(err)
	a.Equal(1, len(originQueryResp.Results))
	diff := cmp.Diff(wantQueryResult, originQueryResp.Results[0], protocmp.Transform(), protocmp.IgnoreMessages(&durationpb.Duration{}))
	a.Empty(diff)

	createIssueAndReturnSQLReviewResult(ctx, a, ctl, ctl.project, database, dmlSQL, false /* wait */)

	finalQueryResp, err := ctl.sqlServiceClient.Query(ctx, &v1pb.QueryRequest{
		Name:         database.Name,
		Statement:    countSQL,
		DataSourceId: "admin",
	})
	a.NoError(err)
	a.Equal(1, len(finalQueryResp.Results))
	diff = cmp.Diff(wantQueryResult, finalQueryResp.Results[0], protocmp.Transform(), protocmp.IgnoreMessages(&durationpb.Duration{}))
	a.Empty(diff)

	// disable the SQL review policy
	policy.Enforce = false
	_, err = ctl.orgPolicyServiceClient.UpdatePolicy(ctx, &v1pb.UpdatePolicyRequest{
		Policy: policy,
		UpdateMask: &fieldmaskpb.FieldMask{
			Paths: []string{"enforce"},
		},
	})
	a.NoError(err)

	// delete the SQL review policy
	_, err = ctl.orgPolicyServiceClient.DeletePolicy(ctx, &v1pb.DeletePolicyRequest{
		Name: policy.Name,
	})
	a.NoError(err)

	result := createIssueAndReturnSQLReviewResult(ctx, a, ctl, ctl.project, database, statements[0], false)
	equalReviewResultProtos(a, noSQLReviewPolicy, result, "")
}

func readTestData(path string) ([]test, error) {
	yamlFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer yamlFile.Close()
	byteValue, err := io.ReadAll(yamlFile)
	if err != nil {
		return nil, err
	}
	type yamlStruct struct {
		Statement string
		Result    []string
		Run       bool
	}
	var yamlTests []yamlStruct
	if err := yaml.Unmarshal(byteValue, &yamlTests); err != nil {
		return nil, err
	}

	var tests []test
	for _, yamlTest := range yamlTests {
		t := test{
			Statement: yamlTest.Statement,
			Run:       yamlTest.Run,
		}
		for _, r := range yamlTest.Result {
			result := &v1pb.PlanCheckRun_Result{}
			if err := common.ProtojsonUnmarshaler.Unmarshal([]byte(r), result); err != nil {
				return nil, err
			}
			t.Result = append(t.Result, result)
		}
		tests = append(tests, t)
	}
	return tests, nil
}

func writeTestData(filepath string, tests []test) error {
	type yamlStruct struct {
		Statement string
		Result    []string
		Run       bool
	}

	var yamlTests []yamlStruct
	for _, t := range tests {
		yamlTest := yamlStruct{
			Statement: t.Statement,
			Run:       t.Run,
		}
		for _, r := range t.Result {
			yamlTest.Result = append(yamlTest.Result, protojson.Format(r))
		}
		yamlTests = append(yamlTests, yamlTest)
	}

	byteValue, err := yaml.Marshal(yamlTests)
	if err != nil {
		return err
	}
	err = os.WriteFile(filepath, byteValue, 0644)
	if err != nil {
		return err
	}
	return nil
}

func createIssueAndReturnSQLReviewResult(ctx context.Context, a *require.Assertions, ctl *controller, project *v1pb.Project, database *v1pb.Database, statement string, wait bool) []*v1pb.PlanCheckRun_Result {
	sheet, err := ctl.sheetServiceClient.CreateSheet(ctx, &v1pb.CreateSheetRequest{
		Parent: project.Name,
		Sheet: &v1pb.Sheet{
			Title:   "statement",
			Content: []byte(statement),
		},
	})
	a.NoError(err)

	plan, err := ctl.planServiceClient.CreatePlan(ctx, &v1pb.CreatePlanRequest{
		Parent: project.Name,
		Plan: &v1pb.Plan{
			Steps: []*v1pb.Plan_Step{
				{
					Specs: []*v1pb.Plan_Spec{
						{
							Id: uuid.NewString(),
							Config: &v1pb.Plan_Spec_ChangeDatabaseConfig{
								ChangeDatabaseConfig: &v1pb.Plan_ChangeDatabaseConfig{
									Target: database.Name,
									Sheet:  sheet.Name,
									Type:   v1pb.Plan_ChangeDatabaseConfig_MIGRATE,
								},
							},
						},
					},
				},
			},
		},
	})
	a.NoError(err)

	result, err := ctl.GetSQLReviewResult(ctx, plan)
	a.NoError(err)

	if wait {
		a.NotNil(result)
		a.Len(result.Results, 1)
		a.Equal(v1pb.PlanCheckRun_Result_SUCCESS, result.Results[0].Status)
		issue, err := ctl.issueServiceClient.CreateIssue(ctx, &v1pb.CreateIssueRequest{
			Parent: project.Name,
			Issue: &v1pb.Issue{
				Type:        v1pb.Issue_DATABASE_CHANGE,
				Title:       fmt.Sprintf("change database %s", database.Name),
				Description: fmt.Sprintf("change database %s", database.Name),
				Plan:        plan.Name,
			},
		})
		a.NoError(err)
		rollout, err := ctl.rolloutServiceClient.CreateRollout(ctx, &v1pb.CreateRolloutRequest{Parent: project.Name, Rollout: &v1pb.Rollout{Plan: plan.Name}})
		a.NoError(err)
		err = ctl.waitRollout(ctx, issue.Name, rollout.Name)
		a.NoError(err)
		// Wait some time till written data becomes consistent.
		time.Sleep(5 * time.Second)
	}

	return result.Results
}

func equalReviewResultProtos(a *require.Assertions, want, got []*v1pb.PlanCheckRun_Result, message string) {
	a.Equal(len(want), len(got))
	for i := 0; i < len(want); i++ {
		diff := cmp.Diff(want[i], got[i], protocmp.Transform())
		a.Empty(diff, message)
	}
}
