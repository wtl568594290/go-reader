package dao

import (
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Book struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	Title     string `gorm:"unique;not null"`
	Length    int    `gorm:"not null"`
	LastPos   int    `gorm:"not null;default:0"`
}

var db *gorm.DB

func init() {
	var err error
	db, err = gorm.Open(sqlite.Open("data.db"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		panic("failed to connect database")
	}
	db.AutoMigrate(&Book{})
}

func CreateBook(title string, length int) (book Book, err error) {
	err = db.Create(&Book{Title: title, Length: length}).Error
	return
}
func UpdateBookPos(title string, pos int) error {
	return db.Model(&Book{}).Where("title = ?", title).Update("last_pos", pos).Error
}

func GetBooks() (books []Book) {
	if err := db.Find(&books).Error; err != nil {
		return []Book{}
	} else {
		return books
	}
}

func GetBookByName(name string) (book Book, err error) {
	err = db.Where("title = ?", name).First(&book).Error
	return
}

func DeleteBook(id uint) error {
	return db.Delete(&Book{}, id).Error
}

func DeleteBookByName(name string) error {
	return db.Where("title = ?", name).Delete(&Book{}).Error
}
