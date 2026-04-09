package repository

import (
	"github.com/lightt77/veranda/internal/db"
)

// City represents a geographic location
type City struct {
	ID        int64   `json:"id"`
	Country   string  `json:"country"`
	City      string  `json:"city"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Timezone  string  `json:"timezone"`
}

// CitiesRepository handles city data access
type CitiesRepository struct {
	db *db.DB
}

// NewCitiesRepository creates a new cities repository
func NewCitiesRepository(database *db.DB) *CitiesRepository {
	return &CitiesRepository{db: database}
}

// GetAll returns all cities
func (r *CitiesRepository) GetAll() ([]City, error) {
	query := `SELECT id, country, city, latitude, longitude, timezone FROM cities ORDER BY country, city`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cities []City
	for rows.Next() {
		var c City
		if err := rows.Scan(&c.ID, &c.Country, &c.City, &c.Latitude, &c.Longitude, &c.Timezone); err != nil {
			return nil, err
		}
		cities = append(cities, c)
	}
	return cities, rows.Err()
}

// GetByCountry returns cities for a specific country
func (r *CitiesRepository) GetByCountry(country string) ([]City, error) {
	query := `SELECT id, country, city, latitude, longitude, timezone FROM cities WHERE country = ? ORDER BY city`
	rows, err := r.db.Query(query, country)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cities []City
	for rows.Next() {
		var c City
		if err := rows.Scan(&c.ID, &c.Country, &c.City, &c.Latitude, &c.Longitude, &c.Timezone); err != nil {
			return nil, err
		}
		cities = append(cities, c)
	}
	return cities, rows.Err()
}

// GetByID returns a city by ID
func (r *CitiesRepository) GetByID(id int64) (*City, error) {
	query := `SELECT id, country, city, latitude, longitude, timezone FROM cities WHERE id = ?`
	var c City
	err := r.db.QueryRow(query, id).Scan(&c.ID, &c.Country, &c.City, &c.Latitude, &c.Longitude, &c.Timezone)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// Search searches for cities by name (fuzzy match)
func (r *CitiesRepository) Search(query string) ([]City, error) {
	searchQuery := `%` + query + `%`
	sql := `SELECT id, country, city, latitude, longitude, timezone FROM cities 
		WHERE city LIKE ? OR country LIKE ? ORDER BY country, city`
	rows, err := r.db.Query(sql, searchQuery, searchQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cities []City
	for rows.Next() {
		var c City
		if err := rows.Scan(&c.ID, &c.Country, &c.City, &c.Latitude, &c.Longitude, &c.Timezone); err != nil {
			return nil, err
		}
		cities = append(cities, c)
	}
	return cities, rows.Err()
}

// GetCountries returns all unique countries
func (r *CitiesRepository) GetCountries() ([]string, error) {
	query := `SELECT DISTINCT country FROM cities ORDER BY country`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var countries []string
	for rows.Next() {
		var country string
		if err := rows.Scan(&country); err != nil {
			return nil, err
		}
		countries = append(countries, country)
	}
	return countries, rows.Err()
}

// DefaultCity returns Mumbai as the default city
func (r *CitiesRepository) DefaultCity() (*City, error) {
	query := `SELECT id, country, city, latitude, longitude, timezone FROM cities WHERE city = 'Mumbai' AND country = 'India'`
	var c City
	err := r.db.QueryRow(query).Scan(&c.ID, &c.Country, &c.City, &c.Latitude, &c.Longitude, &c.Timezone)
	if err != nil {
		// If Mumbai not found, return first city
		return r.GetByID(1)
	}
	return &c, nil
}
