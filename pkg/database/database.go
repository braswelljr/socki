// Package database provides support for access the database.
package database

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // Calls init function.

	"github.com/braswelljr/socki/utils"
)

// Set of error variables for CRUD operations.
var (
	ErrNotFound              = errors.New("not found")
	ErrInvalidID             = errors.New("ID is not in its proper form")
	ErrAuthenticationFailure = errors.New("authentication failed")
	ErrForbidden             = errors.New("attempted action is not allowed")

	// Log
	Logger = utils.NewLogger()
)

// Config is the required properties to use the database.
type Config struct {
	User         string
	Password     string
	Host         string
	Name         string
	MaxIdleConns int
	MaxOpenConns int
	DisableTLS   bool
}

// Open knows how to open a database connection based on the configuration.
func Open(cfg Config) (*sqlx.DB, error) {
	sslMode := "require"
	if cfg.DisableTLS {
		sslMode = "disable"
	}

	q := make(url.Values)
	q.Set("sslmode", sslMode)
	q.Set("timezone", "utc")

	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(cfg.User, cfg.Password),
		Host:     cfg.Host,
		Path:     cfg.Name,
		RawQuery: q.Encode(),
	}

	db, err := sqlx.Open("postgres", u.String())
	if err != nil {
		return nil, err
	}
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetMaxOpenConns(cfg.MaxOpenConns)

	return db, nil
}

// StatusCheck returns nil if it can successfully talk to the database. It
// returns a non-nil error otherwise.
func StatusCheck(ctx context.Context, db *sqlx.DB) error {

	// First check we can ping the database.
	var pingError error
	for attempts := 1; ; attempts++ {
		pingError = db.Ping()
		if pingError == nil {
			break
		}
		time.Sleep(time.Duration(attempts) * 100 * time.Millisecond)
		if ctx.Err() != nil {
			// If the context is cancelled, stop retrying.
			Logger.Info().Msg("database.StatusCheck: ping cancelled")
			return ctx.Err()
		}
	}

	// Make sure we didn't timeout or be cancelled.
	if ctx.Err() != nil {
		return ctx.Err()
	}

	// Run a simple query to determine connectivity. Running this query forces a
	// round trip through the database.
	const q = `SELECT true`
	var tmp bool
	return db.QueryRowContext(ctx, q).Scan(&tmp)
}

// NamedExecContext is a helper function to execute a CUD operation with
// logging and tracing.
func NamedExecContext(ctx context.Context, db *sqlx.DB, query string, data interface{}) error {
	q := queryString(query, data)

	// log action
	Logger.Info().Str("query", q).Msg("database.NamedExecContext")

	if _, err := db.NamedExecContext(ctx, query, data); err != nil {
		// log error
		Logger.Error().Err(err).Str("query", q).Msg("failed to execute query")
		return err
	}

	return nil
}

// NamedQuerySlice is a helper function for executing queries that return a collection of data to be unmarshaled into a slice.
func NamedQuerySlice(ctx context.Context, db *sqlx.DB, query string, data interface{}, dest interface{}) error {
	q := queryString(query, data)
	// log action
	Logger.Info().Str("query", q).Msg("database.NamedQuerySlice")

	val := reflect.ValueOf(dest)
	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Slice {
		err := errors.New("must provide a pointer to a slice")

		// log error
		Logger.Error().Err(err).Msg("database.NamedQuerySlice")

		return err
	}

	rows, err := db.NamedQueryContext(ctx, query, data)
	if err != nil {
		// log error
		Logger.Error().Err(err).Str("query", q).Msg("error executing query")
		return err
	}

	slice := val.Elem()
	for rows.Next() {
		v := reflect.New(slice.Type().Elem())
		if err := rows.StructScan(v.Interface()); err != nil {
			// log error
			Logger.Error().Err(err).Str("query", q).Msg("failed to scan row into struct")

			return err
		}
		slice.Set(reflect.Append(slice, v.Elem()))
	}

	return nil
}

// NamedQueryStruct is a helper function for executing queries that return a single value to be unmarshalled into a struct type.
func NamedQueryStruct(ctx context.Context, db *sqlx.DB, query string, data interface{}, dest interface{}) error {
	q := queryString(query, data)
	// log action
	Logger.Info().Str("query", q).Msg("database.NamedQueryStruct")

	rows, err := db.NamedQueryContext(ctx, query, data)
	if err != nil {
		// log error
		Logger.Error().Err(err).Msg("database.NamedQueryStruct")
		return err
	}

	// check if row exists
	if !rows.Next() {
		// log error
		Logger.Error().Err(ErrNotFound).Str("query", q).Msg("row not found")
		return ErrNotFound
	}

	if err := rows.StructScan(dest); err != nil {
		// log error
		Logger.Error().Err(err).Msg("database.NamedQueryStruct")
		return err
	}

	return nil
}

// queryString provides a pretty print version of the query and parameters.
func queryString(query string, args ...interface{}) string {
	query, params, err := sqlx.Named(query, args)
	if err != nil {
		// log error
		Logger.Error().Err(err).Str("query", query).Msg("failed to create query string")
		return err.Error()
	}

	for _, param := range params {
		var value string
		switch v := param.(type) {
		case string:
			value = fmt.Sprintf("%q", v)
		case []byte:
			value = fmt.Sprintf("%q", string(v))
		default:
			value = fmt.Sprintf("%v", v)
		}
		query = strings.Replace(query, "?", value, 1)
	}

	query = strings.ReplaceAll(query, "\t", "")
	query = strings.ReplaceAll(query, "\n", " ")

	return strings.Trim(query, " ")
}
