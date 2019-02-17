package app

import (
	"fmt"

	"github.com/jinzhu/gorm"
)

type DAO interface {
	Create(m *Model) (*Model, error)
	Delete(id uint) (bool, error)
	FindAll(offset, limit int) ([]Model, error)
	FindByID(id uint) (*Model, error)
	Update(m *Model) (*Model, error)
}

type dao struct {
	db *gorm.DB
}

func NewDAO(db *gorm.DB) DAO {
	return &dao{db: db}
}

// Createはmの値で新規レコードをDBに保存する。
// IDを含んだModelのインスタンスを返却する。
func (d *dao) Create(m *Model) (*Model, error) {
	res := d.db.Create(m)
	if res.Error != nil {
		return nil, fmt.Errorf("dao#Create exec: %v", res.Error)
	}
	return m, nil
}

// Deleteはidに指定されるエンティティを削除する。成功する場合はtrueを返却する。
func (d *dao) Delete(id uint) (bool, error) {
	row := Model{Model: gorm.Model{ID: id}}
	res := d.db.Delete(row)
	if res.Error != nil {
		return false, fmt.Errorf("dao#Delete Exec: %v", res.Error)
	}
	return res.RowsAffected > 0, nil
}

// FindAllは複数のModelを返却する。offset（開始位置）と
// limit（行数）でページネーションを行う。
func (d *dao) FindAll(offset, limit int) ([]Model, error) {
	var rows []Model
	res := d.db.Limit(limit).Offset(offset).Find(&rows)
	if res.Error != nil {
		return nil, fmt.Errorf("dao#FindAll Query: %v", res.Error)
	}
	return rows, nil
}

// FindByIDはIDに指定されたModelを返却する。
// 存在しない場合はnilを返却する。
func (d *dao) FindByID(id uint) (*Model, error) {
	row := &Model{Model: gorm.Model{ID: id}}
	res := d.db.First(row)
	if res.Error != nil {
		return nil, fmt.Errorf("dao#FindByID Scan: %v", res.Error)
	}
	return row, nil
}

// UpdateはmのDB上のレコードを更新する。
// m.IDのModelが存在しない、または変更点がない場合はnilを返却する。
func (d *dao) Update(m *Model) (*Model, error) {
	res := d.db.Model(m).Update("name", m.Name)
	if res.Error != nil {
		return nil, fmt.Errorf("dao#Update Exec: %v", res.Error)
	}
	if res.RowsAffected < 1 {
		return nil, nil // TODO 2019-02-17 @Jimeux カスタムエラー
	}
	return m, nil
}
