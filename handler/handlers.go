package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"donaldle.com/m/config"
	"github.com/julienschmidt/httprouter"
	_ "github.com/lib/pq"
)

type Posts struct {
	ID        int       `json:"id"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreatingPosts struct {
	Body string `json:"body"`
}

func AllBlogs(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// We only accept 'GET' method here
	if r.Method != "GET" {
		http.Error(w, http.StatusText(405), http.StatusMethodNotAllowed)
		return
	}

	// Get all blogs from DB
	rows, err := config.DB.Query("SELECT * FROM posts")
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}
	// Close the db connection at the end
	defer rows.Close()

	// Create blog object list
	blogs := make([]Posts, 0)
	for rows.Next() {
		blog := Posts{}
		err := rows.Scan(&blog.ID, &blog.Body, &blog.CreatedAt, &blog.UpdatedAt) // order matters
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

func CreateBlog(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	var blog CreatingPosts

	err := json.NewDecoder(r.Body).Decode(&blog)

	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, "Error in request")
		return
	}

	_, err = config.DB.Exec("INSERT INTO posts (BODY) VALUES ($1)", blog.Body)
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
	row := config.DB.QueryRow("SELECT * FROM posts WHERE id = $1", blogID)

	// Create blog object
	blog := Posts{}
	err := row.Scan(&blog.ID, &blog.Body, &blog.CreatedAt, &blog.UpdatedAt)
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
	// Needs to convert float64 to int for the value from context

	blogID := ps.ByName("id")
	row := config.DB.QueryRow("SELECT * FROM posts WHERE id = $1", blogID)
	// Create blog object
	updatingBlog := Posts{}
	er := row.Scan(&updatingBlog.ID,
		&updatingBlog.Body, &updatingBlog.CreatedAt, &updatingBlog.UpdatedAt)
	switch {
	case er == sql.ErrNoRows:
		http.NotFound(w, r)
		return
	case er != nil:
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}

	var blog CreatingPosts

	err := json.NewDecoder(r.Body).Decode(&blog)

	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, "Error in request")
		return
	}

	_, err = config.DB.Exec("UPDATE posts SET body = $1 WHERE id = $2", blog.Body, updatingBlog.ID)
	if err != nil {
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}
}

func DeleteBlog(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	blogID := ps.ByName("id")
	row := config.DB.QueryRow("SELECT * FROM posts WHERE id = $1", blogID)
	// Create blog object
	deletingBlog := Posts{}
	er := row.Scan(&deletingBlog.ID,
		&deletingBlog.Body, &deletingBlog.CreatedAt, &deletingBlog.UpdatedAt)
	switch {
	case er == sql.ErrNoRows:
		http.NotFound(w, r)
		return
	case er != nil:
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}

	_, err := config.DB.Exec("DELETE FROM posts WHERE id = $1", deletingBlog.ID)
	if err != nil {
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}
}
