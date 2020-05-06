// +build int

package pgtesthelper_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/robojandro/go-pgtesthelper"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHelper(t *testing.T) {
	var (
		schemaPath = "./testdata/books.sql"
		keepDB     = false
		dbPrefix   = "testing"
		dbUser     = ""
		dbPass     = ""
	)

	if dbUser = os.Getenv("pgtesthelper_user"); dbUser == "" {
		t.Skip("missing env variable pgtesthelper_user")
	}

	if dbPass = os.Getenv("pgtesthelper_pass"); dbPass == "" {
		t.Skip("missing env variable pgtesthelper_pass")
	}

	t.Run("NewHelper", func(t *testing.T) {
		t.Run("happy", func(t *testing.T) {
			h, err := pgtesthelper.NewHelper(schemaPath, dbPrefix, dbUser, dbPass, keepDB)
			require.NoError(t, err)
			assert.Contains(t, h.DBName(), "testing_")
		})
		t.Run("bad creds", func(t *testing.T) {
			h, err := pgtesthelper.NewHelper(schemaPath, dbPrefix, "xxxx", dbPass, keepDB)
			require.Error(t, err)
			assert.NotContains(t, h.DBName(), "testing_")
		})
	})

	t.Run("not_keeping", func(t *testing.T) {
		h, err := pgtesthelper.NewHelper(schemaPath, dbPrefix, dbUser, dbPass, keepDB)
		require.NoError(t, err)

		_, err = h.CreateTempDB()
		require.NoError(t, err)

		// will panic because table database was dropped
		err = h.CleanUp()
		require.NoError(t, err)
		require.Panics(t, func() {
			_ = h.CleanTables([]string{"books"})
		})
	})

	t.Run("keeping_database", func(t *testing.T) {
		h, err := pgtesthelper.NewHelper(schemaPath, dbPrefix, dbUser, dbPass, true)
		require.NoError(t, err)

		_, err = h.CreateTempDB()
		require.NoError(t, err)

		// cleanup should be a noop since, but also shouldn't error out
		err = h.CleanUp()
		require.NoError(t, err)
		require.NotPanics(t, func() {
			h.CleanTables([]string{"books"})
		})

		// now you have to clean up the temp db manually however, starting with closing
		// the connection
		h.CloseConnection()
		pgDB, err := sqlx.Connect("postgres", fmt.Sprintf("user=%s dbname=%s sslmode=disable", dbUser, "postgres"))
		require.NoError(t, err)
		qry, err := pgDB.Query(fmt.Sprintf("DROP DATABASE IF EXISTS %s;", h.DBName()))
		require.NoError(t, err)
		defer qry.Close()
		pgDB.Close()
	})

	t.Run("data_loading", func(t *testing.T) {
		h, err := pgtesthelper.NewHelper(schemaPath, dbPrefix, dbUser, dbPass, keepDB)
		require.NoError(t, err)

		dbh, err := h.CreateTempDB()
		require.NoError(t, err)
		defer h.CleanUp()

		rows, err := dbh.Query(`INSERT INTO books (id, title, isbn, created_at) VALUES
			('cb0b9721-7631-4b2a-94a2-493c559da893','titleA', '9783161484100', NOW());`)
		require.NoError(t, err)
		rows.Close()

		rows, err = dbh.Query("SELECT * FROM BOOKS;")
		defer rows.Close()
		require.NoError(t, err)

		// Book is the model representing an book rows used in these tests.
		type Book struct {
			ID        string     `db:"id" json:"id,omitempty"`
			Title     string     `db:"title" json:"title,omitempty"`
			ISBN      string     `db:"isbn" json:"isbn,omitempty"`
			CreatedAt *time.Time `db:"created_at" json:"created_at,omitempty"`
		}
		res := []Book{}
		for rows.Next() {
			var b Book
			err = rows.Scan(&b.ID, &b.Title, &b.ISBN, &b.CreatedAt)
			require.NoError(t, err)
			res = append(res, b)
		}

		assert.Equal(t, "cb0b9721-7631-4b2a-94a2-493c559da893", res[0].ID)
		assert.Equal(t, "titleA", res[0].Title)
		assert.Equal(t, "9783161484100", res[0].ISBN)
	})
}
