package gorm

import (
	"github.com/gkarlik/quark-go/data/access/rdbms"
	"github.com/gkarlik/quark-go/logger"
)

// DbTransaction represents transaction in relational database using GORM library.
type DbTransaction struct {
	context *DbContext // database context
}

// Context returns DbContext associated with DbTransaction.
func (t *DbTransaction) Context() rdbms.DbContext {
	return t.context
}

// Rollback rollbacks database transaction.
func (t *DbTransaction) Rollback() error {
	logger.Log().DebugWithFields(logger.Fields{"component": componentName}, "Transaction rollback")

	if err := t.context.DB.Rollback().Error; err != nil {
		logger.Log().ErrorWithFields(logger.Fields{
			"error":     err,
			"component": componentName,
		}, "Cannot rollback the transaction")
		return err
	}

	return nil
}

// Commit commits database transaction.
func (t *DbTransaction) Commit() error {
	logger.Log().DebugWithFields(logger.Fields{"component": componentName}, "Transaction commit")

	if err := t.context.DB.Commit().Error; err != nil {
		logger.Log().ErrorWithFields(logger.Fields{
			"error":     err,
			"component": componentName,
		}, "Cannot commit the transaction")

		return err
	}

	return nil
}
