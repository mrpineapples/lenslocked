package models

import (
	"github.com/jinzhu/gorm"
	"golang.org/x/oauth2"
)

const (
	OAuthDropbox = "dropbox"
)

type OAuth struct {
	gorm.Model
	UserID  uint   `gorm:"not null;unique_index:user_id_service"`
	Service string `gorm:"not null;unique_index:user_id_service"`
	oauth2.Token
}

type OAuthService interface {
	OAuthDB
}

type OAuthDB interface {
	Find(userID uint, service string) (*OAuth, error)
	Create(oauth *OAuth) error
	Delete(id uint) error
}

func NewOAuthService(db *gorm.DB) OAuthService {
	return &oauthValidator{&oauthGorm{db}}
}

type oauthValidatorFunc func(*OAuth) error

func runOAuthValidatorFuncs(oauth *OAuth, fns ...oauthValidatorFunc) error {
	for _, fn := range fns {
		if err := fn(oauth); err != nil {
			return err
		}
	}
	return nil
}

type oauthValidator struct {
	OAuthDB
}

func (ov *oauthValidator) Create(oauth *OAuth) error {
	err := runOAuthValidatorFuncs(oauth,
		ov.userIDRequired,
		ov.serviceRequired,
	)
	if err != nil {
		return err
	}

	return ov.OAuthDB.Create(oauth)
}

func (ov *oauthValidator) Delete(id uint) error {
	if id <= 0 {
		return ErrIDInvalid
	}
	return ov.OAuthDB.Delete(id)
}

func (ov *oauthValidator) userIDRequired(oauth *OAuth) error {
	if oauth.UserID <= 0 {
		return ErrUserIDRequired
	}
	return nil
}

func (ov *oauthValidator) serviceRequired(oauth *OAuth) error {
	if oauth.Service == "" {
		return ErrServiceRequired
	}
	return nil
}

var _ OAuthDB = &oauthGorm{}

type oauthGorm struct {
	db *gorm.DB
}

func (og *oauthGorm) Find(userID uint, service string) (*OAuth, error) {
	var oauth OAuth
	db := og.db.Where("user_id = ?", userID).Where("service = ?", service)
	err := first(db, &oauth)
	return &oauth, err
}

func (og *oauthGorm) Create(oauth *OAuth) error {
	return og.db.Create(oauth).Error
}

func (og *oauthGorm) Delete(id uint) error {
	oauth := OAuth{Model: gorm.Model{ID: id}}
	// "unscoped" delete to do a hard delete of oauth tokens
	return og.db.Unscoped().Delete(&oauth).Error
}
