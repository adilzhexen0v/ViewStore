package main

import (
	"context"
	"crypto/tls"
	"finalexam.project/internal/models"
	"flag"
	"github.com/alexedwards/scs/pgxstore"
	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/form/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"
)

var infoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
var errorLog = log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

type application struct {
	errorLog       *log.Logger
	infoLog        *log.Logger
	users          *models.UserModel
	images         *models.ImageModel
	posts          *models.PostModel
	templateCache  map[string]*template.Template
	formDecoder    *form.Decoder
	sessionManager *scs.SessionManager
}

func main() {
	addr := flag.String("addr", ":4000", "HTTP network address")
	flag.Parse()

	dbURL := "postgres://postgres:123@localhost:5432/finalexam"
	db, err1 := openDB(dbURL)
	if err1 != nil {
		errorLog.Fatal(err1)
	}
	defer db.Close()

	templateCache, err2 := newTemplateCache()
	if err2 != nil {
		errorLog.Fatal(err2)
	}
	formDecoder := form.NewDecoder()
	sessionManager := scs.New()
	sessionManager.Store = pgxstore.New(db)
	sessionManager.Lifetime = 12 * time.Hour

	app := &application{
		errorLog:       errorLog,
		infoLog:        infoLog,
		users:          &models.UserModel{DB: db},
		images:         &models.ImageModel{DB: db},
		posts:          &models.PostModel{DB: db},
		templateCache:  templateCache,
		formDecoder:    formDecoder,
		sessionManager: sessionManager,
	}

	tlsConfig := &tls.Config{
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
	}

	srv := &http.Server{
		Addr:           *addr,
		MaxHeaderBytes: 524288,
		ErrorLog:       errorLog,
		Handler:        app.routes(),
		TLSConfig:      tlsConfig,
		IdleTimeout:    time.Minute,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   10 * time.Second,
	}

	infoLog.Printf("Starting server on %s", *addr)
	err := srv.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")
	errorLog.Fatal(err)

}

func openDB(dsn string) (*pgxpool.Pool, error) {
	pool, err1 := pgxpool.Connect(context.Background(), dsn)
	infoLog.Printf("Connected!")
	return pool, err1
}
