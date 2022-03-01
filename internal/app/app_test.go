package app

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/ineverbee/adverts-go-api/internal/models"
	"github.com/stretchr/testify/require"
)

type MockServer struct{}

func (ms *MockServer) GetAdvert(id string, fields []string) (*models.Advert, error) {
	if _, err := strconv.Atoi(id); err != nil {
		return nil, err
	}
	ad := new(models.Advert)
	ad.Photos = []string{"", ""}
	return ad, nil
}

func (ms *MockServer) GetAllAdverts() ([]models.Advert, error) {
	return []models.Advert{{Photos: []string{"", ""}}, {Photos: []string{"", ""}}}, nil
}

func (ms *MockServer) CreateAdvert(*models.Advert) error {
	return nil
}

func TestHandlers(t *testing.T) {
	api = &MockServer{}
	router := mux.NewRouter()

	router.Handle("/adverts/{id}", loggingHandler(limit(errorHandler(GetAdvertHandler())))).Methods("GET")
	router.Handle("/adverts", loggingHandler(limit(errorHandler(GetAdvertsHandler())))).Methods("GET")
	router.Handle("/advert", loggingHandler(limit(errorHandler(CreateAdvertHandler())))).Methods("POST")

	tc := []struct {
		method, target string
		body           io.Reader
		code           int
	}{
		{"GET", "/", nil, http.StatusNotFound},
		{"GET", "/adverts/1", nil, http.StatusFound},
		{"GET", "/adverts/smth", nil, http.StatusInternalServerError},
		{"GET", "/adverts/1?fields=photos", nil, http.StatusFound},
		{"GET", "/adverts/1?fields=name", nil, http.StatusBadRequest},
		{"GET", "/adverts/1?fields=title,title", nil, http.StatusBadRequest},
		{"GET", "/adverts", nil, http.StatusFound},
		{"GET", "/adverts?sort=title", nil, http.StatusBadRequest},
		{"GET", "/adverts?sort=date&order=desc", nil, http.StatusFound},
		{"GET", "/adverts?order=desc", nil, http.StatusFound},
		{"GET", "/adverts?order=badinput", nil, http.StatusBadRequest},
		{"GET", "/adverts?sort=price&order=asc&page=100", nil, http.StatusFound},
		{"GET", "/adverts?sort=price&order=desc&page=NaN", nil, http.StatusBadRequest},
		{"POST", "/advert", strings.NewReader("{\"title\": \"IPhone\", \"ad_description\": \"red & white\", \"price\": 15, \"photos\": []}"), http.StatusCreated},
		{"POST", "/advert", strings.NewReader(fmt.Sprintf("{\"title\": \"%s\", \"ad_description\": \"red & white\", \"price\": 15, \"photos\": []}", strings.Repeat("a", 201))), http.StatusBadRequest},
		{"POST", "/advert", strings.NewReader(fmt.Sprintf("{\"title\": \"IPhone\", \"ad_description\": \"%s\", \"price\": 15, \"photos\": []}", strings.Repeat("a", 1001))), http.StatusBadRequest},
		{"POST", "/advert", strings.NewReader("{\"title\": \"IPhone\", \"ad_description\": \"red & white\", \"price\": 15, \"photos\": [\"\",\"\",\"\",\"\"]}"), http.StatusBadRequest},
		{"POST", "/advert", strings.NewReader("no]/:fie;OeFM"), http.StatusBadRequest},
	}
	for _, c := range tc {
		request(t, router, c.method, c.target, c.body, c.code)
	}

	req := httptest.NewRequest("GET", "/adverts/1", nil)
	rr := httptest.NewRecorder()
	for i := 0; i < 40; i++ {
		rr = httptest.NewRecorder()
		router.ServeHTTP(rr, req)
	}
	require.Equal(t, http.StatusTooManyRequests, rr.Code)
}

func request(t *testing.T, handler http.Handler, method, target string, body io.Reader, code int) {
	req := httptest.NewRequest(method, target, body)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	require.Equal(t, code, rr.Code)
}

func TestServer(t *testing.T) {
	timeout = 6 * time.Second
	require.Error(t, StartServer(), "there is no database to connect to")
}
