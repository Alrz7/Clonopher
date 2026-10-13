package loggy

import (
	"errors"
	"fmt"
	"os"
	"time"

	"charm.land/log/v2"
)

var DefaultLogger = log.NewWithOptions(os.Stderr, log.Options{
	ReportCaller:    true,
	ReportTimestamp: true,
	TimeFormat:      time.Kitchen,
	Prefix:          "clonopher",
	CallerOffset:    1,
})

func Say(msg string) error {
	return errors.New(msg)
}

func Sayf(format string, params ...any) error {
	return fmt.Errorf(format, params...)
}

func Sayr(msg string, err error) error {
	return fmt.Errorf("%s, err : %v", msg, err)
}

func Info(msg any, keyvals ...any) {
	DefaultLogger.Info(msg, keyvals...)
}
func Infof(format string, args ...any) {
	DefaultLogger.Infof(format, args...)
}
func Warn(msg any, keyvals ...any) {
	DefaultLogger.Warn(msg, keyvals...)
}
func Warnf(format string, args ...any) {
	DefaultLogger.Warnf(format, args...)
}
func Error(msg any, keyvals ...any) {
	DefaultLogger.Error(msg, keyvals...)
}
func Errorf(format string, args ...any) {
	DefaultLogger.Errorf(format, args...)
}
func Fatal(msg any, keyvals ...any) {
	DefaultLogger.Fatal(msg, keyvals...)
}
func Fatalf(format string, args ...any) {
	DefaultLogger.Fatalf(format, args...)
}

const (
	ErrDbConnection        string = "dbConnection"
	ErrDbQuery             string = "dbQuery"
	ErrNoRecord            string = "noRecord"
	ErrDuplicate           string = "duplicateRecord"
	ErrMigration           string = "migrationFailed"
	ErrDuplicateRecord     string = "DUPLICATE_RECORD"
	ErrForeignKeyViolation string = "FOREIGN_KEY_VIOLATION"
	ErrNotNullViolation    string = "NOT_NULL_VIOLATION"
	ErrCheckViolation      string = "CHECK_VIOLATION"
	ErrValueTooLong        string = "VALUE_TOO_LONG"
	ErrTransactionConflict string = "TRANSACTION_CONFLICT"
	ErrDeadlockDetected    string = "DEADLOCK_DETECTED"
	ErrQueryTimeout        string = "QUERY_TIMEOUT"
	ErrDatabaseUnavailable string = "DATABASE_UNAVAILABLE"

	ErrUnauthorized string = "unauthorized"
	ErrForbidden    string = "forbidden"
	ErrInvalidToken string = "invalidToken"
	ErrExpiredToken string = "expiredToken"

	ErrBadRequest       string = "badRequest"
	ErrValidationFailed string = "validationFailed"
	ErrNotFound         string = "notFound"
	ErrConflict         string = "conflict"
	ErrPayloadTooLarge  string = "payloadTooLarge"

	ErrTimeout     string = "timeout"
	ErrExternalApi string = "externalApiFailed"
	ErrWsHandshake string = "wsHandshakeFailed"
	ErrWsClosed    string = "wsClosed"

	ErrInternalServer string = "internalServer"
	ErrNotImplemented string = "notImplemented"
	ErrFileIO         string = "fileSystemError"

	ErrUnexpected string = "unexpectedError"
)
