package main

import (
	"mini_projek/config"
	"mini_projek/controller"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	e := echo.New()
	db := config.InitDB()
	defer db.Close()

	userController := &controller.UserController{DB: db}

	e.POST("/register", userController.Register)
	e.POST("/login", userController.Login)

	r := e.Group("/users")
	r.Use(middleware.JWTWithConfig(middleware.JWTConfig{
		SigningKey: []byte("reihan_iziz10"),
	}))
	r.GET("", userController.AmbilSemuaUsers)
	r.GET("/:id", userController.AmbilUserByID)

	e.Start(":8080")
}
