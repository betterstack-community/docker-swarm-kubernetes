package main

import (
	"log"
	"net/http"

	"donaldle.com/m/handler"
	"donaldle.com/m/helper"
	"github.com/julienschmidt/httprouter"
	_ "github.com/lib/pq"
)

func main() {
	router := httprouter.New()
	router.GET("/", handler.AllBlogs)
	router.POST("/signup", handler.Signup)
	router.POST("/login", handler.Login)
	router.GET("/logout", handler.Logout)
	router.POST("/blog", helper.ValidateJWTTokenMiddleware(handler.CreateBlog)) // Needs Authorization header 'Bearer [token]'
	router.GET("/blog/:id", helper.ValidateJWTTokenMiddleware(handler.OneBlog))
	router.PUT("/blog/:id", helper.ValidateJWTTokenMiddleware(handler.UpdateBlog))
	router.DELETE("/blog/:id", helper.ValidateJWTTokenMiddleware(handler.DeleteBlog))

	log.Fatal(http.ListenAndServe(":8080", router))
}
