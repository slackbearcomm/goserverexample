package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/jackc/pgx/v4/pgxpool"
)

var (
	DBClient *pgxpool.Pool // this is variable because we will assign new value once db is connected
)

const (
	// these are constants because the db credentials should not change
	DB_HOST     string = "localhost"
	DB_PORT     int    = 5432
	DB_USERNAME string = "root"
	DB_PASSWORD string = "secret"
	DB_NAME     string = "nextcrm"
)

func main() {
	// make db connection
	dbSource := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", DB_HOST, DB_PORT, DB_USERNAME, DB_PASSWORD, DB_NAME)
	dbClient, err := connectPostgres(dbSource)
	if err != nil {
		log.Fatal(err)
	}
	defer dbClient.Close()

	if dbClient == nil {
		log.Fatal("db connection failed")
	}

	// run api server
	runServer()
}

// connectPostgres makes db connection and return db client or error
func connectPostgres(source string) (*pgxpool.Pool, error) {
	var err error
	client, err := pgxpool.Connect(context.Background(), source)
	if err != nil {
		return nil, err
	}
	DBClient = client

	if err = client.Ping(context.Background()); err != nil {
		return nil, err
	}

	log.Println("Successfully connected with postgres db!")

	return client, nil
}

// runserver initates the go chi api server
func runServer() {
	serverAddress := ":8000"
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	routes(router)

	log.Printf("server running on port %s", serverAddress)
	http.ListenAndServe(serverAddress, router)
}

// routes host all the application routes
func routes(router *chi.Mux) {
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("welcome"))
	})
	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("check"))
	})

	// group other routes with /api
	router.Route("/api", func(r chi.Router) {
		pingRoutes(r)
		bookRoutes(r)
	})
}

func pingRoutes(r chi.Router) {
	// ping and king routes are basically the same
	// king route is divided into route and handler
	r.Route("/", func(r chi.Router) {
		r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("pong"))
		})
		r.Get("/king", kingHandler)
	})
}

func bookRoutes(r chi.Router) {
	r.Route("/books", func(r chi.Router) {
		r.Get("/", bookListHandler)
		r.Get("/{id}", bookHandler)
	})
}

func kingHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("kong"))
}

func bookListHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("books list"))
}

func bookHandler(w http.ResponseWriter, r *http.Request) {
	obj := bookService()
	render.JSON(w, r, obj)
}

type Book struct {
	ID         int64     `json:"id"`
	Code       string    `json:"code"`
	Name       string    `json:"name"`
	Auther     string    `json:"auther"`
	IsArchived bool      `json:"isArchived"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

func bookService() Book {
	return Book{
		ID:     1,
		Code:   "B0001",
		Name:   "Concept of Physics 1",
		Auther: "HC Verma",
	}
}
