package models

import (
	"errors"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

var (
	// ErrNotFound is returned when a resource is not found in the database.
	ErrNotFound = errors.New("models: resource not found")

	// ErrInvalidID is returned when an invalid ID is provided.
	ErrInvalidID = errors.New("models: ID provided was invalid")
)

// first finds the first item in the query and place it into dst.
// dst should be a pointer.
func first(db *gorm.DB, dst interface{}) error {
	err := db.First(dst).Error
	if err == gorm.ErrRecordNotFound {
		return ErrNotFound
	}
	return err
}

// NewUserService provides a UserService object to peform user database actions.
func NewUserService(connectionInfo string) (*UserService, error) {
	db, err := gorm.Open("postgres", connectionInfo)
	if err != nil {
		return nil, err
	}
	// TODO: remove this
	db.LogMode(true)
	return &UserService{
		db: db,
	}, nil
}

// UserService is a utility to perform actions for a user
type UserService struct {
	db *gorm.DB
}

// ByID will look up and return a user by the provided ID.
func (us *UserService) ByID(id uint) (*User, error) {
	var user User
	db := us.db.Where("id = ?", id)
	err := first(db, &user)
	return &user, err
}

// ByEmail will look up and return a user by the provided email.
func (us *UserService) ByEmail(email string) (*User, error) {
	var user User
	db := us.db.Where("email = ?", email)
	err := first(db, &user)
	return &user, err
}

// Create will add the user to the database.
func (us *UserService) Create(user *User) error {
	return us.db.Create(user).Error
}

// Update will update the user with the provided user object.
func (us *UserService) Update(user *User) error {
	return us.db.Save(user).Error
}

// Delete will delete the user with the provided ID.
func (us *UserService) Delete(id uint) error {
	if id == 0 {
		return ErrInvalidID
	}
	user := User{Model: gorm.Model{ID: id}}
	return us.db.Delete(&user).Error
}

// Close closes the UserService database connection.
func (us *UserService) Close() error {
	return us.db.Close()
}

// DestructiveReset drops the user table and rebuilds it.
func (us *UserService) DestructiveReset() error {
	if err := us.db.DropTableIfExists(&User{}).Error; err != nil {
		return err
	}
	return us.AutoMigrate()
}

// AutoMigrate will attempt to automatically migrate the users table.
func (us *UserService) AutoMigrate() error {
	return us.db.AutoMigrate(&User{}).Error
}

// User represents a user's information.
type User struct {
	gorm.Model
	Name  string
	Email string `gorm:"not null;unique_index"`
}
