package app

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/ineverbee/adverts-go-api/internal/models"
	"golang.org/x/time/rate"
)

var fieldsMap = map[string]struct{}{
	"title":          {},
	"ad_description": {},
	"price":          {},
	"photos":         {},
}
var limiter = rate.NewLimiter(10, 30)

func limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !limiter.Allow() {
			http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Error represents a handler error. It provides methods for a HTTP status
// code and embeds the built-in error interface.
type Error interface {
	error
	Status() int
}

// StatusError represents an error with an associated HTTP status code.
type StatusError struct {
	Code int
	Err  error
}

// Allows StatusError to satisfy the error interface.
func (se StatusError) Error() string {
	return se.Err.Error()
}

// Returns our HTTP status code.
func (se StatusError) Status() int {
	return se.Code
}

func loggingHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Before executing the handler.
		start := time.Now()
		log.Printf("Strated %s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
		// After executing the handler.
		log.Printf("Completed %s in %v", r.URL.Path, time.Since(start))
	})
}

type errorHandler func(http.ResponseWriter, *http.Request) error

func (f errorHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := f(w, r)
	if err != nil {
		switch e := err.(type) {
		case Error:
			// We can retrieve the status here and write out a specific
			// HTTP status code.
			log.Printf("HTTP %d - %s", e.Status(), e)
			http.Error(w, e.Error(), e.Status())
		default:
			// Any error types we don't specifically look out for default
			// to serving a HTTP 500
			log.Printf("HTTP - %s", e)
			http.Error(w, http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError)
		}
	}
}

// GetAdvertHandler is used to get data inside the Adverts defined on our Advert catalog
func GetAdvertHandler() errorHandler {
	return func(rw http.ResponseWriter, r *http.Request) error {
		var fields []string
		onlyFirstPhoto := true
		params := mux.Vars(r)
		query := r.URL.Query()
		if fieldsString := query.Get("fields"); fieldsString != "" {
			for f := range fieldsMap {
				if strings.Count(fieldsString, f) > 1 {
					return &StatusError{
						http.StatusBadRequest,
						fmt.Errorf("error: wrong number of fields '%s'; available fields: title,ad_description,price,photos", fieldsString),
					}
				}
			}
			fields = strings.Split(fieldsString, ",")
			for _, s := range fields {
				if s == "photos" {
					onlyFirstPhoto = false
				}
				if _, exists := fieldsMap[s]; !exists {
					return &StatusError{
						http.StatusBadRequest,
						fmt.Errorf("error: there is no field like '%s'; available fields: title,ad_description,price,photos", s),
					}
				}
			}
		}
		ad, err := api.GetAdvert(params["id"], fields)
		if err != nil {
			return err
		}
		if onlyFirstPhoto && len(ad.Photos) > 1 {
			ad.Photos = ad.Photos[0:1]
		}
		data, _ := json.Marshal(ad)
		rw.Header().Add("content-type", "application/json")
		rw.WriteHeader(http.StatusFound)
		rw.Write(data)
		return nil
	}
}

// GetAdvertsHandler is used to get data inside the Adverts defined on our Advert catalog
func GetAdvertsHandler() errorHandler {
	return func(rw http.ResponseWriter, r *http.Request) error {
		ads, err := api.GetAllAdverts()
		if err != nil {
			return err
		}
		for i, ad := range ads {
			if len(ad.Photos) > 1 {
				ads[i].Photos = ads[i].Photos[0:1]
			}
		}
		query := r.URL.Query()
		order := query.Get("order")
		switch order {
		case "":
		case "asc":
		case "desc":
		default:
			return &StatusError{
				http.StatusBadRequest,
				fmt.Errorf("error: there is no order like %s; available orders: asc,desc", order),
			}
		}

		switch sorting := query.Get("sort"); {
		case sorting == "date" || sorting == "":
			if order == "desc" {
				sort.Slice(ads, func(i, j int) bool {
					return ads[i].ID > ads[j].ID
				})
			}
		case sorting == "price":
			if order == "desc" {
				sort.Slice(ads, func(i, j int) bool {
					return ads[i].Price > ads[j].Price
				})
			} else {
				sort.Slice(ads, func(i, j int) bool {
					return ads[i].Price < ads[j].Price
				})
			}
		default:
			return &StatusError{
				http.StatusBadRequest,
				fmt.Errorf("error: there is no sort like %s; available sorts: date,price", sorting),
			}
		}
		page := 0
		if val := query.Get("page"); val != "" {
			page, err = strconv.Atoi(val)
			if err != nil {
				return &StatusError{
					http.StatusBadRequest,
					fmt.Errorf("error: page should be integer, got '%v'", val),
				}
			}
		}

		start, end := Paginate(page, 10, len(ads))
		data, _ := json.Marshal(ads[start:end])
		rw.Header().Add("content-type", "application/json")
		rw.WriteHeader(http.StatusFound)
		rw.Write(data)
		return nil
	}
}

// CreateAdvertHandler is used to create a new Advert and add to our Advert store.
func CreateAdvertHandler() errorHandler {
	return func(rw http.ResponseWriter, r *http.Request) error {
		ad := new(models.Advert)
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(ad)
		if err != nil {
			return &StatusError{http.StatusBadRequest, err}
		}
		if len([]rune(ad.Ad_description)) > 1000 {
			return &StatusError{http.StatusBadRequest, fmt.Errorf("error: description shouldn't be more than 1000 characters")}
		}
		if len([]rune(ad.Title)) > 200 {
			return &StatusError{http.StatusBadRequest, fmt.Errorf("error: title shouldn't be more than 200 characters")}
		}
		if len(ad.Photos) > 3 {
			return &StatusError{http.StatusBadRequest, fmt.Errorf("error: shouldn't be more than 3 photo links")}
		}

		err = api.CreateAdvert(ad)
		if err != nil {
			return err
		}
		resp := map[string]interface{}{
			"message": "Success! Added new advert",
			"id":      ad.ID,
		}
		data, _ := json.Marshal(resp)
		rw.WriteHeader(http.StatusCreated)
		rw.Write(data)
		return nil
	}
}

func Paginate(pageNum int, pageSize int, sliceLength int) (int, int) {
	start := pageNum * pageSize

	if start > sliceLength {
		start = sliceLength
	}

	end := start + pageSize
	if end > sliceLength {
		end = sliceLength
	}

	return start, end
}
