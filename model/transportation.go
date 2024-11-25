package model

import (
	"database/sql"
	"time"
)

type Transportation struct {
	ID           int     `json:"id"`
	Type         string  `json:"type"`
	EmissionRate float64 `json:"emission_rate"` // gram CO2 per kilometer
}

type Journey struct {
	ID             int       `json:"id"`
	UserID         int       `json:"user_id"`
	TransportType  string    `json:"transport_type"`
	Distance       float64   `json:"distance"` // dalam kilometer
	EmissionAmount float64   `json:"emission_amount"`
	Date           time.Time `json:"date"`
}

// Konstanta emisi karbon (gram CO2 per kilometer)

func HitungEmisi(db *sql.DB, transportType string, distance float64) (float64, error) {
	var emissionRate float64
	query := `SELECT emission_rate FROM mode_transportasi WHERE tipe_kendaraan = ?`
	err := db.QueryRow(query, transportType).Scan(&emissionRate)
	if err != nil {
		return 0, err
	}
	return emissionRate * distance, nil
}

func SimpanPerjalanan(db *sql.DB, userID int, transportType string, distance float64) error {
	emissionAmount, err := HitungEmisi(db, transportType, distance)
	if err != nil {
		return err
	}

	query := `INSERT INTO journeys (user_id, transport_type, distance, emission_amount, date) 
              VALUES (?, ?, ?, ?, NOW())`

	_, err = db.Exec(query, userID, transportType, distance, emissionAmount)
	return err
}

func AmbilRiwayatPerjalanan(db *sql.DB, userID int) ([]Journey, error) {
	query := `SELECT id, transport_type, distance, emission_amount, date 
              FROM journeys 
              WHERE user_id = ? 
              ORDER BY date DESC`

	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var journeys []Journey
	for rows.Next() {
		var j Journey
		var dateBytes []byte
		err := rows.Scan(&j.ID, &j.TransportType, &j.Distance, &j.EmissionAmount, &dateBytes)
		if err != nil {
			return nil, err
		}

		date, err := time.Parse("2006-01-02 15:04:05", string(dateBytes))
		if err != nil {
			return nil, err
		}
		j.Date = date

		journeys = append(journeys, j)
	}
	return journeys, nil
}

func TotalEmisi(db *sql.DB, userID int) (map[string]float64, error) {
	query := `SELECT transport_type, SUM(emission_amount) as total_emission 
              FROM journeys 
              WHERE user_id = ? 
              GROUP BY transport_type`

	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]float64)
	for rows.Next() {
		var transportType string
		var totalEmission float64
		err := rows.Scan(&transportType, &totalEmission)
		if err != nil {
			return nil, err
		}
		result[transportType] = totalEmission
	}
	return result, nil
}
