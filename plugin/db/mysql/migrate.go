package mysql

import (
	"context"
	"database/sql"
	"fmt"
	common2 "github.com/youzi-1122/bytebase/common"
	log2 "github.com/youzi-1122/bytebase/common/log"
	db2 "github.com/youzi-1122/bytebase/plugin/db"
	util2 "github.com/youzi-1122/bytebase/plugin/db/util"
	"strings"

	// embed will embeds the migration schema.
	_ "embed"

	"go.uber.org/zap"
)

var (
	//go:embed mysql_migration_schema.sql
	migrationSchema string

	_ util2.MigrationExecutor = (*Driver)(nil)
)

// NeedsSetupMigration returns whether it needs to setup migration.
func (driver *Driver) NeedsSetupMigration(ctx context.Context) (bool, error) {
	const query = `
		SELECT
		    1
		FROM information_schema.TABLES
		WHERE TABLE_SCHEMA = 'bytebase' AND TABLE_NAME = 'migration_history'
		`
	return util2.NeedsSetupMigrationSchema(ctx, driver.db, query)
}

// SetupMigrationIfNeeded sets up migration if needed.
func (driver *Driver) SetupMigrationIfNeeded(ctx context.Context) error {
	setup, err := driver.NeedsSetupMigration(ctx)
	if err != nil {
		return nil
	}

	if setup {
		log2.Info("Bytebase migration schema not found, creating schema...",
			zap.String("environment", driver.connectionCtx.EnvironmentName),
			zap.String("database", driver.connectionCtx.InstanceName),
		)
		// Do not wrap it in a single transaction here because:
		// 1. For MySQL, each DDL is in its own transaction. See https://dev.mysql.com/doc/refman/8.0/en/implicit-commit.html
		// 2. For TiDB, the created database/table is not visible to the followup statements from the same transaction.
		if _, err := driver.db.ExecContext(ctx, migrationSchema); err != nil {
			log2.Error("Failed to initialize migration schema.",
				zap.Error(err),
				zap.String("environment", driver.connectionCtx.EnvironmentName),
				zap.String("database", driver.connectionCtx.InstanceName),
			)
			return util2.FormatErrorWithQuery(err, migrationSchema)
		}
		log2.Info("Successfully created migration schema.",
			zap.String("environment", driver.connectionCtx.EnvironmentName),
			zap.String("database", driver.connectionCtx.InstanceName),
		)
	}

	return nil
}

// FindLargestVersionSinceBaseline will find the largest version since last baseline or branch.
func (driver Driver) FindLargestVersionSinceBaseline(ctx context.Context, tx *sql.Tx, namespace string) (*string, error) {
	largestBaselineSequence, err := driver.FindLargestSequence(ctx, tx, namespace, true /* baseline */)
	if err != nil {
		return nil, err
	}
	const getLargestVersionSinceLastBaselineQuery = `
		SELECT MAX(version) FROM bytebase.migration_history
		WHERE namespace = ? AND sequence >= ?
	`
	var version sql.NullString
	if err := tx.QueryRowContext(ctx, getLargestVersionSinceLastBaselineQuery,
		namespace, largestBaselineSequence,
	).Scan(&version); err != nil {
		if err == sql.ErrNoRows {
			return nil, common2.FormatDBErrorEmptyRowWithQuery(getLargestVersionSinceLastBaselineQuery)
		}
		return nil, util2.FormatErrorWithQuery(err, getLargestVersionSinceLastBaselineQuery)
	}
	if version.Valid {
		return &version.String, nil
	}
	return nil, nil
}

// FindLargestSequence will return the largest sequence number.
func (Driver) FindLargestSequence(ctx context.Context, tx *sql.Tx, namespace string, baseline bool) (int, error) {
	findLargestSequenceQuery := `
		SELECT MAX(sequence) FROM bytebase.migration_history
		WHERE namespace = ?`
	if baseline {
		findLargestSequenceQuery = fmt.Sprintf("%s AND (type = '%s' OR type = '%s')", findLargestSequenceQuery, db2.Baseline, db2.Branch)
	}
	var sequence sql.NullInt32
	if err := tx.QueryRowContext(ctx, findLargestSequenceQuery,
		namespace,
	).Scan(&sequence); err != nil {
		if err == sql.ErrNoRows {
			return -1, common2.FormatDBErrorEmptyRowWithQuery(findLargestSequenceQuery)
		}
		return -1, util2.FormatErrorWithQuery(err, findLargestSequenceQuery)
	}
	if sequence.Valid {
		return int(sequence.Int32), nil
	}
	// Returns 0 if we haven't applied any migration for this namespace.
	return 0, nil
}

// InsertPendingHistory will insert the migration record with pending status and return the inserted ID.
func (Driver) InsertPendingHistory(ctx context.Context, tx *sql.Tx, sequence int, prevSchema string, m *db2.MigrationInfo, storedVersion, statement string) (int64, error) {
	const insertHistoryQuery = `
		INSERT INTO bytebase.migration_history (
			created_by,
			created_ts,
			updated_by,
			updated_ts,
			release_version,
			namespace,
			sequence,
			source,
			type,
			status,
			version,
			description,
			statement,
			` + "`schema`," + `
			schema_prev,
			execution_duration_ns,
			issue_id,
			payload
		)
		VALUES (?, unix_timestamp(), ?, unix_timestamp(), ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0, ?, ?)
		`
	res, err := tx.ExecContext(ctx, insertHistoryQuery,
		m.Creator,
		m.Creator,
		m.ReleaseVersion,
		m.Namespace,
		sequence,
		m.Source,
		m.Type,
		db2.Pending,
		storedVersion,
		m.Description,
		statement,
		prevSchema,
		prevSchema,
		m.IssueID,
		m.Payload,
	)
	if err != nil {
		return int64(0), util2.FormatErrorWithQuery(err, insertHistoryQuery)
	}

	insertedID, err := res.LastInsertId()
	if err != nil {
		return int64(0), util2.FormatErrorWithQuery(err, insertHistoryQuery)
	}
	return insertedID, nil
}

// UpdateHistoryAsDone will update the migration record as done.
func (Driver) UpdateHistoryAsDone(ctx context.Context, tx *sql.Tx, migrationDurationNs int64, updatedSchema string, insertedID int64) error {
	const updateHistoryAsDoneQuery = `
		UPDATE
			bytebase.migration_history
		SET
			status = ?,
			execution_duration_ns = ?,
		` + "`schema` = ?" + `
		WHERE id = ?
		`
	_, err := tx.ExecContext(ctx, updateHistoryAsDoneQuery, db2.Done, migrationDurationNs, updatedSchema, insertedID)
	return err
}

// UpdateHistoryAsFailed will update the migration record as failed.
func (Driver) UpdateHistoryAsFailed(ctx context.Context, tx *sql.Tx, migrationDurationNs int64, insertedID int64) error {
	const updateHistoryAsFailedQuery = `
		UPDATE
			bytebase.migration_history
		SET
			status = ?,
			execution_duration_ns = ?
		WHERE id = ?
		`
	_, err := tx.ExecContext(ctx, updateHistoryAsFailedQuery, db2.Failed, migrationDurationNs, insertedID)
	return err
}

// ExecuteMigration will execute the migration.
func (driver *Driver) ExecuteMigration(ctx context.Context, m *db2.MigrationInfo, statement string) (int64, string, error) {
	return util2.ExecuteMigration(ctx, driver, m, statement, db2.BytebaseDatabase)
}

// FindMigrationHistoryList finds the migration history.
func (driver *Driver) FindMigrationHistoryList(ctx context.Context, find *db2.MigrationHistoryFind) ([]*db2.MigrationHistory, error) {
	baseQuery := `
	SELECT
		id,
		created_by,
		created_ts,
		updated_by,
		updated_ts,
		release_version,
		namespace,
		sequence,
		source,
		type,
		status,
		version,
		description,
		statement,
		` + "`schema`," + `
		schema_prev,
		execution_duration_ns,
		issue_id,
		payload
		FROM bytebase.migration_history `
	paramNames, params := []string{}, []interface{}{}
	if v := find.ID; v != nil {
		paramNames, params = append(paramNames, "id"), append(params, *v)
	}
	if v := find.Database; v != nil {
		paramNames, params = append(paramNames, "namespace"), append(params, *v)
	}
	if v := find.Version; v != nil {
		// TODO(d): support semantic versioning.
		storedVersion, err := util2.ToStoredVersion(false, *v, "")
		if err != nil {
			return nil, err
		}
		paramNames, params = append(paramNames, "version"), append(params, storedVersion)
	}
	if v := find.Source; v != nil {
		paramNames, params = append(paramNames, "source"), append(params, *v)
	}
	var query = baseQuery +
		db2.FormatParamNameInQuestionMark(paramNames) +
		`ORDER BY created_ts DESC`
	if v := find.Limit; v != nil {
		query += fmt.Sprintf(" LIMIT %d", *v)
	}
	// TODO(zp):  modified param database of `util.FindMigrationHistoryList` when we support *mysql* database level.
	history, err := util2.FindMigrationHistoryList(ctx, query, params, driver, db2.BytebaseDatabase, find, baseQuery)
	// TODO(d): remove this block once all existing customers all migrated to semantic versioning.
	if err != nil {
		if !strings.Contains(err.Error(), "invalid stored version") {
			return nil, err
		}
		if err := driver.updateMigrationHistoryStorageVersion(ctx); err != nil {
			return nil, err
		}
		return util2.FindMigrationHistoryList(ctx, query, params, driver, db2.BytebaseDatabase, find, baseQuery)
	}
	return history, err
}

func (driver *Driver) updateMigrationHistoryStorageVersion(ctx context.Context) error {
	sqldb, err := driver.GetDbConnection(ctx, db2.BytebaseDatabase)
	if err != nil {
		return err
	}
	query := `SELECT id, version FROM bytebase.migration_history`
	rows, err := sqldb.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()
	type ver struct {
		id      int
		version string
	}
	var vers []ver
	for rows.Next() {
		var v ver
		if err := rows.Scan(&v.id, &v.version); err != nil {
			return err
		}
		vers = append(vers, v)
	}
	if err := rows.Err(); err != nil {
		return util2.FormatErrorWithQuery(err, query)
	}

	updateQuery := `
		UPDATE
			bytebase.migration_history
		SET
			version = ?
		WHERE id = ? AND version = ?
	`
	for _, v := range vers {
		if strings.HasPrefix(v.version, util2.NonSemanticPrefix) {
			continue
		}
		newVersion := fmt.Sprintf("%s%s", util2.NonSemanticPrefix, v.version)
		if _, err := sqldb.Exec(updateQuery, newVersion, v.id, v.version); err != nil {
			return err
		}
	}
	return nil
}
