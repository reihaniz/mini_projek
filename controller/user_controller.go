package controller

import (
	"database/sql"
	"mini_projek/model"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
)

type UserController struct {
	DB *sql.DB
}

var jwtSecret = []byte("reihan_iziz10")

func (uc *UserController) Register(c echo.Context) error {
	username := c.FormValue("username")
	password := c.FormValue("password")
	err := model.Register(uc.DB, username, password)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Registrasi Gagal"})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "User sukses terdaftar"})
}

func (uc *UserController) Login(c echo.Context) error {
	username := c.FormValue("username")
	password := c.FormValue("password")
	user, err := model.Login(uc.DB, username, password)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"message": "Username atau password salah"})
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(72 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal generate token"})
	}

	return c.JSON(http.StatusOK, map[string]string{"token": tokenString})
}

func (uc *UserController) AmbilSemuaUsers(c echo.Context) error {
	users, err := model.AmbilSemuaUsers(uc.DB)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Error fetching users"})
	}
	return c.JSON(http.StatusOK, users)
}

func (uc *UserController) AmbilUserByID(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid ID"})
	}

	user, err := model.AmbilUserByID(uc.DB, id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"message": "User not found"})
	}
	return c.JSON(http.StatusOK, user)
}
