package controller

import (
	"database/sql"
	"mini_projek/model"
	"net/http"
	"strconv"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
)

type TransportController struct {
	DB *sql.DB
}

// Helper function untuk mengambil user ID dari token JWT
func getUserIDFromToken(c echo.Context) (int, error) {
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userID := int(claims["user_id"].(float64))
	return userID, nil
}

func (tc *TransportController) CatatPerjalanan(c echo.Context) error {
	// Mengambil user ID dari token
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"message": "Invalid token"})
	}

	transportType := c.FormValue("transport_type")
	distance, err := strconv.ParseFloat(c.FormValue("distance"), 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid distance value"})
	}

	if _, exists := model.EmissionRates[transportType]; !exists {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid transport type"})
	}

	err = model.SimpanPerjalanan(tc.DB, userID, transportType, distance)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to save journey"})
	}

	emissions := model.HitungEmisi(transportType, distance)
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message":   "Journey recorded successfully",
		"emissions": emissions,
	})
}

func (tc *TransportController) AmbilRiwayat(c echo.Context) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"message": "Invalid token"})
	}

	journeys, err := model.AmbilRiwayatPerjalanan(tc.DB, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": "Failed to fetch journey history",
			"error":   err.Error(),
		})
	}
	return c.JSON(http.StatusOK, journeys)
}

func (tc *TransportController) BandingkanEmisi(c echo.Context) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"message": "Invalid token"})
	}

	comparison, err := model.BandingkanEmisi(tc.DB, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to compare emissions"})
	}
	return c.JSON(http.StatusOK, comparison)
}

func (tc *TransportController) HitungEmisiPerjalanan(c echo.Context) error {
	// Tidak perlu user ID untuk perhitungan sederhana
	transportType := c.QueryParam("transport_type")
	distance, err := strconv.ParseFloat(c.QueryParam("distance"), 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid distance value"})
	}

	if _, exists := model.EmissionRates[transportType]; !exists {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid transport type"})
	}

	emissions := model.HitungEmisi(transportType, distance)
	return c.JSON(http.StatusOK, map[string]interface{}{
		"transport_type": transportType,
		"distance":       distance,
		"emissions":      emissions,
	})
}
