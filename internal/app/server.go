package app

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/gorilla/mux"
	"github.com/ineverbee/adverts-go-api/internal/models"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type Server interface {
	GetAdvert(string, []string) (*models.Advert, error)
	GetAllAdverts() ([]models.Advert, error)
	CreateAdvert(*models.Advert) error
}

type ApiServer struct {
	server   http.Server
	database *pgxpool.Pool
}

var api Server
var timeout = 30 * time.Second

func StartServer() error {
	var err error
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	router := mux.NewRouter()

	router.Handle("/adverts/{id}", loggingHandler(limit(errorHandler(GetAdvertHandler())))).Methods("GET")
	router.Handle("/adverts", loggingHandler(limit(errorHandler(GetAdvertsHandler())))).Methods("GET")
	router.Handle("/advert", loggingHandler(limit(errorHandler(CreateAdvertHandler())))).Methods("POST")

	connStr := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?connect_timeout=5",
		os.Getenv("DB_USERNAME"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)

	dbConn, err := NewDB(ctx, connStr)
	if err != nil {
		return err
	}

	api = &ApiServer{http.Server{Addr: ":8080", Handler: router}, dbConn}

	log.Println("Staring server on Port 8080")
	err = http.ListenAndServe(":8080", router)
	return err
}

func NewDB(ctx context.Context, connStr string) (*pgxpool.Pool, error) {
	log.Printf("Trying to connect to %s\n", connStr)
	var (
		conn *pgxpool.Pool
		err  error
	)

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	timeoutExceeded := time.After(timeout)
LOOP:
	for {
		select {
		case <-timeoutExceeded:
			return nil, fmt.Errorf("db connection failed after %s timeout", timeout)

		case <-ticker.C:
			conn, err = pgxpool.Connect(ctx, connStr)
			if err == nil {
				break LOOP
			}
			log.Println("Failed! Trying to reconnect..")
		}
	}

	err = conn.Ping(ctx)
	if err != nil {
		return nil, err
	}

	log.Println("Connect success!")

	return conn, nil
}

func (s *ApiServer) GetAdvert(id string, fields []string) (*models.Advert, error) {
	ad := new(models.Advert)
	q := "SELECT "
	if l := len(fields); l != 0 {
		for i, field := range fields {
			q += field
			if i+1 != l {
				q += ", "
			}
		}
		q += " FROM adverts WHERE id=" + id
	} else {
		q += "title, price, photos FROM adverts WHERE id=" + id
	}
	err := pgxscan.Get(context.TODO(), s.database, ad, q)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &StatusError{http.StatusBadRequest, err}
		}
		return nil, err
	}
	return ad, nil
}

func (s *ApiServer) GetAllAdverts() ([]models.Advert, error) {
	var l int
	err := s.database.QueryRow(context.TODO(), "select count(*) from adverts").Scan(&l)
	if err != nil {
		return nil, err
	}

	ads := make([]models.Advert, l)
	rows, err := s.database.Query(context.TODO(), "select id, title, price, photos from adverts")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for i := 0; rows.Next(); i++ {
		rows.Scan(&ads[i].ID, &ads[i].Title, &ads[i].Price, &ads[i].Photos)
	}
	return ads, nil
}

func (s *ApiServer) CreateAdvert(ad *models.Advert) error {
	err := s.database.QueryRow(context.TODO(), "insert into adverts (title, ad_description, price, photos) values ($1,$2,$3,$4) returning id",
		ad.Title,
		ad.Ad_description,
		ad.Price,
		ad.Photos,
	).Scan(&ad.ID)
	if err != nil {
		return err
	}
	return nil
}
