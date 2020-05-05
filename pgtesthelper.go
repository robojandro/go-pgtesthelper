package pgtesthelper

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	errors "github.com/pkg/errors"
)

var driver = "postgres"

// Helper is a struct containing private references to the database handles necessary for creating and using temporary db, a prefix for naming it, and connection details.
type Helper struct {
	pgDB *sqlx.DB
	db   *sqlx.DB

	schemaPath string
	dbPrefix   string
	dbName     string
	dbUser     string
	dbPass     string

	keepDB bool
}

// NewHelper returns a new pgtesthelper.Helper value after establishing a connection to the local postgres database
func NewHelper(schemaPath, dbPrefix, dbUser, dbPass string, keepDB bool) (Helper, error) {
	now := time.Now()
	h := Helper{
		dbName:     fmt.Sprintf("%s_%d", dbPrefix, now.Unix()),
		schemaPath: schemaPath,
		dbUser:     dbUser,
		dbPass:     dbPass,
		dbPrefix:   dbPrefix,
		keepDB:     keepDB,
	}
	if err := h.pgDBConnect(); err != nil {
		return Helper{}, err
	}

	return h, nil
}

// CreateTestingDB creates database based on the schemaPath schema for use with testing.  It is left upto the called to the call CleanUp() to clean up artifacts.
// The database name will be suffixed with a unix timestamp giving down to the second uniqueness.
func (h *Helper) CreateTestingDB() error {
	if err := h.privExecute(fmt.Sprintf("DROP DATABASE IF EXISTS %s;", h.dbName)); err != nil {
		return errors.Wrapf(err, "failed to drop %s\n", h.dbName)
	}

	log.Printf("creating db: %s", h.dbName)
	if err := h.privExecute(fmt.Sprintf("CREATE DATABASE %s;", h.dbName)); err != nil {
		return errors.Wrapf(err, "failed to create %s\n", h.dbName)
	}

	if err := h.privExecute(fmt.Sprintf("GRANT ALL PRIVILEGES ON DATABASE %s TO %s", h.dbName, h.dbUser)); err != nil {
		return errors.Wrapf(err, "failed to grant privileges %s\n", h.dbName)
	}

	//connect to the just created DB
	db, err := sqlx.Connect("postgres",
		fmt.Sprintf("user=%s dbname=%s sslmode=disable", h.dbUser, h.dbName))
	if err != nil {
		return errors.Wrapf(err, "cannot connect to %s\n", h.dbName)
	}
	if err := db.Ping(); err != nil {
		return errors.Wrapf(err, "could not ping %s\n", h.dbName)
	}
	h.db = db

	//read and apply the schema
	schema, err := ioutil.ReadFile(h.schemaPath)
	if err != nil {
		return errors.Wrapf(err, "failed to read schemaPath %s\n", h.schemaPath)
	}

	if err := h.execute(string(schema)); err != nil {
		return errors.Wrapf(err, "failed to create scheam %s\n", schema)
	}
	return nil
}

func (h *Helper) DBName() string {
	return h.dbName
}

// ParseMockData allows the calling code to control how to insert data.
func (h *Helper) ParseMockData(mockDataFile string, fn func(mocked []byte) error) error {
	mockData, err := ioutil.ReadFile(mockDataFile)
	if err != nil {
		return err
	}
	return fn(mockData)
}

// LoadData allows the calling code to control how to insert data.
func (h *Helper) LoadData(mockObj interface{}, insertFn func(db *sqlx.DB) error) error {
	return insertFn(h.db)
}

func (h *Helper) CloseConnection() {
	h.db.Close()
}

// CleanUp will remove artifacts from testing. Currently that means dropping the temporary database created for this usage.
func (h *Helper) CleanUp() error {
	if h.keepDB {
		log.Printf("keeping db: %s", h.dbName)
		return nil
	}
	h.CloseConnection()
	if err := h.privExecute(fmt.Sprintf("DROP DATABASE IF EXISTS %s;", h.dbName)); err != nil {
		return errors.Wrapf(err, "could not drop database %s", h.dbName)
	}
	log.Printf("removed db: %s", h.dbName)
	return nil
}

// CleanTables loops over the given list of tables and attempts to truncate them. It will return an error rather than halting execution.
func (h *Helper) CleanTables(tables []string) error {
	tx := h.db.MustBegin()
	for _, table := range tables {
		log.Printf("clearing out table: %s\n", table)
		res := tx.MustExec(fmt.Sprintf("TRUNCATE TABLE %s", table))
		if res == nil {
			if err := tx.Rollback(); err != nil {
				return errors.Wrapf(err, "failed to rollback trucate %s\n", h.dbName)
			}
			return errors.New(fmt.Sprintf("failed to truncate %s\n", h.dbName))
		}
	}
	if err := tx.Commit(); err != nil {
		return errors.Wrapf(err, "failed to commit truncating %s: \n", h.dbName)
	}
	return nil
}

// Query passthrough for arbitrary queries to the temporary database.
func (h *Helper) Query(query string) (*sql.Rows, error) {
	rows, err := h.db.Query(query)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (h *Helper) execute(query string) error {
	qry, err := h.db.Query(query)
	if err != nil {
		return err
	}
	defer qry.Close()
	return nil
}

func (h *Helper) privExecute(query string) error {
	qry, err := h.pgDB.Query(query)
	if err != nil {
		return err
	}
	defer qry.Close()
	return nil
}

func (h *Helper) pgDBConnect() error {
	pgDB, err := sqlx.Connect(driver, fmt.Sprintf("user=%s dbname=%s sslmode=disable", h.dbUser, "postgres"))
	if err != nil {
		return errors.Wrap(err, "cannot connect to postgres db")
	}
	if err := pgDB.Ping(); err != nil {
		return errors.Wrap(err, "failed to ping postgres db")
	}
	h.pgDB = pgDB
	return nil
}
