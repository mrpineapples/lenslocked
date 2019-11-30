package models

import "github.com/jinzhu/gorm"

// Gallery represents a user's collection of images.
type Gallery struct {
	gorm.Model
	UserID uint   `gorm:"not null;index"`
	Title  string `gorm:"not null"`
}

type GalleryService interface {
	GalleryDB
}

type GalleryDB interface {
	Create(gallery *Gallery) error
}

func NewGalleryService(db *gorm.DB) GalleryService {
	return &galleryService{
		GalleryDB: &galleryValidator{
			&galleryGorm{db},
		},
	}
}

type galleryService struct {
	GalleryDB
}

type galleryValidatorFunc func(*Gallery) error

func runGalleryValidatorFuncs(gallery *Gallery, fns ...galleryValidatorFunc) error {
	for _, fn := range fns {
		if err := fn(gallery); err != nil {
			return err
		}
	}
	return nil
}

type galleryValidator struct {
	GalleryDB
}

func (gv *galleryValidator) Create(gallery *Gallery) error {
	err := runGalleryValidatorFuncs(gallery,
		gv.userIDRequired,
		gv.titleRequired,
	)
	if err != nil {
		return err
	}

	return gv.GalleryDB.Create(gallery)
}

func (gv *galleryValidator) userIDRequired(gallery *Gallery) error {
	if gallery.UserID <= 0 {
		return ErrUserIDRequired
	}
	return nil
}

func (gv *galleryValidator) titleRequired(gallery *Gallery) error {
	if gallery.Title == "" {
		return ErrTitleRequired
	}
	return nil
}

var _ GalleryDB = &galleryGorm{}

type galleryGorm struct {
	db *gorm.DB
}

func (gg *galleryGorm) Create(gallery *Gallery) error {
	return gg.db.Create(gallery).Error
}
