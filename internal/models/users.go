package models

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type User struct {
	ID             int
	Name           string
	Surname        string
	Email          string
	HashedPassword []byte
	Created        time.Time
	ProfilePicture string
}

type UserModel struct {
	DB *pgxpool.Pool
}

func (m *UserModel) Insert(name, surname, email, password string) error {
	conn, err := m.DB.Acquire(context.Background())
	if err != nil {
		fmt.Printf("Unable to acquire a database connection: %v\n", err)
		return err
	}
	defer conn.Release()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}
	stmt := `INSERT INTO users (name, surname, email, hashed_password, created, profile_picture)
	VALUES($1, $2, $3, $4, current_timestamp, 'default')`

	conn.QueryRow(context.Background(),
		stmt,
		name, surname, email, string(hashedPassword))
	if err != nil {
		fmt.Printf("Unable to INSERT: %v\n", err)
		return err
	}
	return nil
}

func (sm *UserModel) Get(userId int) (*User, error) {
	conn, err := sm.DB.Acquire(context.Background())
	if err != nil {
		fmt.Printf("Unable to acquire a database connection: %v\n", err)
		return nil, nil
	}
	defer conn.Release()
	row := conn.QueryRow(context.Background(),
		"SELECT id, name, surname, email, hashed_password, created, profile_picture FROM users WHERE id = $1",
		userId)
	u := &User{}
	err = row.Scan(&u.ID, &u.Name, &u.Surname, &u.Email, &u.HashedPassword, &u.Created, &u.ProfilePicture)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNoRecord
		} else {
			return nil, err
		}
	}
	return u, err
}

func (m *UserModel) Authenticate(email, password string) (int, error) {
	var id int
	var hashedPassword []byte
	stmt := `SELECT id, hashed_password FROM users WHERE email = $1`
	row := m.DB.QueryRow(context.Background(), stmt, email)
	err := row.Scan(&id, &hashedPassword)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrInvalidCredentials
		} else {
			return 0, err
		}
	}

	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return 0, ErrInvalidCredentials
		} else {
			return 0, err
		}
	}
	return id, nil

}

func (m *UserModel) Exists(id int) (bool, error) {
	return false, nil
}

func (m *UserModel) Update(name, surname, email, picture string, userId int) error {
	conn, err := m.DB.Acquire(context.Background())
	if err != nil {
		fmt.Printf("Unable to acquire a database connection: %v\n", err)
		return nil
	}
	defer conn.Release()
	conn.Exec(context.Background(),
		"UPDATE users SET name = $1, surname = $2, email = $3, profile_picture = $4 WHERE id = $5",
		name, surname, email, picture, userId,
	)

	return err
}

func (m *UserModel) GetAllSubscriptions(id int) ([]*User, error) {
	conn, err := m.DB.Acquire(context.Background())
	if err != nil {
		fmt.Printf("Unable to acquire a database connection: %v\n", err)
		return nil, nil
	}
	defer conn.Release()
	rows, err := conn.Query(context.Background(),
		"SELECT id, name, surname, email, hashed_password, created, profile_picture FROM users WHERE id IN (SELECT sub_id FROM subs WHERE user_id = $1)", id,
	)
	subs := []*User{}
	for rows.Next() {
		u := &User{}
		err = rows.Scan(&u.ID, &u.Name, &u.Surname, &u.Email, &u.HashedPassword, &u.Created, &u.ProfilePicture)
		if err != nil {
			return nil, err
		}

		subs = append(subs, u)
	}
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNoRecord
		} else {
			return nil, err
		}
	}
	return subs, nil
}

func (m *UserModel) GetAllSubscribers(id int) ([]*User, error) {
	conn, err := m.DB.Acquire(context.Background())
	if err != nil {
		fmt.Printf("Unable to acquire a database connection: %v\n", err)
		return nil, nil
	}
	defer conn.Release()
	rows, err := conn.Query(context.Background(),
		"SELECT id, name, surname, email, hashed_password, created, profile_picture FROM users WHERE id IN (SELECT user_id FROM subs WHERE sub_id = $1)", id,
	)
	subs := []*User{}
	for rows.Next() {
		u := &User{}
		err = rows.Scan(&u.ID, &u.Name, &u.Surname, &u.Email, &u.HashedPassword, &u.Created, &u.ProfilePicture)
		if err != nil {
			return nil, err
		}

		subs = append(subs, u)
	}
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNoRecord
		} else {
			return nil, err
		}
	}
	return subs, nil
}

func (m *UserModel) GetAllNonSubscriptions(id int) ([]*User, error) {
	conn, err := m.DB.Acquire(context.Background())
	if err != nil {
		fmt.Printf("Unable to acquire a database connection: %v\n", err)
		return nil, nil
	}
	defer conn.Release()
	rows, err := conn.Query(context.Background(),
		"SELECT id, name, surname, email, hashed_password, created, profile_picture FROM users WHERE id NOT IN (SELECT sub_id FROM subs WHERE user_id = $1) AND id != $1", id,
	)
	subs := []*User{}
	for rows.Next() {
		u := &User{}
		err = rows.Scan(&u.ID, &u.Name, &u.Surname, &u.Email, &u.HashedPassword, &u.Created, &u.ProfilePicture)
		if err != nil {
			return nil, err
		}

		subs = append(subs, u)
	}
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNoRecord
		} else {
			return nil, err
		}
	}
	return subs, nil
}

func (m *UserModel) Subscribe(userId, subId int) error {
	conn, err := m.DB.Acquire(context.Background())
	if err != nil {
		fmt.Printf("Unable to acquire a database connection: %v\n", err)
		return nil
	}
	defer conn.Release()
	conn.Exec(context.Background(),
		"INSERT INTO subs (user_id, sub_id) VALUES($1, $2)",
		userId, subId,
	)
	return err
}

func (m *UserModel) Unsubscribe(userId, subId int) error {
	conn, err := m.DB.Acquire(context.Background())
	if err != nil {
		fmt.Printf("Unable to acquire a database connection: %v\n", err)
		return nil
	}
	defer conn.Release()
	conn.Exec(context.Background(),
		"DELETE FROM subs WHERE user_id = $1 AND sub_id = $2",
		userId, subId,
	)
	return err
}
