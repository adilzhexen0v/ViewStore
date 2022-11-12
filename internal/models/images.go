package models

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type Image struct {
	ID     int
	Link   string
	PostId int
}

type ImageModel struct {
	DB *pgxpool.Pool
}

func (i *ImageModel) GetAllMyImages(postId int) ([]*Image, error) {
	conn, err := i.DB.Acquire(context.Background())
	if err != nil {
		fmt.Printf("Unable to acquire a database connection: %v\n", err)
		return nil, nil
	}
	defer conn.Release()
	rows, err := conn.Query(context.Background(),
		"SELECT id, link, post_id FROM images WHERE post_id = $1", postId,
	)
	images := []*Image{}
	for rows.Next() {
		img := &Image{}
		err = rows.Scan(&img.ID, &img.Link, &img.PostId)
		if err != nil {
			return nil, err
		}
		images = append(images, img)
	}
	if err != nil {

		if errors.Is(err, pgx.ErrNoRows) {
			fmt.Println("ErrNoRecord from snippets.go")
			return nil, ErrNoRecord
		} else {
			fmt.Println("Another error from snippets.go")

			return nil, err
		}
	}
	return images, nil
}

func (i *ImageModel) Insert(id int, link string, authorId int) error {
	conn, err := i.DB.Acquire(context.Background())
	if err != nil {
		fmt.Printf("Unable to acquire a database connection: %v\n", err)
		return err
	}
	defer conn.Release()

	if err != nil {
		return err
	}
	stmt := `INSERT INTO images (link, post_id, author_id)
	VALUES($1, $2, $3)`

	conn.QueryRow(context.Background(),
		stmt,
		link, id, authorId,
	)
	if err != nil {
		fmt.Printf("Unable to INSERT: %v\n", err)
		return err
	}
	return nil
}
