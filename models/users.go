package models

import (
	"errors"

	// This needs to be imported to initialize the gorm postgres package
	_ "github.com/jinzhu/gorm/dialects/postgres"

	"github.com/jinzhu/gorm"
	"github.com/mrpineapples/lenslocked/hash"
	"github.com/mrpineapples/lenslocked/rand"
	"golang.org/x/crypto/bcrypt"
)

var (
	// ErrNotFound is returned when a resource is not found in the database.
	ErrNotFound = errors.New("models: resource not found")

	// ErrInvalidID is returned when an invalid ID is provided.
	ErrInvalidID = errors.New("models: ID provided was invalid")

	// ErrInvalidPassword is returned when a password match is not found in the database.
	ErrInvalidPassword = errors.New("models: incorrect password provided")
)

const userPwPepper = "u3lx@T!I8gdKLwsB*q8TsCVxI0LW50rF"
const hmacSecretKey = "yjqRz4166W6@RvFd#b59yGT6uSIsVJh#"

// User represents the user model stored in our database.
type User struct {
	gorm.Model
	Name         string
	Email        string `gorm:"not null;unique_index"`
	Password     string `gorm:"-"`
	PasswordHash string `gorm:"not null"`
	Remember     string `gorm:"-"`
	RememberHash string `gorm:"not null;unique_index"`
}

// UserDB is used to interact with the users database.
type UserDB interface {
	// Methods for querying a single user
	ByID(id uint) (*User, error)
	ByEmail(email string) (*User, error)
	ByRemember(token string) (*User, error)

	// Methods for altering users
	Create(user *User) error
	Update(user *User) error
	Delete(id uint) error

	// Used to close a DB connection
	Close() error

	// Migration helpers
	AutoMigrate() error
	DestructiveReset() error
}

// UserService is a set of methods used to work with the user model.
type UserService interface {
	// Authenticate will verify that the user's email and password are
	// valid in the database; If verified, the user will be returned.
	Authenticate(email, password string) (*User, error)
	UserDB
}

// NewUserService provides a UserService object to peform user database actions.
func NewUserService(connectionInfo string) (UserService, error) {
	ug, err := newUserGorm(connectionInfo)
	if err != nil {
		return nil, err
	}
	hmac := hash.NewHMAC(hmacSecretKey)
	uv := &userValidator{
		hmac:   hmac,
		UserDB: ug,
	}
	return &userService{
		UserDB: uv,
	}, nil
}

// Test that userService fulfills the UserService interface.
var _ UserService = &userService{}

type userService struct {
	UserDB
}

// Authenticate verifies if a user's email and password exists.
func (us *userService) Authenticate(email, password string) (*User, error) {
	foundUser, err := us.ByEmail(email)
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(foundUser.PasswordHash), []byte(password+userPwPepper))
	if err != nil {
		switch err {
		case bcrypt.ErrMismatchedHashAndPassword:
			return nil, ErrInvalidPassword
		default:
			return nil, err
		}
	}
	return foundUser, nil
}

type userValidatorFunc func(*User) error

func runUserValidatorFuncs(user *User, fns ...userValidatorFunc) error {
	for _, fn := range fns {
		if err := fn(user); err != nil {
			return err
		}
	}
	return nil
}

// Test that userValidator fulfills the UserDB interface.
var _ UserDB = &userValidator{}

type userValidator struct {
	UserDB
	hmac hash.HMAC
}

// ByRemember hashes the remember token and then calls ByRemember
// on the subsequent UserDB layer.
func (uv *userValidator) ByRemember(token string) (*User, error) {
	user := User{
		Remember: token,
	}

	err := runUserValidatorFuncs(&user, uv.hmacRemember)
	if err != nil {
		return nil, err
	}

	return uv.UserDB.ByRemember(user.RememberHash)
}

// Create hashes the user's password and sets a remember token on the user.
func (uv *userValidator) Create(user *User) error {
	err := runUserValidatorFuncs(user, uv.bcryptPassword, uv.setRememberIfNotSet, uv.hmacRemember)
	if err != nil {
		return err
	}

	return uv.UserDB.Create(user)
}

// Update will hash a user's password and remember token.
func (uv *userValidator) Update(user *User) error {
	err := runUserValidatorFuncs(user, uv.bcryptPassword, uv.hmacRemember)
	if err != nil {
		return err
	}

	return uv.UserDB.Update(user)
}

// Delete will delete the user with the provided ID.
func (uv *userValidator) Delete(id uint) error {
	if id == 0 {
		return ErrInvalidID
	}
	return uv.UserDB.Delete(id)
}

// bcryptPassword will hash a user's password if it exists.
func (uv *userValidator) bcryptPassword(user *User) error {
	if user.Password == "" {
		return nil
	}

	pwBytes := []byte(user.Password + userPwPepper)
	hashedBytes, err := bcrypt.GenerateFromPassword(pwBytes, bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.PasswordHash = string(hashedBytes)
	user.Password = ""

	return nil
}

func (uv *userValidator) hmacRemember(user *User) error {
	if user.Remember == "" {
		return nil
	}
	user.RememberHash = uv.hmac.Hash(user.Remember)
	return nil
}

func (uv *userValidator) setRememberIfNotSet(user *User) error {
	if user.Remember != "" {
		return nil
	}

	token, err := rand.RememberToken()
	if err != nil {
		return err
	}

	user.Remember = token
	return nil
}

// Test that userGorm fulfills the UserDB interface.
var _ UserDB = &userGorm{}

type userGorm struct {
	db *gorm.DB
}

func newUserGorm(connectionInfo string) (*userGorm, error) {
	db, err := gorm.Open("postgres", connectionInfo)
	if err != nil {
		return nil, err
	}
	// TODO: remove this
	db.LogMode(true)
	return &userGorm{
		db: db,
	}, nil
}

// ByID will look up and return a user by the provided ID.
func (ug *userGorm) ByID(id uint) (*User, error) {
	var user User
	db := ug.db.Where("id = ?", id)
	err := first(db, &user)
	return &user, err
}

// ByEmail will look up and return a user by the provided email.
func (ug *userGorm) ByEmail(email string) (*User, error) {
	var user User
	db := ug.db.Where("email = ?", email)
	err := first(db, &user)
	return &user, err
}

// ByRemember will look up and return a user by their remember token;
// This method takes in a hashed remember token.
func (ug *userGorm) ByRemember(rememberHash string) (*User, error) {
	var user User
	db := ug.db.Where("remember_hash = ?", rememberHash)
	err := first(db, &user)
	return &user, err
}

// Create will add the user to the database.
func (ug *userGorm) Create(user *User) error {
	return ug.db.Create(user).Error
}

// Update will update the user with the provided user object.
func (ug *userGorm) Update(user *User) error {
	return ug.db.Save(user).Error
}

// Delete will delete the user with the provided ID.
func (ug *userGorm) Delete(id uint) error {
	user := User{Model: gorm.Model{ID: id}}
	return ug.db.Delete(&user).Error
}

// Close closes the UserService database connection.
func (ug *userGorm) Close() error {
	return ug.db.Close()
}

// DestructiveReset drops the user table and rebuilds it.
func (ug *userGorm) DestructiveReset() error {
	if err := ug.db.DropTableIfExists(&User{}).Error; err != nil {
		return err
	}
	return ug.AutoMigrate()
}

// AutoMigrate will attempt to automatically migrate the users table.
func (ug *userGorm) AutoMigrate() error {
	return ug.db.AutoMigrate(&User{}).Error
}

// first finds the first item in the query and places it into dst;
// dst should be a pointer.
func first(db *gorm.DB, dst interface{}) error {
	err := db.First(dst).Error
	if err == gorm.ErrRecordNotFound {
		return ErrNotFound
	}
	return err
}
