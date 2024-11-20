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
	transportController := &controller.TransportController{DB: db}

	e.POST("/register", userController.Register)
	e.POST("/login", userController.Login)

	// User routes
	r := e.Group("/users")
	r.Use(middleware.JWTWithConfig(middleware.JWTConfig{
		SigningKey: []byte("reihan_iziz10"),
	}))
	r.GET("", userController.AmbilSemuaUsers)
	r.GET("/:id", userController.AmbilUserByID)

	// Transport routes
	t := e.Group("/transport")
	t.Use(middleware.JWTWithConfig(middleware.JWTConfig{
		SigningKey: []byte("reihan_iziz10"),
	}))
	t.POST("/journey", transportController.CatatPerjalanan)
	t.GET("/history", transportController.AmbilRiwayat)
	t.GET("/total", transportController.TotalEmisi)
	t.GET("/calculate", transportController.HitungEmisiPerjalanan)

	e.Start(":8080")
}
