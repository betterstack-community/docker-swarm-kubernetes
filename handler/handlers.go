package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"net/http"
	"time"

	"donaldle.com/m/config"
	"donaldle.com/m/helper"
	"github.com/julienschmidt/httprouter"
	_ "github.com/lib/pq"
)

type Blog struct {
	ID        int       `json:"id"`
	UserId    int       `json:"user_id"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"created_at"`
}

type SignUpUser struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserCredentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Token struct {
	Token string `json:"token"`
}

type CreatingBlog struct {
	Body string `json:"body"`
}

func AllBlogs(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// We only accept 'GET' method here
	if r.Method != "GET" {
		http.Error(w, http.StatusText(405), http.StatusMethodNotAllowed)
		return
	}

	// Get all blogs from DB
	rows, err := config.DB.Query("SELECT * FROM blog")
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}
	// Close the db connection at the end
	defer rows.Close()

	// Create blog object list
	blogs := make([]Blog, 0)
	for rows.Next() {
		blog := Blog{}
		err := rows.Scan(&blog.ID, &blog.UserId, &blog.Body, &blog.CreatedAt, &blog.UpdatedAt) // order matters
		if err != nil {
			http.Error(w, http.StatusText(500), 500)
			return
		}
		blogs = append(blogs, blog)
	}
	if err = rows.Err(); err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}

	// Returns as JSON (List of Blog objects)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(blogs); err != nil {
		panic(err)
	}
}

func Signup(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	var user SignUpUser

	err := json.NewDecoder(r.Body).Decode(&user)

	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, "Error in request")
		return
	}

	hash, _ := helper.HashPassword(user.Password)

	_, err = config.DB.Exec("INSERT INTO users (USERNAME, EMAIL, PASSWORD) VALUES ($1, $2, $3)", user.Username, user.Email, hash)
	if err != nil {
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}
}

func Login(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	var user UserCredentials

	err := json.NewDecoder(r.Body).Decode(&user)

	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, "Error in request")
		return
	}

	row := config.DB.QueryRow("SELECT * FROM users WHERE email = $1 LIMIT 1", user.Email)

	userDict := User{}
	er := row.Scan(&userDict.ID, &userDict.Username, &userDict.Email, &userDict.Password, &userDict.CreatedAt)
	switch {
	case err == sql.ErrNoRows:
		http.NotFound(w, r)
		return
	case er != nil:
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}

	if !helper.CheckPasswordHash(user.Password, userDict.Password) {
		w.WriteHeader(http.StatusForbidden)
		http.Error(w, "Email and/or password do not match", http.StatusForbidden)
		return
	}

	token := jwt.New(jwt.SigningMethodRS256)
	claims := make(jwt.MapClaims)
	claims["exp"] = time.Now().Add(time.Hour * time.Duration(1)).Unix()
	claims["iat"] = time.Now().Unix()
	claims["user_id"] = userDict.ID
	token.Claims = claims

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "Error extracting the key")
		panic(err)
	}

	tokenString, err := token.SignedString("secret")

	// Set Cookie (Maybe not needed?)
	expireCookie := time.Now().Add(time.Hour * 1)
	cookie := http.Cookie{Name: "Auth", Value: tokenString, Expires: expireCookie, HttpOnly: true}
	http.SetCookie(w, &cookie)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "Error while signing the token")
		panic(err)
	}

	response := Token{tokenString}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		panic(err)
	}
}

func Logout(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// We don't need logout since it just needs to delete token from client (browser)
	// But we can delete cookie if we use it
	deleteCookie := http.Cookie{Name: "Auth", Value: "none", Expires: time.Now()}
	http.SetCookie(w, &deleteCookie)
	return
}

func CreateBlog(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Get user_id from decoded JWT token
	userId := r.Context().Value("user_id")

	var blog CreatingBlog

	err := json.NewDecoder(r.Body).Decode(&blog)

	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, "Error in request")
		return
	}

	_, err = config.DB.Exec("INSERT INTO blog (USER_ID, BODY) VALUES ($1, $2)", userId, blog.Body)
	if err != nil {
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}
}

func OneBlog(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// We only accept 'GET' method here
	if r.Method != "GET" {
		http.Error(w, http.StatusText(405), http.StatusMethodNotAllowed)
		return
	}

	blogID := ps.ByName("id")

	// Get the specific blog from DB
	row := config.DB.QueryRow("SELECT * FROM blog WHERE id = $1", blogID)

	// Create blog object
	blog := Blog{}
	err := row.Scan(&blog.ID, &blog.UserId, &blog.Body, &blog.CreatedAt, &blog.UpdatedAt)
	switch {
	case err == sql.ErrNoRows:
		http.NotFound(w, r)
		return
	case err != nil:
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}

	// Returns as JSON (single Blog object)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(blog); err != nil {
		panic(err)
	}
}

func UpdateBlog(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Get user_id from decoded JWT token
	// Needs to convert float64 to int for the value from context
	rawUserId := r.Context().Value("user_id").(float64)
	userId := int(rawUserId)

	// Check if the user is the author of the blog
	blogID := ps.ByName("id")
	row := config.DB.QueryRow("SELECT * FROM blog WHERE id = $1", blogID)
	// Create blog object
	updatingBlog := Blog{}
	er := row.Scan(&updatingBlog.ID, &updatingBlog.UserId, &updatingBlog.Body, &updatingBlog.CreatedAt, &updatingBlog.UpdatedAt)
	switch {
	case er == sql.ErrNoRows:
		http.NotFound(w, r)
		return
	case er != nil:
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}

	if updatingBlog.UserId != userId {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, "Unauthorized access to this resource")
		return
	}

	var blog CreatingBlog

	err := json.NewDecoder(r.Body).Decode(&blog)

	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, "Error in request")
		return
	}

	_, err = config.DB.Exec("UPDATE blog SET body = $1 WHERE id = $2", blog.Body, updatingBlog.ID)
	if err != nil {
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}
}

func DeleteBlog(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Get user_id from decoded JWT token
	// Needs to convert float64 to int for the value from context
	rawUserId := r.Context().Value("user_id").(float64)
	userId := int(rawUserId)

	// Check if the user is the author of the blog
	blogID := ps.ByName("id")
	row := config.DB.QueryRow("SELECT * FROM blog WHERE id = $1", blogID)
	// Create blog object
	deletingBlog := Blog{}
	er := row.Scan(&deletingBlog.ID, &deletingBlog.UserId, &deletingBlog.Body, &deletingBlog.CreatedAt, &deletingBlog.UpdatedAt)
	switch {
	case er == sql.ErrNoRows:
		http.NotFound(w, r)
		return
	case er != nil:
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}

	if deletingBlog.UserId != userId {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, "Unauthorized access to this resource")
		return
	}

	_, err := config.DB.Exec("DELETE FROM blog WHERE id = $1", deletingBlog.ID)
	if err != nil {
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}
}
