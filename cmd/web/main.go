package main

import (
	"crypto/tls"
	"database/sql"
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/alexedwards/scs/mysqlstore"
	"github.com/alexedwards/scs/v2"
	"github.com/devel1x/snippetbox/internal/models"

	"github.com/go-playground/form/v4"
	_ "github.com/go-sql-driver/mysql"
)

type application struct {
	errorLog      *log.Logger
	infoLog       *log.Logger
	snippets      *models.SnippetModel
	users         *models.UserModel
	templateCache map[string]*template.Template
	formDecoder   *form.Decoder
	sessionManager *scs.SessionManager
}

func main() {
	addr := flag.String("addr", ":4000", "HTTP network addres")
	dsn := flag.String("dsn", "web:1488@/snippetbox?parseTime=true", "MySQL data source name")

	flag.Parse()

	infoLog := log.New(os.Stdout, "\u001b[32mINFO\u001b[0m\t", log.Ldate|log.Ltime)
	errLog := log.New(os.Stderr, "\u001b[31mERROR\u001b[0m\t", log.Ldate|log.Ltime|log.Lshortfile)

	db, err := openDB((*dsn))
	if err != nil {
		errLog.Fatal(err)
	}
	defer db.Close()

	templateCache, err := newTemplateCache()
	if err != nil {
		errLog.Fatal(err)
	}

	formDecoder := form.NewDecoder()

	sessionManager := scs.New()
	sessionManager.Store = mysqlstore.New(db)
	sessionManager.Lifetime = 12 * time.Hour
	sessionManager.Cookie.Secure = true

	tlsConfig := &tls.Config{
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
	}

	app := &application{
		errorLog:      errLog,
		infoLog:       infoLog,
		snippets:      &models.SnippetModel{DB: db},
		users:         &models.UserModel{DB: db},
		templateCache: templateCache,
		formDecoder:   formDecoder,
		sessionManager: sessionManager,
	}

	srv := &http.Server{
		Addr:         *addr,
		ErrorLog:     errLog,
		Handler:      app.routes(),
		TLSConfig:    tlsConfig,
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	infoLog.Printf("Starting server on https://localhost%s", *addr)
	err = srv.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")
	errLog.Fatal(err)
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
