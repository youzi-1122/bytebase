package mysql

import (
	"context"
	"database/sql"
	"fmt"
	common2 "github.com/youzi-1122/bytebase/common"
	log2 "github.com/youzi-1122/bytebase/common/log"
	db2 "github.com/youzi-1122/bytebase/plugin/db"
	util2 "github.com/youzi-1122/bytebase/plugin/db/util"
	mysqlutil2 "github.com/youzi-1122/bytebase/resources/mysqlutil"
	"strings"

	"github.com/go-sql-driver/mysql"
	"go.uber.org/zap"
)

var (
	baseTableType = "BASE TABLE"
	viewTableType = "VIEW"

	_ db2.Driver = (*Driver)(nil)
)

func init() {
	db2.Register(db2.MySQL, newDriver)
	db2.Register(db2.TiDB, newDriver)
}

// Driver is the MySQL driver.
type Driver struct {
	connectionCtx db2.ConnectionContext
	connCfg       db2.ConnectionConfig
	dbType        db2.Type
	mysqlutil     mysqlutil2.Instance
	binlogDir     string
	db            *sql.DB
}

func newDriver(config db2.DriverConfig) db2.Driver {
	return &Driver{}
}

// Open opens a MySQL driver.
func (driver *Driver) Open(ctx context.Context, dbType db2.Type, connCfg db2.ConnectionConfig, connCtx db2.ConnectionContext) (db2.Driver, error) {
	protocol := "tcp"
	if strings.HasPrefix(connCfg.Host, "/") {
		protocol = "unix"
	}

	params := []string{"multiStatements=true"}

	port := connCfg.Port
	if port == "" {
		port = "3306"
		if dbType == db2.TiDB {
			port = "4000"
		}
	}

	tlsConfig, err := connCfg.TLSConfig.GetSslConfig()

	if err != nil {
		return nil, fmt.Errorf("sql: tls config error: %v", err)
	}

	loggedDSN := fmt.Sprintf("%s:<<redacted password>>@%s(%s:%s)/%s?%s", connCfg.Username, protocol, connCfg.Host, port, connCfg.Database, strings.Join(params, "&"))
	dsn := fmt.Sprintf("%s@%s(%s:%s)/%s?%s", connCfg.Username, protocol, connCfg.Host, port, connCfg.Database, strings.Join(params, "&"))
	if connCfg.Password != "" {
		dsn = fmt.Sprintf("%s:%s@%s(%s:%s)/%s?%s", connCfg.Username, connCfg.Password, protocol, connCfg.Host, port, connCfg.Database, strings.Join(params, "&"))
	}
	tlsKey := "db.mysql.tls"
	if tlsConfig != nil {
		if err := mysql.RegisterTLSConfig(tlsKey, tlsConfig); err != nil {
			return nil, fmt.Errorf("sql: failed to register tls config: %v", err)
		}
		// TLS config is only used during sql.Open, so should be safe to deregister afterwards.
		defer mysql.DeregisterTLSConfig(tlsKey)
		dsn += fmt.Sprintf("?tls=%s", tlsKey)
	}
	log2.Debug("Opening MySQL driver",
		zap.String("dsn", loggedDSN),
		zap.String("environment", connCtx.EnvironmentName),
		zap.String("database", connCtx.InstanceName),
	)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	driver.dbType = dbType
	driver.db = db
	driver.connectionCtx = connCtx
	driver.connCfg = connCfg

	return driver, nil
}

// Close closes the driver.
func (driver *Driver) Close(ctx context.Context) error {
	return driver.db.Close()
}

// Ping pings the database.
func (driver *Driver) Ping(ctx context.Context) error {
	return driver.db.PingContext(ctx)
}

// GetDbConnection gets a database connection.
func (driver *Driver) GetDbConnection(ctx context.Context, database string) (*sql.DB, error) {
	return driver.db, nil
}

// getDatabases gets all databases of an instance.
func getDatabases(ctx context.Context, txn *sql.Tx) ([]string, error) {
	var dbNames []string
	query := "SHOW DATABASES"
	rows, err := txn.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		dbNames = append(dbNames, name)
	}
	if err := rows.Err(); err != nil {
		return nil, util2.FormatErrorWithQuery(err, query)
	}
	return dbNames, nil
}

// GetVersion gets the version.
func (driver *Driver) GetVersion(ctx context.Context) (string, error) {
	query := "SELECT VERSION()"
	var version string
	if err := driver.db.QueryRowContext(ctx, query).Scan(&version); err != nil {
		if err == sql.ErrNoRows {
			return "", common2.FormatDBErrorEmptyRowWithQuery(query)
		}
		return "", util2.FormatErrorWithQuery(err, query)
	}
	return version, nil
}

// Execute executes a SQL statement.
func (driver *Driver) Execute(ctx context.Context, statement string) error {
	tx, err := driver.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, statement)

	if err == nil {
		if err := tx.Commit(); err != nil {
			return err
		}
	}

	return err
}

// Query queries a SQL statement.
func (driver *Driver) Query(ctx context.Context, statement string, limit int) ([]interface{}, error) {
	return util2.Query(ctx, driver.db, statement, limit)
}
