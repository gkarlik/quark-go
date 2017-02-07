package gorm

import (
	"github.com/gkarlik/quark-go/data/access/rdbms"
	"github.com/gkarlik/quark-go/logger"
	"github.com/jinzhu/gorm"
)

// DbContext represents relational database access mechanism using GORM library.
type DbContext struct {
	DB              *gorm.DB // raw gorm database object
	isInTransaction bool     // indicates if database context has already started the database transaction
}

// NewDbContext creates DbContext instance with specified SQL dialect and connection string.
// It should be create once per goroutine and reused. It opens database connection which should be closed in defer function using Dispose method.
func NewDbContext(dialect string, connStr string) (*DbContext, error) {
	logger.Log().DebugWithFields(logger.Fields{
		"dialect":           dialect,
		"connection_string": connStr,
		"component":         componentName,
	}, "Creating new database context")

	db, err := gorm.Open(dialect, connStr)
	if err != nil {
		logger.Log().ErrorWithFields(logger.Fields{
			"error":     err,
			"component": componentName,
		}, "Cannot create database context")

		return nil, err
	}

	return &DbContext{
		DB:              db,
		isInTransaction: false,
	}, nil
}

// IsInTransaction indicates if DbContext has started DbTransaction.
func (c *DbContext) IsInTransaction() bool {
	return c.isInTransaction
}

// BeginTransaction starts new database transaction. Panics if DbContext is already in transaction.
func (c *DbContext) BeginTransaction() rdbms.DbTransaction {
	if c.IsInTransaction() {
		logger.Log().PanicWithFields(logger.Fields{
			"component": componentName,
		}, "DbContext is already in transaction. Nested transactions are not supported!")
	}

	db := c.DB.Begin()

	return &DbTransaction{
		context: &DbContext{
			DB:              db,
			isInTransaction: true,
		},
	}
}

// Dispose closes database context and cleans up DbContext instance.
func (c *DbContext) Dispose() {
	logger.Log().InfoWithFields(logger.Fields{"component": componentName}, "Disposing database context")

	if c.DB != nil {
		err := c.DB.Close()
		if err != nil {
			logger.Log().ErrorWithFields(logger.Fields{
				"error":     err,
				"component": componentName,
			}, "Cannot close database context")
		}
		c.DB = nil
	}
}
