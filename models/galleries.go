package models

import "github.com/jinzhu/gorm"

// Gallery represents a user's collection of images.
type Gallery struct {
	gorm.Model
	UserID uint   `gorm:"not null;index"`
	Title  string `gorm:"not null"`
}
