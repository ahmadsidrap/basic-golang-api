package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

// Book struct
type Book struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Author string `json:"author"`
}

// User struct
type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// In-memory book storage
var books = map[string]Book{
	"1": {ID: "1", Title: "1984", Author: "George Orwell"},
	"2": {ID: "2", Title: "To Kill a Mockingbird", Author: "Harper Lee"},
}

// In-memory user storage
var users = map[string]User{
	"admin":    {ID: "1", Username: "admin", Password: "password123"},
	"user1":    {ID: "2", Username: "user1", Password: "securepass"},
	"john_doe": {ID: "3", Username: "john_doe", Password: "mypassword"},
}

// Secret key for signing JWTs (store this securely in env variables!)
var jwtSecret []byte

func init() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		fmt.Println("Error loading .env file")
		return
	}

	// Read the JWT_SECRET from environment variables
	jwtSecret = []byte(os.Getenv("JWT_SECRET"))
}

// Logging middleware
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s [%s]", r.RemoteAddr, r.Method, r.RequestURI, time.Since(start))
	})
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing token", http.StatusUnauthorized)
			return
		}

		// Extract token from "Bearer <token>"
		tokenString := authHeader[len("Bearer "):]

		// Parse and validate token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Generate JWT Token
func generateToken(w http.ResponseWriter, r *http.Request) {
	var creds struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Validate credentials by directly accessing the map
	user, exists := users[creds.Username]
	if !exists || user.Password != creds.Password {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": creds.Username,
		"exp":      time.Now().Add(time.Hour * 1).Unix(), // Token expires in 1 hour
	})

	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
}

// Get all books
func getBooks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(books)
}

// Get all books
func getBookDetail(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]

	// Check if book exists
	book, exists := books[id]
	if !exists {
		http.Error(w, "Book not found", http.StatusNotFound)
		return
	}

	// Decode request body into book
	json.NewDecoder(r.Body).Decode(&book)
	book.ID = id // Ensure ID remains unchanged

	// Update the book in the map
	books[id] = book

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(book)
}

// Create a new book
func createBook(w http.ResponseWriter, r *http.Request) {
	var newBook Book
	if err := json.NewDecoder(r.Body).Decode(&newBook); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Check if ID is provided
	if newBook.ID == "" {
		http.Error(w, "Book ID is required", http.StatusBadRequest)
		return
	}

	// Check if the book ID already exists
	if _, exists := books[newBook.ID]; exists {
		http.Error(w, "Book ID already exists", http.StatusConflict)
		return
	}

	// Store the new book in the map
	books[newBook.ID] = newBook

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newBook)
}

// Update a book
func updateBook(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]

	// Check if book exists
	book, exists := books[id]
	if !exists {
		http.Error(w, "Book not found", http.StatusNotFound)
		return
	}

	// Decode request body into book
	json.NewDecoder(r.Body).Decode(&book)
	book.ID = id // Ensure ID remains unchanged

	// Update the book in the map
	books[id] = book

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(book)
}

// Delete a book
func deleteBook(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]

	if _, exists := books[id]; !exists {
		http.Error(w, "Book not found", http.StatusNotFound)
		return
	}

	delete(books, id) // Delete book from the map
	w.WriteHeader(http.StatusNoContent)
}

func main() {
	r := mux.NewRouter()

	// Apply logging middleware to all routes
	r.Use(loggingMiddleware)

	// Public routes
	r.HandleFunc("/login", generateToken).Methods("POST")
	r.HandleFunc("/books", getBooks).Methods("GET")
	r.HandleFunc("/books/{id}", getBookDetail).Methods("GET")

	protected := r.PathPrefix("/").Subrouter()
	protected.Use(authMiddleware)
	protected.HandleFunc("/books", createBook).Methods("POST")
	protected.HandleFunc("/books/{id}", updateBook).Methods("PUT")
	protected.HandleFunc("/books/{id}", deleteBook).Methods("DELETE")

	fmt.Println("Server running on port 8080...")
	http.ListenAndServe(":8080", r)
}
