package app

import (
	"database/sql"

	"golang.org/x/xerrors"
)

const (
	createStmt = "insert into model (name) values (?)"
	deleteStmt = "delete from model where id = ?"
	updateStmt = "update model set name = ? where id = ?"

	findByIDQuery = "select id, name from model where id = ? limit 1"
	findAllQuery  = "select id, name from model limit ? offset ?"
)

type DAO struct {
	db         *sql.DB
	encryption *EncryptionManager
}

func NewDAO(db *sql.DB, encryption *EncryptionManager) *DAO {
	return &DAO{db: db, encryption: encryption}
}

// Createはmの値で新規レコードをDBに保存する。
// IDを含んだModelのインスタンスを返却する。
func (d *DAO) Create(m *Model) (*Model, error) {
	if err := d.encryption.Encrypt(m); err != nil {
		return nil, xerrors.Errorf("could not encrypt pre-create: %w", err)
	}

	res, err := d.db.Exec(createStmt, m.Name)
	if err != nil {
		return nil, xerrors.Errorf("error at exec: %w", err)
	}

	lastID, err := res.LastInsertId()
	if err != nil {
		return nil, xerrors.Errorf("error at LastInsertId: %w", err)
	}

	m.ID = lastID
	if err := d.encryption.Decrypt(m); err != nil {
		return nil, xerrors.Errorf("could not decrypt ID %d post-create: %w", lastID, err)
	}
	return m, nil
}

// Deleteはidに指定されるエンティティを削除する。成功する場合はtrueを返却する。
func (d *DAO) Delete(id int64) (bool, error) {
	res, err := d.db.Exec(deleteStmt, id)
	if err != nil {
		return false, xerrors.Errorf("error at Exec: %w", err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return false, xerrors.Errorf("error at RowsAffected: %w", err)
	}

	return affected > 0, nil
}

// FindAllは複数のModelを返却する。offset（開始位置）と
// limit（行数）でページネーションを行う。
func (d *DAO) FindAll(offset, limit int) ([]Model, error) {
	var models []Model

	rows, err := d.db.Query(findAllQuery, limit, offset)
	if err != nil {
		return nil, xerrors.Errorf("error at Query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var m Model
		if err := rows.Scan(&m.ID, &m.Name); err != nil {
			return nil, xerrors.Errorf("error at rows.Scan: %w", err)
		}

		// 復号化
		if err := d.encryption.Decrypt(&m); err != nil {
			return nil, xerrors.Errorf("row decryption error: %w", err)
		}

		models = append(models, m)
	}
	if err := rows.Err(); err != nil {
		return nil, xerrors.Errorf("row iteration error: %w", err)
	}

	return models, nil
}

// FindByIDはIDに指定されたModelを返却する。
// 存在しない場合はnilを返却する。
func (d *DAO) FindByID(id int64) (*Model, error) {
	var model Model

	row := d.db.QueryRow(findByIDQuery, id)
	err := row.Scan(&model.ID, &model.Name)
	if err != nil && err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, xerrors.Errorf("error at Scan: %w", err)
	}

	// 復号化
	if err := d.encryption.Decrypt(&model); err != nil {
		return nil, err
	}

	return &model, nil
}

// UpdateはmのDB上のレコードを更新する。
// m.IDのModelが存在しない、または変更点がない場合はnilを返却する。
func (d *DAO) Update(m *Model) (*Model, error) {
	if err := d.encryption.Encrypt(m); err != nil {
		return nil, xerrors.Errorf("could not encrypt pre-update: %w", err)
	}

	res, err := d.db.Exec(updateStmt, m.Name, m.ID)
	if err != nil {
		return nil, xerrors.Errorf("error at Exec: %w", err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return nil, xerrors.Errorf("error at RowsAffected: %w", err)
	}
	if affected < 1 {
		return nil, nil
	}

	if err := d.encryption.Decrypt(m); err != nil {
		return nil, err
	}

	return m, nil
}
