package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

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

	server := http.Server{
		Addr:    ":8080",
		Handler: loggingMiddleware(router),
	}
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

func (h *Handler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	var users []User
	h.db.Find(&users)

	t, err := template.ParseFiles("templates/users.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = t.Execute(w, users)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
