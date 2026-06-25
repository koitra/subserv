package testdb

import (
	"database/sql"
	"net/url"
	"os"
	"testing"

	"github.com/peterldowns/pgtestdb"
	"github.com/peterldowns/pgtestdb/migrators/goosemigrator"
	"github.com/stephenafamo/bob"
	"github.com/stretchr/testify/require"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/koitra/subserv"
)

func NewStd(t *testing.T) *sql.DB {
	m := goosemigrator.New("migrations",
		goosemigrator.WithFS(subserv.MigrationsFS),
	)
	dburl, err := url.Parse(getURL())
	require.NoError(t, err)
	password, _ := dburl.User.Password()

	db := pgtestdb.New(t, pgtestdb.Config{
		DriverName: "pgx",
		Host:       dburl.Hostname(),
		User:       dburl.User.Username(),
		Password:   password,
		Port:       dburl.Port(),
		Options:    dburl.Query().Encode(),
	}, m)

	return db
}

func New(t *testing.T) bob.DB {
	db := NewStd(t)
	return bob.NewDB(db)
}

func getURL() string {
	dsn := os.Getenv("PSQL_TEST_DSN")
	if dsn == "" {
		dsn = "postgres://subserv:verystron6@localhost:33210/subserv?sslmode=disable"
	}

	return dsn
}
