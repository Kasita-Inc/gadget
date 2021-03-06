package database

import (
	"database/sql"
	"strings"
	"testing"

	"github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"

	"github.com/Kasita-Inc/gadget/errors"
	"github.com/Kasita-Inc/gadget/generator"
	"github.com/Kasita-Inc/gadget/log"
)

func TestExecutionError(t *testing.T) {
	err := NewExecutionError(Insert, "bar", errors.New("foo"), log.NewStackLogger()).(*SQLExecutionError)
	err2 := NewExecutionError(Insert, "bar", errors.New("foo"), log.NewStackLogger()).(*SQLExecutionError)
	assert.True(t, strings.HasPrefix(err.ReferenceID, dbErrPrefix))
	assert.NotEqual(t, err.ReferenceID, err2.ReferenceID)
	assert.Contains(t, err.Error(), err.ReferenceID)
	assert.Contains(t, err.Error(), err.message)
}

func TestNewNotFoundError(t *testing.T) {
	err := NewNotFoundError()
	assert.EqualError(t, err, NewNotFoundError().Error())
}

func TestNewSystemError(t *testing.T) {
	err := NewSystemError(Insert, "bar", errors.New("foo"), log.NewStackLogger()).(*SQLSystemError)
	err2 := NewSystemError(Insert, "bar", errors.New("foo"), log.NewStackLogger()).(*SQLSystemError)
	assert.True(t, strings.HasPrefix(err.ReferenceID, dbErrPrefix))
	assert.NotEqual(t, err.ReferenceID, err2.ReferenceID)
	assert.Contains(t, err.Error(), err.ReferenceID)
	assert.Contains(t, err.Error(), err.message)
}

func TestNewDuplicateRecordError(t *testing.T) {
	err := NewDuplicateRecordError(Insert, "bar", errors.New("foo"), log.NewStackLogger()).(*DuplicateRecordError)
	err2 := NewDuplicateRecordError(Insert, "bar", errors.New("foo"), log.NewStackLogger()).(*DuplicateRecordError)
	assert.True(t, strings.HasPrefix(err.ReferenceID, dbErrPrefix))
	assert.NotEqual(t, err.ReferenceID, err2.ReferenceID)
	assert.Contains(t, err.Error(), err.ReferenceID)
	assert.Contains(t, err.Error(), err.message)
}

func TestNewDataTooLongError(t *testing.T) {
	err := NewDataTooLongError(Insert, "bar", errors.New("foo"), log.NewStackLogger()).(*DataTooLongError)
	err2 := NewDataTooLongError(Insert, "bar", errors.New("foo"), log.NewStackLogger()).(*DataTooLongError)
	assert.True(t, strings.HasPrefix(err.ReferenceID, dbErrPrefix))
	assert.NotEqual(t, err.ReferenceID, err2.ReferenceID)
	assert.Contains(t, err.Error(), err.ReferenceID)
	assert.Contains(t, err.Error(), err.message)
}

func TestNewInvalidForeignKeyError(t *testing.T) {
	err := NewInvalidForeignKeyError(Insert, "bar", errors.New("foo"), log.NewStackLogger()).(*InvalidForeignKeyError)
	err2 := NewInvalidForeignKeyError(Insert, "bar", errors.New("foo"), log.NewStackLogger()).(*InvalidForeignKeyError)
	assert.True(t, strings.HasPrefix(err.ReferenceID, dbErrPrefix))
	assert.NotEqual(t, err.ReferenceID, err2.ReferenceID)
	assert.Contains(t, err.Error(), err.ReferenceID)
	assert.Contains(t, err.Error(), err.message)
}

func TestTranslateError(t *testing.T) {
	testData := []struct {
		err      error
		expected error
	}{
		{err: sql.ErrNoRows, expected: &NotFoundError{}},
		{err: &mysql.MySQLError{Number: mysqlDuplicateEntry, Message: "foo ... " + primaryKeyConstraintCheck}, expected: &DuplicateRecordError{}},
		{err: &mysql.MySQLError{Number: mysqlDuplicateEntry}, expected: &UniqueConstraintError{}},
		{err: &mysql.MySQLError{Number: mysqlDataTooLong}, expected: &DataTooLongError{}},
		{err: &mysql.MySQLError{Number: mysqlInvalidForeignKey}, expected: &InvalidForeignKeyError{}},
		{err: &mysql.MySQLError{}, expected: &SQLExecutionError{}},
		{err: errors.New("foo"), expected: &SQLSystemError{}},
	}
	for _, data := range testData {
		assert.IsType(t, data.expected, TranslateError(data.err, Select, generator.String(5), log.NewStackLogger()))
	}
}
