package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/alexedwards/scs/pgxstore"

	"github.com/alexedwards/scs/v2" // New import
	//"github.com/alexedwards/scs/pgxstore" 	// New import

	"github.com/go-playground/form/v4"

	"github.com/jackc/pgx/v4/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"

	"html/template"
	"log"
	"net/http"
	"os"
	"snippbox/internal/models"
	"time" // New import
)

// Add a new sessionManager field to the application struct.
type application struct {
	errorLog       *log.Logger
	infoLog        *log.Logger
	snippets       *models.SnippetModel
	templateCache  map[string]*template.Template
	formDecoder    *form.Decoder
	sessionManager *scs.SessionManager
}

var sessionManager *scs.SessionManager

func main() {
	addr := flag.String("addr", ":4000", "HTTP network address")
	dbUrl := "postgresql://postgres:1234@localhost:5432/pgx_snippetbox"
	//dsn := flag.String("dsn", "root:1234@/snippetbox?parseTime=true", "MySQL data source name")
	flag.Parse()
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	/*db, err := openDB(*dsn)
	if err != nil {
		errorLog.Fatal(err)
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {

		}
	}(db)

	*/
	dbPool, err := pgxpool.Connect(context.Background(), dbUrl)
	if err != nil {
		fmt.Println("failed to connect to postgresql", err)
		return
	}
	// to close DB pool
	defer dbPool.Close()

	templateCache, err := newTemplateCache()
	if err != nil {
		errorLog.Fatal(err)
	}
	formDecoder := form.NewDecoder()
	// Use the scs.New() function to initialize a new session manager. Then we
	// configure it to use our MySQL database as the session store, and set a
	// lifetime of 12 hours (so that sessions automatically expire 12 hours
	// after first being created).
	// Initialize a new session manager and configure it to use pgxstore as the session store.
	sessionManager = scs.New()
	sessionManager.Store = pgxstore.New(dbPool)
	sessionManager.Lifetime = 12 * time.Hour
	// And add the session manager to our application dependencies.
	app := &application{
		errorLog:       errorLog,
		infoLog:        infoLog,
		snippets:       &models.SnippetModel{DB: dbPool},
		templateCache:  templateCache,
		formDecoder:    formDecoder,
		sessionManager: sessionManager,
	}
	srv := &http.Server{
		Addr:     *addr,
		ErrorLog: errorLog,
		Handler:  app.routes(),
	}
	infoLog.Printf("Starting server on %s", *addr)
	err = srv.ListenAndServe()
	errorLog.Fatal(err)
}

// The openDB() function wraps sql.Open() and returns a sql.DB connection pool
// for a given DSN.
/*
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

*/
