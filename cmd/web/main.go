package main

import (
	"flag"
	"log"
	"net/http"
	"os"
)

type application struct {
	errorLog *log.Logger
	infoLog  *log.Logger
}

func main() {
	addr := flag.String("addr", ":4000", "HTTP network addres")

	flag.Parse()

	infoLog := log.New(os.Stdout, "\u001b[32mINFO\u001b[0m\t", log.Ldate|log.Ltime)
	errLog := log.New(os.Stderr, "\u001b[31mERROR\u001b[0m\t", log.Ldate|log.Ltime|log.Lshortfile)

	app := &application{
		errorLog: errLog,
		infoLog:  infoLog,
	}

	srv := &http.Server{
		Addr:     *addr,
		ErrorLog: errLog,
		Handler:  app.routes(),
	}

	infoLog.Printf("Starting server on http://localhost%s", *addr)
	err := srv.ListenAndServe()
	errLog.Fatal(err)
}
