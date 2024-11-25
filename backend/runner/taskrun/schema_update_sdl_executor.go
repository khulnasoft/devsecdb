package taskrun

import (
	"context"
	"log/slog"

	"github.com/pkg/errors"

	"github.com/khulnasoft/devsecdb/backend/common"
	"github.com/khulnasoft/devsecdb/backend/common/log"
	"github.com/khulnasoft/devsecdb/backend/component/config"
	"github.com/khulnasoft/devsecdb/backend/component/dbfactory"
	"github.com/khulnasoft/devsecdb/backend/component/state"
	enterprise "github.com/khulnasoft/devsecdb/backend/enterprise/api"
	"github.com/khulnasoft/devsecdb/backend/plugin/db"
	"github.com/khulnasoft/devsecdb/backend/runner/schemasync"
	"github.com/khulnasoft/devsecdb/backend/runner/utils"
	"github.com/khulnasoft/devsecdb/backend/store"
	"github.com/khulnasoft/devsecdb/backend/store/model"
	backendutils "github.com/khulnasoft/devsecdb/backend/utils"
	storepb "github.com/khulnasoft/devsecdb/proto/generated-go/store"
)

// NewSchemaUpdateSDLExecutor creates a schema update (SDL) task executor.
func NewSchemaUpdateSDLExecutor(store *store.Store, dbFactory *dbfactory.DBFactory, license enterprise.LicenseService, stateCfg *state.State, schemaSyncer *schemasync.Syncer, profile *config.Profile) Executor {
	return &SchemaUpdateSDLExecutor{
		store:        store,
		dbFactory:    dbFactory,
		license:      license,
		stateCfg:     stateCfg,
		schemaSyncer: schemaSyncer,
		profile:      profile,
	}
}

// SchemaUpdateSDLExecutor is the schema update (SDL) task executor.
type SchemaUpdateSDLExecutor struct {
	store        *store.Store
	dbFactory    *dbfactory.DBFactory
	license      enterprise.LicenseService
	stateCfg     *state.State
	schemaSyncer *schemasync.Syncer
	profile      *config.Profile
}

// RunOnce will run the schema update (SDL) task executor once.
func (exec *SchemaUpdateSDLExecutor) RunOnce(ctx context.Context, driverCtx context.Context, task *store.TaskMessage, taskRunUID int) (bool, *storepb.TaskRunResult, error) {
	payload := &storepb.TaskDatabaseUpdatePayload{}
	if err := common.ProtojsonUnmarshaler.Unmarshal([]byte(task.Payload), payload); err != nil {
		return true, nil, errors.Wrap(err, "invalid database schema update payload")
	}

	sheetID := int(payload.SheetId)
	statement, err := exec.store.GetSheetStatementByID(ctx, sheetID)
	if err != nil {
		return true, nil, err
	}

	instance, err := exec.store.GetInstanceV2(ctx, &store.FindInstanceMessage{UID: &task.InstanceID})
	if err != nil {
		return true, nil, err
	}
	database, err := exec.store.GetDatabaseV2(ctx, &store.FindDatabaseMessage{UID: task.DatabaseID})
	if err != nil {
		return true, nil, err
	}

	materials := backendutils.GetSecretMapFromDatabaseMessage(database)
	// To avoid leaking the rendered statement, the error message should use the original statement and not the rendered statement.
	renderedStatement := backendutils.RenderStatement(statement, materials)

	version := model.Version{Version: payload.SchemaVersion}
	ddl, err := utils.ComputeDatabaseSchemaDiff(ctx, instance, database, exec.dbFactory, renderedStatement)
	if err != nil {
		return true, nil, errors.Wrap(err, "invalid database schema diff")
	}
	terminated, result, err := runMigration(ctx, driverCtx, exec.store, exec.dbFactory, exec.stateCfg, exec.schemaSyncer, exec.profile, task, taskRunUID, db.MigrateSDL, ddl, version, &sheetID)

	if err := exec.schemaSyncer.SyncDatabaseSchema(ctx, database, false /* force */); err != nil {
		slog.Error("failed to sync database schema",
			slog.String("instanceName", instance.ResourceID),
			slog.String("databaseName", database.DatabaseName),
			log.BBError(err),
		)
	}

	return terminated, result, err
}
