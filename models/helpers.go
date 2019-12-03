package models

import "github.com/jinzhu/gorm"

// first finds the first item in the query and places it into dst;
// dst should be a pointer.
func first(db *gorm.DB, dst interface{}) error {
	err := db.First(dst).Error
	if err == gorm.ErrRecordNotFound {
		return ErrNotFound
	}
	return err
}
