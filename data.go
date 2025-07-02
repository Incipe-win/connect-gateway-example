package main

import (
	"context"
	"errors"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const defaultShelfSize = 100

// 使用 GORM
func NewDB(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database: " + err.Error())
	}
	db.AutoMigrate(&Shelf{}, &Book{})
	return db, nil
}

// 定义 Model

// Shelf 书架
type Shelf struct {
	ID       int64 `gorm:"primaryKey"`
	Theme    string
	Size     int64
	CreateAt time.Time
	UpdateAt time.Time
}

type Book struct {
	ID       int64 `gorm:"primaryKey"`
	Author   string
	Title    string
	ShelfID  int64
	CreateAt time.Time
	UpdateAt time.Time
}

// 数据库操作
type bookstore struct {
	db *gorm.DB
}

func (b *bookstore) CreateShelf(ctx context.Context, data Shelf) (*Shelf, error) {
	if len(data.Theme) <= 0 {
		return nil, errors.New("invalid theme")
	}
	size := data.Size
	if size <= 0 {
		size = defaultShelfSize
	}
	v := Shelf{
		Theme:    data.Theme,
		Size:     size,
		CreateAt: time.Now(),
		UpdateAt: time.Now(),
	}
	err := b.db.WithContext(ctx).Create(&v).Error
	return &v, err
}

func (b *bookstore) GetShelf(ctx context.Context, id int64) (*Shelf, error) {
	v := Shelf{}
	err := b.db.WithContext(ctx).First(&v, id).Error
	return &v, err
}

func (b *bookstore) ListShelves(ctx context.Context) ([]*Shelf, error) {
	var vl []*Shelf
	err := b.db.WithContext(ctx).Find(&vl).Error
	return vl, err
}

func (b *bookstore) DeleteShelf(ctx context.Context, id int64) error {
	err := b.db.WithContext(ctx).Delete(&Shelf{}, id).Error
	return err
}

func (b *bookstore) GetBookListByShelfID(ctx context.Context, shelfID int64, cursor string, pageSize int) ([]*Book, error) {
	var vl []*Book
	err := b.db.WithContext(ctx).Where("shelf_id = ? and id > ?", shelfID, cursor).Order("id asc").Limit(pageSize).Find(&vl).Error
	return vl, err
}

func (b *bookstore) CreateBook(ctx context.Context, data *Book) (*Book, error) {
	if data.ShelfID <= 0 {
		return nil, errors.New("invalid shelf ID")
	}
	if len(data.Author) <= 0 || len(data.Title) <= 0 {
		return nil, errors.New("invalid author or title")
	}
	d := &Book{
		Author:  data.Author,
		Title:   data.Title,
		ShelfID: data.ShelfID,
	}
	err := b.db.WithContext(ctx).Create(d).Error
	return d, err
}
