package gorm

import (
	"github.com/gkarlik/quark-go/data/access/rdbms"
	"github.com/jinzhu/gorm"
)

const componentName = "GORM"

// Entity represents object stored in relational database.
type Entity struct {
	rdbms.Entity
	gorm.Model
}
