package controller

import (
	"database/sql"
	"encoding/json"
	"mini_projek/model"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func setupMockDB() (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		panic(err)
	}
	return db, mock
}

func TestAmbilSemuaUsers(t *testing.T) {
	e := echo.New()
	db, mock := setupMockDB()
	defer db.Close()

	userController := &UserController{DB: db}

	// Mock query
	mock.ExpectQuery("SELECT id, username FROM users").
		WillReturnRows(sqlmock.NewRows([]string{"id", "username"}).
			AddRow(1, "testuser1").
			AddRow(2, "testuser2"))

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := userController.AmbilSemuaUsers(c)
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)

		// Check response body
		var users []model.User
		err := json.Unmarshal(rec.Body.Bytes(), &users)
		assert.NoError(t, err)
		assert.Len(t, users, 2)
		assert.Equal(t, "testuser1", users[0].Username)
		assert.Equal(t, "testuser2", users[1].Username)
	}
}

func TestAmbilUserByID(t *testing.T) {
	e := echo.New()
	db, mock := setupMockDB()
	defer db.Close()

	userController := &UserController{DB: db}

	// Mock query
	mock.ExpectQuery("SELECT id, username FROM users WHERE id = ?").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username"}).
			AddRow(1, "testuser"))

	req := httptest.NewRequest(http.MethodGet, "/users/1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("1")

	err := userController.AmbilUserByID(c)
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)

		var user model.User
		err := json.Unmarshal(rec.Body.Bytes(), &user)
		assert.NoError(t, err)
		assert.Equal(t, 1, user.ID)
		assert.Equal(t, "testuser", user.Username)
	}
}

func TestAmbilUserByID_NotFound(t *testing.T) {
	e := echo.New()
	db, mock := setupMockDB()
	defer db.Close()

	userController := &UserController{DB: db}

	// Mock query
	mock.ExpectQuery("SELECT id, username FROM users WHERE id = ?").
		WithArgs(999).
		WillReturnError(sql.ErrNoRows)

	req := httptest.NewRequest(http.MethodGet, "/users/999", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("999")

	err := userController.AmbilUserByID(c)
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusNotFound, rec.Code)
		assert.JSONEq(t, `{"message":"User tidak ditemukan~~"}`, rec.Body.String())
	}
}
