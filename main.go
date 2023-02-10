package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

var (
	DBClient *pgxpool.Pool // this is variable because we will assign new value once db is connected
)

const (
	// these are constants because the db credentials should not change
	serverAddress string = ":8080"
	DB_HOST       string = "localhost"
	DB_PORT       int    = 5432
	DB_USERNAME   string = "root"
	DB_PASSWORD   string = "secret"
	DB_NAME       string = "nextcrm"
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
		r.Post("/", bookCreateHandler)
		r.Put("/{id}", bookUpdateHandler)
		r.Delete("/{id}", bookDeleteHandler)
	})
}

//////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////// handlers /////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////

func kingHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("kong"))
}

func bookListHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	result, err := bookListService(ctx)
	if err != nil {
		render.JSON(w, r, err.Error())
	}

	// This will also work
	// result, err := bookListStore(ctx)
	// if err != nil {
	// 	render.JSON(w, r, err.Error())
	// }

	render.JSON(w, r, result)
}

func bookHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")
	id, err := StringToInt64(idStr)
	if err != nil {
		render.JSON(w, r, err.Error())
	}
	obj, err := bookService(ctx, id)
	if err != nil {
		render.JSON(w, r, err.Error())
	}
	render.JSON(w, r, obj)
}

// book create handler
func bookCreateHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	req := Book{}

	if decodeErr := json.NewDecoder(r.Body).Decode(&req); decodeErr != nil {
		err := fmt.Errorf("invalid json request")
		render.JSON(w, r, err.Error())
		return
	}
	defer r.Body.Close()

	obj, err := bookCreateService(ctx, req)
	if err != nil {
		render.JSON(w, r, err.Error())
	}
	render.JSON(w, r, obj)
}

func bookUpdateHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("book update"))
}

func bookDeleteHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("book delete"))
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

//////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////// Services /////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////

func bookListService(ctx context.Context) ([]Book, error) {
	result, err := bookListStore(ctx)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func bookService(ctx context.Context, id int64) (*Book, error) {
	obj, err := bookGetByIDStore(ctx, id)
	if err != nil {
		return nil, err
	}

	return obj, nil
}

func bookCreateService(ctx context.Context, req Book) (*Book, error) {
	// get all books list
	books, err := bookListStore(ctx)
	if err != nil {
		return nil, err
	}

	// logic
	code := fmt.Sprintf("B%05d", len(books)+1)
	req.Code = code

	// connect to dbstore
	tx, err := BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer RollbackTx(ctx, tx)

	obj, err := bookInsertStore(ctx, tx, req)
	if err != nil {
		return nil, err
	}

	if err := CommitTx(ctx, tx); err != nil {
		return nil, err
	}

	return obj, nil
}

//////////////////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////// Store /////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////

func bookListStore(ctx context.Context) ([]Book, error) {
	result := []Book{}
	obj := Book{}

	queryStmt := `SELECT * FROM books`
	rows, err := DBClient.Query(ctx, queryStmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		if err := rows.Scan(
			&obj.ID,
			&obj.Code,
			&obj.Name,
			&obj.Auther,
			&obj.IsArchived,
			&obj.CreatedAt,
			&obj.UpdatedAt,
		); err != nil {
			return nil, err
		}
		result = append(result, obj)
	}

	return result, nil
}

func bookGetByIDStore(ctx context.Context, id int64) (*Book, error) {
	obj := Book{}

	queryStmt := `
		SELECT * FROM books
		WHERE books.id = $1
	`
	row := DBClient.QueryRow(ctx, queryStmt, id)
	if err := row.Scan(
		&obj.ID,
		&obj.Code,
		&obj.Name,
		&obj.Auther,
		&obj.IsArchived,
		&obj.CreatedAt,
		&obj.UpdatedAt,
	); err != nil {
		return nil, err
	}

	return &obj, nil
}

func bookInsertStore(ctx context.Context, tx pgx.Tx, arg Book) (*Book, error) {
	obj := &Book{}

	queryStmt := `
	INSERT INTO
	books(
		code,
		name,
		auther,
		is_archived
	)
	VALUES ($1, $2, $3, $4)
	RETURNING *
	`

	row := tx.QueryRow(ctx, queryStmt,
		&arg.Code,
		&arg.Name,
		&arg.Auther,
		&arg.IsArchived,
	)

	if err := row.Scan(
		&obj.ID,
		&obj.Code,
		&obj.Name,
		&obj.Auther,
		&obj.IsArchived,
		&obj.CreatedAt,
		&obj.UpdatedAt,
	); err != nil {
		return nil, err
	}

	return obj, nil
}

//////////////////////////////////////////////////////////////////////////////////////////
/////////////////////////////////////// Helpers /////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////

func StringToInt64(str string) (int64, error) {
	i, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return 0, err
	}

	return i, nil
}

//////////////////////////////////////////////////////////////////////////////////////////
/////////////////////////////// Database Transactions ///////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////

func BeginTx(ctx context.Context) (pgx.Tx, error) {
	tx, err := DBClient.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	log.Println("Beign Transaction")

	return tx, nil

}

func CommitTx(ctx context.Context, tx pgx.Tx) error {
	err := tx.Commit(ctx)
	if err != nil {
		return err
	}
	log.Println("Commit Transaction")

	return nil
}

func RollbackTx(ctx context.Context, tx pgx.Tx) error {
	err := tx.Rollback(ctx)
	if err != nil {
		log.Println(err.Error())
		return nil
	}
	log.Println("Rollback Transaction")

	return nil
}
