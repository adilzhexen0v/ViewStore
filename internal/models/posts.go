package models

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"time"
)

type Post struct {
	ID            int
	Content       string
	AuthorID      int
	Created       time.Time
	AuthorId      int
	AuthorName    string
	AuthorSurname string
	AuthorPicture string
	ImagesLength  string
	Images        []*Image
}

type PostModel struct {
	DB *pgxpool.Pool
}

func (p *PostModel) GetAllPosts(userId int) ([]*Post, error) {
	conn, err := p.DB.Acquire(context.Background())
	if err != nil {
		fmt.Printf("Unable to acquire a database connection: %v\n", err)
		return nil, nil
	}
	defer conn.Release()
	rows, err := conn.Query(context.Background(),
		"SELECT p.id, p.content, p.author_id, p.created, u.id, u.name, u.surname, u.profile_picture FROM posts p JOIN users u ON p.author_id=u.id WHERE p.author_id IN (SELECT sub_id FROM subs WHERE user_id = $1) ORDER BY p.created DESC", userId,
	)
	posts := []*Post{}
	for rows.Next() {
		post := &Post{}
		err = rows.Scan(&post.ID, &post.Content, &post.AuthorID, &post.Created, &post.AuthorId, &post.AuthorName, &post.AuthorSurname, &post.AuthorPicture)
		if err != nil {
			return nil, err
		}
		images, err := p.GetAllMyImages(post.ID)
		if err != nil {
			fmt.Println(err)
		}
		post.Images = images
		if len(images) == 3 {
			post.ImagesLength = "three"
		} else if len(images) == 2 {
			post.ImagesLength = "two"
		} else {
			post.ImagesLength = "other"
		}
		posts = append(posts, post)
	}

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNoRecord
		} else {
			return nil, err
		}
	}
	return posts, err
}

func (p *PostModel) GetAllMyPosts(authorId int) ([]*Post, error) {
	conn, err := p.DB.Acquire(context.Background())
	if err != nil {
		fmt.Printf("Unable to acquire a database connection: %v\n", err)
		return nil, nil
	}
	defer conn.Release()
	rows, err := conn.Query(context.Background(),
		"SELECT p.id, p.content, p.author_id, p.created, u.id, u.name, u.surname, u.profile_picture FROM posts p JOIN users u ON p.author_id=u.id WHERE author_id = $1 ORDER BY p.created DESC", authorId,
	)
	posts := []*Post{}
	for rows.Next() {
		post := &Post{}
		err = rows.Scan(&post.ID, &post.Content, &post.AuthorID, &post.Created, &post.AuthorId, &post.AuthorName, &post.AuthorSurname, &post.AuthorPicture)
		if err != nil {
			return nil, err
		}
		images, err := p.GetAllMyImages(post.ID)
		if err != nil {
			fmt.Println(err)
		}
		post.Images = images
		if len(images) == 3 {
			post.ImagesLength = "three"
		} else if len(images) == 2 {
			post.ImagesLength = "two"
		} else {
			post.ImagesLength = "other"
		}
		posts = append(posts, post)
	}

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNoRecord
		} else {
			return nil, err
		}
	}
	return posts, err
}

func (p *PostModel) GetAllMyImages(id int) ([]*Image, error) {
	conn, err := p.DB.Acquire(context.Background())
	if err != nil {
		fmt.Printf("Unable to acquire a database connection: %v\n", err)
		return nil, nil
	}
	defer conn.Release()
	rows, err2 := conn.Query(context.Background(),
		"select i.id, link, post_id from posts p join images i on p.id = i.post_id where p.id = $1", id,
	)
	if err2 != nil {
		fmt.Println("no images")
		fmt.Println(err2)
	}
	imgs := []*Image{}

	for rows.Next() {
		i := &Image{}
		err = rows.Scan(&i.ID, &i.Link, &i.PostId)
		if err != nil {
			fmt.Println(err)
		}
		/*
			filepath := fmt.Sprintf("%v.jpg", i.Link)
			//newFilepath := fmt.Sprintf("./static/files/%v.jpg", i.Link)
			file, err := os.Create(filepath)
			if err != nil {
				log.Fatal(err)
			}
			defer file.Close()

			url := fmt.Sprintf("https://drive.google.com/uc?export=download&id=%v", i.Link)

			r, err := http.Get(url)
			if err != nil {
				log.Fatal(err)
			}
			defer r.Body.Close()

			_, err = io.Copy(file, r.Body)
			if err != nil {
				log.Fatal(err)
			}
		*/

		imgs = append(imgs, i)
	}
	return imgs, nil
}

func (p *PostModel) Insert(content string, authorId int) (int, error) {
	conn, err := p.DB.Acquire(context.Background())
	if err != nil {
		fmt.Printf("Unable to acquire a database connection: %v\n", err)
		return 0, err
	}
	defer conn.Release()

	if err != nil {
		return 0, err
	}
	stmt := `INSERT INTO posts (content, author_id, created)
	VALUES($1, $2, current_timestamp) RETURNING id`

	row := conn.QueryRow(context.Background(),
		stmt,
		content, authorId,
	)
	var postId int
	err = row.Scan(&postId)
	if err != nil {
		fmt.Printf("Unable to INSERT: %v\n", err)
		return 0, err
	}
	return postId, nil
}
