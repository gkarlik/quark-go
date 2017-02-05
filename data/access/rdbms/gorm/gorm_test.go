package gorm_test

import (
	"testing"

	"github.com/gkarlik/quark-go/data/access/rdbms"
	"github.com/gkarlik/quark-go/data/access/rdbms/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/stretchr/testify/assert"
)

const (
	dbConnStr = "host=localhost user=postgres dbname=quark_go_test sslmode=disable password="
	dialect   = "postgres"
)

type User struct {
	ID        uint `gorm:"primary_key"`
	Age       int
	Name      string `gorm:"size:200"`
	Emails    []Email
	Languages []Language `gorm:"many2many:user_language;"`
}

type Email struct {
	ID     int
	UserID int    `gorm:"index"`
	Email  string `gorm:"type:varchar(100);unique_index"`
}

type Language struct {
	ID   int
	Name string `gorm:"index:idx_name_code"`
	Code string `gorm:"index:idx_name_code"`
}

func NewDbContext() rdbms.DbContext {
	context, err := gorm.NewDbContext(dialect, dbConnStr)
	if err != nil {
		panic("Cannot connect to database!")
	}

	context.DB.SingularTable(true)
	context.DB.AutoMigrate(&User{}, &Email{}, &Language{})

	context.DB.LogMode(true)

	return context
}

func TestNewGormContext(t *testing.T) {
	context, err := gorm.NewDbContext(dialect, dbConnStr)
	defer context.Dispose()

	assert.NoError(t, err, "NewGormDbContext returned an error")
}

func TestNewGormContextInvalidDialect(t *testing.T) {
	assert.Panics(t, func() {
		context, err := gorm.NewDbContext("invalid_dialect", dbConnStr)
		defer context.Dispose()

		assert.Error(t, err, "NewGormDbContext should return an error")
	})
}

type UserRepository struct {
	*gorm.RepositoryBase
}

func NewUserRepository(c rdbms.DbContext) *UserRepository {
	repo := &UserRepository{
		RepositoryBase: &gorm.RepositoryBase{},
	}

	repo.SetContext(c)

	return repo
}

func (ur *UserRepository) FindUsersByLangugeCodeSortedByNamePaging(code string, page int, pageSize int) ([]User, error) {
	c := ur.Context().(*gorm.DbContext)

	var users []User

	if err := c.DB.Preload("Emails").Preload("Languages", "code = ?", code).Limit(pageSize).Offset(page * pageSize).Order("name desc").Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func TestRepositoryContext(t *testing.T) {
	context := NewDbContext()
	defer context.Dispose()

	repo := NewUserRepository(context)

	u := &User{
		Age:  35,
		Name: "Grzegorz",
	}

	repo.Save(u)

	tx := repo.Context().BeginTransaction()

	repo.SetContext(tx.Context())
	assert.Equal(t, tx.Context(), repo.Context())

	repo.Delete(u)
	repo.ResetContext()
	assert.Equal(t, context, repo.Context())

	tx.Commit()
}

func nestedTransaction(c rdbms.DbContext) {
	tx := c.BeginTransaction()

	tx.Rollback()
}

func TestNestedTransactions(t *testing.T) {
	context := NewDbContext()
	defer context.Dispose()

	tx := context.BeginTransaction()

	assert.Panics(t, func() {
		nestedTransaction(tx.Context())
	})

	tx.Rollback()
}

func TestRepositryBaseTest(t *testing.T) {
	context := NewDbContext()
	defer context.Dispose()

	tx := context.BeginTransaction()

	repo := NewUserRepository(tx.Context())

	user1 := &User{
		Age:  100,
		Name: "Joe",
		Emails: []Email{
			{0, 0, "joe@joe.pl"},
			{0, 0, "joe2@joe.pl"},
		},
		Languages: []Language{
			{0, "English", "EN"},
			{0, "German", "DE"},
		},
	}

	user2 := &User{
		Age:  50,
		Name: "John",
		Emails: []Email{
			{0, 0, "john@john.pl"},
		},
		Languages: []Language{
			{0, "English", "EN"},
		},
	}

	repo.Save(user1)
	repo.Save(user2)

	var u User
	repo.First(&u, User{ID: user1.ID})
	assert.Equal(t, user1.ID, u.ID)

	users, _ := repo.FindUsersByLangugeCodeSortedByNamePaging("EN", 1, 1)

	assert.Equal(t, 1, len(users))
	assert.Equal(t, user1.Name, users[0].Name)
	assert.Equal(t, len(user1.Emails), len(users[0].Emails))
	assert.Equal(t, len(user1.Languages[0].Code), len(users[0].Languages[0].Code))

	var email Email
	repo.Find(&email, "email = ?", "joe2@joe.pl")

	assert.Equal(t, "joe2@joe.pl", email.Email)

	tx.Rollback()
}
