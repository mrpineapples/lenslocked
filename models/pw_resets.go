package models

import (
	"github.com/jinzhu/gorm"
	"github.com/mrpineapples/lenslocked/hash"
	"github.com/mrpineapples/lenslocked/rand"
)

type pwReset struct {
	gorm.Model
	UserID    uint   `gorm:"not null`
	Token     string `gorm:"-"`
	TokenHash string `gorm:"not null;unique_index"`
}

type pwResetDB interface {
	ByToken(token string) (*pwReset, error)
	Create(pwr *pwReset) error
	Delete(id uint) error
}

func newPwResetValidator(db pwResetDB, hmac hash.HMAC) *pwResetValidator {
	return &pwResetValidator{
		pwResetDB: db,
		hmac:      hmac,
	}
}

type pwResetValidatorFunc func(*pwReset) error

func runPwResetValidatorFuncs(pwr *pwReset, fns ...pwResetValidatorFunc) error {
	for _, fn := range fns {
		if err := fn(pwr); err != nil {
			return err
		}
	}
	return nil
}

type pwResetValidator struct {
	pwResetDB
	hmac hash.HMAC
}

func (pwrv *pwResetValidator) ByToken(token string) (*pwReset, error) {
	pwr := pwReset{Token: token}
	err := runPwResetValidatorFuncs(&pwr, pwrv.hmacToken)
	if err != nil {
		return nil, err
	}
	return pwrv.pwResetDB.ByToken(pwr.TokenHash)
}

func (pwrv *pwResetValidator) Create(pwr *pwReset) error {
	err := runPwResetValidatorFuncs(pwr,
		pwrv.requireUserID,
		pwrv.setTokenIfNotSet,
		pwrv.hmacToken,
	)
	if err != nil {
		return err
	}
	return pwrv.pwResetDB.Create(pwr)
}

func (pwrv *pwResetValidator) Delete(id uint) error {
	if id <= 0 {
		return ErrIDInvalid
	}

	return pwrv.pwResetDB.Delete(id)
}

func (pwrv *pwResetValidator) requireUserID(pwr *pwReset) error {
	if pwr.UserID <= 0 {
		return ErrUserIDRequired
	}
	return nil
}

func (pwrv *pwResetValidator) setTokenIfNotSet(pwr *pwReset) error {
	if pwr.Token != "" {
		return nil
	}

	// using remember token for now but could probably use something shorter
	token, err := rand.RememberToken()
	if err != nil {
		return err
	}
	pwr.Token = token
	return nil
}

func (pwrv *pwResetValidator) hmacToken(pwr *pwReset) error {
	if pwr.Token == "" {
		return nil
	}
	pwr.TokenHash = pwrv.hmac.Hash(pwr.Token)
	return nil
}

type pwResetGorm struct {
	db *gorm.DB
}

func (pwrg *pwResetGorm) ByToken(tokenHash string) (*pwReset, error) {
	var pwr pwReset
	err := first(pwrg.db.Where("token_hash = ?", tokenHash), &pwr)
	if err != nil {
		return nil, err
	}
	return &pwr, nil
}

func (pwrg *pwResetGorm) Create(pwr *pwReset) error {
	return pwrg.db.Create(pwr).Error
}

func (pwrg *pwResetGorm) Delete(id uint) error {
	pwr := pwReset{Model: gorm.Model{ID: id}}
	return pwrg.db.Delete(&pwr).Error
}
