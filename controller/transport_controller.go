package controller

import (
	"database/sql"
	"fmt"
	"mini_projek/model"
	"net/http"
	"strconv"

	"context"

	"github.com/golang-jwt/jwt"
	"github.com/google/generative-ai-go/genai"
	"github.com/labstack/echo/v4"
	"google.golang.org/api/option"
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
func getEmissionRates(db *sql.DB) (map[string]float64, error) {
	query := "SELECT tipe_kendaraan, emission_rate FROM mode_transportasi"
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	emissionRates := make(map[string]float64)
	for rows.Next() {
		var transportType string
		var emissionRate float64
		if err := rows.Scan(&transportType, &emissionRate); err != nil {
			return nil, err
		}
		emissionRates[transportType] = emissionRate
	}
	return emissionRates, nil
}

func (tc *TransportController) CatatPerjalanan(c echo.Context) error {
	// Ambil user ID dari token
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"message": "Invalid token"})
	}

	transportType := c.FormValue("transport_type")
	distanceStr := c.FormValue("distance")
	distance, err := strconv.ParseFloat(distanceStr, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid distance value"})
	}

	emissionRates, err := getEmissionRates(tc.DB)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to retrieve emission rates"})
	}

	if _, exists := emissionRates[transportType]; !exists {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid transport type"})
	}

	// Simpan perjalanan
	err = model.SimpanPerjalanan(tc.DB, userID, transportType, distance)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to save journey"})
	}

	// Hitung emisi dari perjalanan ini
	emissions, err := model.HitungEmisi(tc.DB, transportType, distance)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to calculate emissions"})
	}

	// Temukan transportasi dengan emisi terendah yang bukan sepeda atau jalan kaki
	minEmissionTransport := ""
	minEmission := float64(1<<63 - 1) // Set to maximum possible float64 value

	for t, rate := range emissionRates {
		if rate > 0 && rate < minEmission {
			minEmission = rate
			minEmissionTransport = t
		}
	}

	// Hitung emisi untuk 1 km transportasi dengan emisi terendah
	lowestEmissions, err := model.HitungEmisi(tc.DB, minEmissionTransport, distance)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to calculate lowest emissions"})
	}

	// Buat prompt untuk API Gemini
	prompt := fmt.Sprintf("Bagaimana pendapat anda jika memakai transportasi %s dengan jarak %s km dengan emisi %.2f gram CO2? jika emisi tinggi sarankan dan jelaskan mengapa", transportType, distanceStr, emissions)

	// Panggil API Gemini
	hasilAI, err := AIPanggil(prompt)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to call Gemini API"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message":                   "Journey recorded successfully",
		"emissions":                 emissions,
		"lowest_emission_transport": minEmissionTransport,
		"lowest_emissions":          lowestEmissions,
		"hasil_ai":                  hasilAI,
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
	transportType := c.QueryParam("transport_type")
	distanceStr := c.QueryParam("distance")

	if distanceStr == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Distance parameter is required"})
	}

	distance, err := strconv.ParseFloat(distanceStr, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid distance value"})
	}

	emissions, err := model.HitungEmisi(tc.DB, transportType, distance)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to calculate emissions"})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"transport_type": transportType,
		"distance":       distance,
		"emissions":      emissions,
	})
}
func AIPanggil(prompt string) (string, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey("AIzaSyDfi6KqSa6wM9_wPBj4tR_kNkqPsut0-qg"))
	if err != nil {
		return "", fmt.Errorf("failed to create Gemini client: %v", err)
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-1.5-flash")
	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("failed to generate content from Gemini API: %v", err)
	}

	// Return the generated response
	return fmt.Sprintf("%s", resp.Candidates[0].Content.Parts[0]), nil
}
