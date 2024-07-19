package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username string
	Email    string
}

type Handler struct {
	db *gorm.DB
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.RequestURI, time.Since(start))
	})
}

func prometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		timer := prometheus.NewTimer(prometheus.ObserverFunc(func(v float64) {
			requestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(v)
		}))

		next.ServeHTTP(w, r)
		timer.ObserveDuration()
	})
}

var (
	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "http_request_duration_seconds",
			Help: "Duration of HTTP requests.",
		},
		[]string{"method", "path"},
	)
)

func init() {
	prometheus.MustRegister(requestDuration)
}

func main() {
	conf := struct {
		Host     string
		Port     int
		User     string
		Password string
		DBName   string
	}{
		Host:     os.Getenv("DB_HOST"),
		Port:     5432,
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		DBName:   os.Getenv("DB_NAME"),
	}

	dsn := url.URL{
		User:     url.UserPassword(conf.User, conf.Password),
		Scheme:   "postgres",
		Host:     fmt.Sprintf("%s:%d", conf.Host, conf.Port),
		Path:     conf.DBName,
		RawQuery: (&url.Values{"sslmode": []string{"disable"}}).Encode(),
	}

	db, err := gorm.Open(postgres.Open(dsn.String()), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect to the database: ", err)
	}
	db.AutoMigrate(&User{})

	handler := Handler{db: db}

	router := http.NewServeMux()
	router.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "templates/index.html")
	})
	router.HandleFunc("GET /user", handler.GetUser)
	router.HandleFunc("POST /user", handler.CreateUser)
	router.HandleFunc("GET /users", handler.GetAllUsers)
	router.Handle("GET /metrics", promhttp.Handler())

	server := http.Server{
		Addr:    ":8080",
		Handler: prometheusMiddleware(loggingMiddleware(router)),
	}
	fmt.Println("Server is running on port 8080")
	server.ListenAndServe()
}

type GetUserResponse struct {
	User  User   `json:"user"`
	Error string `json:"error"`
}

func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")

	var user User
	h.db.Where("username = ?", username).First(&user)

	response := GetUserResponse{}
	if user.ID == 0 {
		response.Error = "User not found"
	} else {
		response.User = user
	}

	t, err := template.ParseFiles("templates/user.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = t.Execute(w, response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	// Get the username and email from the request body (form data)
	username := r.FormValue("username")
	email := r.FormValue("email")

	// Create a new user
	user := User{Username: username, Email: email}

	// Save the user in the database
	h.db.Create(&user)

	// Redirect to the user page
	http.Redirect(w, r, "/user?username="+username, http.StatusSeeOther)
}

type GetAllUsersResponse struct {
	Users      []User `json:"users"`
	Page       int    `json:"page"`
	TotalPages int    `json:"totalPages"`
	HasPrev    bool   `json:"hasPrev"`
	PrevPage   int    `json:"prevPage"`
	HasNext    bool   `json:"hasNext"`
	NextPage   int    `json:"nextPage"`
}

func (h *Handler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	// Default values for pagination
	const pageSize = 10
	var defaultPage = 1

	// Parse query parameters
	pageStr := r.URL.Query().Get("page")

	// Convert page to int
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = defaultPage
	}

	// Calculate offset
	offset := (page - 1) * pageSize

	// Count total number of users
	var totalUsers int64
	h.db.Model(&User{}).Count(&totalUsers)

	// Calculate total pages
	totalPages := int(totalUsers) / pageSize
	if int(totalUsers)%pageSize != 0 {
		totalPages++
	}

	var users []User
	// Apply pagination using Limit and Offset
	h.db.Limit(pageSize).Offset(offset).Find(&users)

	response := GetAllUsersResponse{
		Users:      users,
		Page:       page,
		TotalPages: totalPages,
		HasPrev:    page > 1,
		PrevPage:   page - 1,
		HasNext:    page < totalPages,
		NextPage:   page + 1,
	}

	t, err := template.ParseFiles("templates/users.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = t.Execute(w, response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
