package app

import (
	"database/sql"
	"fmt"
)

const (
	createStmt = "insert into model (name) values (?)"
	deleteStmt = "delete from model where id = ?"
	updateStmt = "update model set name = ? where id = ?"

	findByIDQuery = "select id, name from model where id = ? limit 1"
	findAllQuery  = "select id, name from model limit ? offset ?"
)

type DAO struct {
	db *sql.DB
}

func NewDAO(db *sql.DB) *DAO {
	return &DAO{db: db}
}

func (d *DAO) Create(m *Model) (*Model, error) {
	res, err := d.db.Exec(createStmt, m.Name)
	if err != nil {
		return nil, fmt.Errorf("DAO#Create exec: %v", err)
	}

	lastID, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("DAO#Create LastInsertId: %v", err)
	}

	m.ID = lastID
	return m, nil
}

func (d *DAO) Delete(id int64) (bool, error) {
	res, err := d.db.Exec(deleteStmt, id)
	if err != nil {
		return false, fmt.Errorf("DAO#Delete Exec: %v", err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("DAO#Delete RowsAffected: %v", err)
	}

	return affected > 0, nil
}

func (d *DAO) FindAll(offset, limit int) ([]Model, error) {
	var models []Model

	rows, err := d.db.Query(findAllQuery, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("DAO#FindAll Query: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var m Model
		if err := rows.Scan(&m.ID, &m.Name); err != nil {
			return nil, fmt.Errorf("DAO#FindAll rows.Scan: %v", err)
		}
		models = append(models, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("DAO#FindAll rows iteration: %v", err)
	}

	return models, nil
}

func (d *DAO) FindByID(id int64) (*Model, error) {
	var model Model

	row := d.db.QueryRow(findByIDQuery, id)
	err := row.Scan(&model.ID, &model.Name)
	if err != nil && err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("DAO#FindByID Scan: %v", err)
	}
	return &model, nil
}

func (d *DAO) Update(m *Model) (*Model, error) {
	res, err := d.db.Exec(updateStmt, m.Name, m.ID)
	if err != nil {
		return nil, fmt.Errorf("DAO#Update Exec: %v", err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("DAO#Update RowsAffected: %v", err)
	}
	if affected < 1 {
		return nil, nil
	}

	return m, nil
}
