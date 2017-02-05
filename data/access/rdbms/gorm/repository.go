package gorm

import (
	"github.com/gkarlik/quark-go/data/access/rdbms"
)

// RepositoryBase represents base Repository implementation used to create specific repositories like UserRepository etc. (using GORM library)
type RepositoryBase struct {
	context     *DbContext // database context
	prevContext *DbContext // previous database context
}

// Context returns current database context associated with Repository.
func (r *RepositoryBase) Context() rdbms.DbContext {
	return r.context
}

// Save stores enity in relational database.
func (r *RepositoryBase) Save(entity interface{}) error {
	if err := r.context.DB.Save(entity).Error; err != nil {
		return err
	}
	return nil
}

// First retrives first entity from relational database which matches where criteria.
func (r *RepositoryBase) First(out interface{}, where ...interface{}) error {
	if err := r.context.DB.First(out, where...).Error; err != nil {
		return err
	}
	return nil
}

// Find retrives entities of out type which match where criteria.
func (r *RepositoryBase) Find(out interface{}, where ...interface{}) error {
	if err := r.context.DB.Find(out, where...).Error; err != nil {
		return err
	}
	return nil
}

// Delete removes specified entity or all entities of out type which match where criteria.
func (r *RepositoryBase) Delete(entity interface{}, where ...interface{}) error {
	if err := r.context.DB.Delete(entity, where...).Error; err != nil {
		return err
	}
	return nil
}

// SetContext sets current database context for Repository.
func (r *RepositoryBase) SetContext(c rdbms.DbContext) {
	r.prevContext = r.context
	r.context = c.(*DbContext)
}

// ResetContext restores last saved Repository database context.
func (r *RepositoryBase) ResetContext() {
	r.context = r.prevContext
}
