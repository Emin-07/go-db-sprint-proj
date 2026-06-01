package main

import (
	"context"
	"database/sql"
	"errors"
)

type ParcelStore struct {
	db *sql.DB
}

func NewParcelStore(db *sql.DB) ParcelStore {
	return ParcelStore{db: db}
}

const addParcelQuery = `INSERT INTO parcel(client, status, address, created_at) 
                      VALUES(:client, :status, :address, :created_at)`

func (s ParcelStore) Add(ctx context.Context, p Parcel) (int, error) {
	// реализуйте добавление строки в таблицу parcel, используйте данные из переменной p
	res, err := s.db.ExecContext(ctx, addParcelQuery,
		sql.Named("client", p.Client),
		sql.Named("status", p.Status),
		sql.Named("address", p.Address),
		sql.Named("created_at", p.CreatedAt),
	)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	// верните идентификатор последней добавленной записи
	return int(id), nil
}

const getParcelQuery = `SELECT number, client, status, address, created_at 
                      FROM parcel WHERE number = :number`

func (s ParcelStore) Get(ctx context.Context, number int) (Parcel, error) {
	// реализуйте чтение строки по заданному number
	// здесь из таблицы должна вернуться только одна строка
	row := s.db.QueryRowContext(ctx, getParcelQuery, sql.Named("number", number))

	// заполните объект Parcel данными из таблицы
	p := Parcel{}
	err := row.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
	if err != nil {
		return Parcel{}, err
	}

	return p, nil
}

const getParcelsByClientQuery = `SELECT number, client, status, address, created_at
                               FROM parcel WHERE client = :client`

func (s ParcelStore) GetByClient(ctx context.Context, client int) ([]Parcel, error) {
	// реализуйте чтение строк из таблицы parcel по заданному client
	// здесь из таблицы может вернуться несколько строк
	rows, err := s.db.QueryContext(ctx, getParcelsByClientQuery, sql.Named("client", client))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []Parcel{}, nil
		}
		return nil, err
	}
	defer rows.Close()
	// заполните срез Parcel данными из таблицы
	var res []Parcel
	for rows.Next() {
		p := Parcel{}
		err := rows.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
		if err != nil {
			return nil, err
		}
		res = append(res, p)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	if res == nil {
		return []Parcel{}, nil
	}

	return res, nil
}

const updateParcelStatusQuery = `UPDATE parcel SET status = :status WHERE number = :number`

func (s ParcelStore) SetStatus(ctx context.Context, number int, status string) error {
	// реализуйте обновление статуса в таблице parcel
	_, err := s.db.ExecContext(ctx, updateParcelStatusQuery, sql.Named("status", status), sql.Named("number", number))
	if err != nil {
		return err
	}

	return nil
}

const updateParcelAddressQuery = `UPDATE parcel SET address = :address 
                                WHERE number = :number AND status = :status`

func (s ParcelStore) SetAddress(ctx context.Context, number int, address string) error {
	// реализуйте обновление адреса в таблице parcel
	// менять адрес можно только если значение статуса registered
	_, err := s.db.ExecContext(ctx, updateParcelAddressQuery,
		sql.Named("address", address),
		sql.Named("number", number),
		sql.Named("status", ParcelStatusRegistered),
	)
	if err != nil {
		return err
	}

	return nil
}

const deleteParcelQuery = `DELETE FROM parcel 
                         WHERE number = :number AND status = :status`

func (s ParcelStore) Delete(ctx context.Context, number int) error {
	// реализуйте удаление строки из таблицы parcel
	// удалять строку можно только если значение статуса registered
	_, err := s.db.ExecContext(ctx, deleteParcelQuery, sql.Named("number", number), sql.Named("status", ParcelStatusRegistered))
	if err != nil {
		return err
	}

	return nil
}
