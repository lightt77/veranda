package db

import "database/sql"

// timersTable stores timer data with millisecond precision
const timersTable = `
CREATE TABLE IF NOT EXISTS timers (
	id INTEGER PRIMARY KEY,                    -- Unix epoch timestamp in seconds (natural sortable ID)
	label TEXT NOT NULL,                       -- Timer name/description
	duration_ms INTEGER NOT NULL,              -- Total duration in milliseconds
	remaining_ms INTEGER NOT NULL,             -- Remaining time in milliseconds
	status TEXT NOT NULL DEFAULT 'pending',    -- pending, running, paused, completed
	started_at_ms INTEGER,                     -- When timer started (epoch ms)
	paused_at_ms INTEGER,                      -- When timer was paused (epoch ms)
	completed_at_ms INTEGER,                   -- When timer completed (epoch ms)
	deleted_at INTEGER,                        -- Soft delete timestamp (epoch seconds)
	updated_at_ms INTEGER NOT NULL             -- Last update timestamp (epoch ms)
);

CREATE INDEX IF NOT EXISTS idx_timers_status ON timers(status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_timers_deleted_at ON timers(deleted_at);
`

// stopwatchesTable stores stopwatch data
const stopwatchesTable = `
CREATE TABLE IF NOT EXISTS stopwatches (
	id INTEGER PRIMARY KEY,                    -- Unix epoch timestamp in seconds
	label TEXT NOT NULL,                       -- Stopwatch name/description
	elapsed_ms INTEGER NOT NULL DEFAULT 0,     -- Total elapsed time in milliseconds
	status TEXT NOT NULL DEFAULT 'stopped',    -- stopped, running
	started_at_ms INTEGER,                     -- When stopwatch started (epoch ms)
	stopped_at_ms INTEGER,                     -- When stopwatch stopped (epoch ms)
	deleted_at INTEGER,                        -- Soft delete timestamp (epoch seconds)
	updated_at_ms INTEGER NOT NULL             -- Last update timestamp (epoch ms)
);

CREATE INDEX IF NOT EXISTS idx_stopwatches_status ON stopwatches(status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_stopwatches_deleted_at ON stopwatches(deleted_at);
`

// stopwatchLapsTable stores individual lap records for stopwatches
const stopwatchLapsTable = `
CREATE TABLE IF NOT EXISTS stopwatch_laps (
	id INTEGER PRIMARY KEY AUTOINCREMENT,      -- Auto-increment ID
	stopwatch_id INTEGER NOT NULL,             -- Reference to stopwatches.id (epoch seconds)
	lap_number INTEGER NOT NULL,               -- Sequential lap number (1, 2, 3...)
	lap_duration_ms INTEGER NOT NULL,          -- Duration of this specific lap
	total_elapsed_ms INTEGER NOT NULL,         -- Total elapsed time at lap completion
	created_at_ms INTEGER NOT NULL,            -- When lap was created (epoch ms)
	
	FOREIGN KEY (stopwatch_id) REFERENCES stopwatches(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_laps_stopwatch_id ON stopwatch_laps(stopwatch_id);
CREATE INDEX IF NOT EXISTS idx_laps_lap_number ON stopwatch_laps(stopwatch_id, lap_number);
`

// journalEntriesTable stores journal log entries
const journalEntriesTable = `
CREATE TABLE IF NOT EXISTS journal_entries (
	id INTEGER PRIMARY KEY,                    -- Unix epoch timestamp in seconds
	log TEXT NOT NULL,                         -- Journal entry content
	deleted_at INTEGER,                        -- Soft delete timestamp (epoch seconds)
	created_at_ms INTEGER NOT NULL             -- Creation timestamp (epoch ms)
);

CREATE INDEX IF NOT EXISTS idx_journal_deleted_at ON journal_entries(deleted_at);
CREATE INDEX IF NOT EXISTS idx_journal_created_at ON journal_entries(created_at_ms);
`

// scheduledTasksTable stores recurring/scheduled tasks (future feature)
const scheduledTasksTable = `
CREATE TABLE IF NOT EXISTS scheduled_tasks (
	id INTEGER PRIMARY KEY,                    -- Unix epoch timestamp in seconds
	task_name TEXT NOT NULL,                   -- Task identifier/name
	schedule_type TEXT NOT NULL,               -- cron, interval, once
	schedule_data TEXT NOT NULL,               -- JSON: cron expression or interval config
	next_due_at_ms INTEGER,                    -- Next scheduled run time (epoch ms)
	last_completed_at_ms INTEGER,              -- Last completion time (epoch ms)
	deleted_at INTEGER,                        -- Soft delete timestamp (epoch seconds)
	created_at_ms INTEGER NOT NULL,            -- Creation timestamp (epoch ms)
	updated_at_ms INTEGER NOT NULL             -- Last update timestamp (epoch ms)
);

CREATE INDEX IF NOT EXISTS idx_scheduled_tasks_next_due ON scheduled_tasks(next_due_at_ms) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_scheduled_tasks_deleted_at ON scheduled_tasks(deleted_at);
`

// settingsTable stores application settings as key-value pairs
const settingsTable = `
CREATE TABLE IF NOT EXISTS settings (
	key TEXT PRIMARY KEY,                      -- Setting identifier
	value TEXT NOT NULL,                       -- Setting value (JSON or string)
	updated_at_ms INTEGER NOT NULL             -- Last update timestamp (epoch ms)
);
`

// citiesTable stores geographic locations for starfield observation
const citiesTable = `
CREATE TABLE IF NOT EXISTS cities (
	id INTEGER PRIMARY KEY AUTOINCREMENT,      -- Auto-increment ID
	country TEXT NOT NULL,                     -- Country name
	city TEXT NOT NULL,                        -- City name
	latitude REAL NOT NULL,                    -- Latitude in degrees (-90 to 90)
	longitude REAL NOT NULL,                   -- Longitude in degrees (-180 to 180)
	timezone TEXT NOT NULL,                    -- IANA timezone identifier
	UNIQUE(country, city)
);

CREATE INDEX IF NOT EXISTS idx_cities_country ON cities(country);
CREATE INDEX IF NOT EXISTS idx_cities_location ON cities(latitude, longitude);
`

// seedCities populates the cities table with major world cities
func seedCities(db *sql.DB) error {
	// Check if cities already exist
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM cities").Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil // Already seeded
	}

	cities := []struct {
		country   string
		city      string
		latitude  float64
		longitude float64
		timezone  string
	}{
		// India (default)
		{"India", "Mumbai", 19.0760, 72.8777, "Asia/Kolkata"},
		{"India", "Delhi", 28.6139, 77.2090, "Asia/Kolkata"},
		{"India", "Bangalore", 12.9716, 77.5946, "Asia/Kolkata"},

		// United States
		{"United States", "New York", 40.7128, -74.0060, "America/New_York"},
		{"United States", "Los Angeles", 34.0522, -118.2437, "America/Los_Angeles"},
		{"United States", "Chicago", 41.8781, -87.6298, "America/Chicago"},

		// United Kingdom
		{"United Kingdom", "London", 51.5074, -0.1278, "Europe/London"},
		{"United Kingdom", "Manchester", 53.4808, -2.2426, "Europe/London"},
		{"United Kingdom", "Edinburgh", 55.9533, -3.1883, "Europe/London"},

		// Germany
		{"Germany", "Berlin", 52.5200, 13.4050, "Europe/Berlin"},
		{"Germany", "Munich", 48.1351, 11.5820, "Europe/Berlin"},
		{"Germany", "Hamburg", 53.5511, 9.9937, "Europe/Berlin"},

		// France
		{"France", "Paris", 48.8566, 2.3522, "Europe/Paris"},
		{"France", "Lyon", 45.7640, 4.8357, "Europe/Paris"},
		{"France", "Marseille", 43.2965, 5.3698, "Europe/Paris"},

		// Japan
		{"Japan", "Tokyo", 35.6762, 139.6503, "Asia/Tokyo"},
		{"Japan", "Osaka", 34.6937, 135.5023, "Asia/Tokyo"},
		{"Japan", "Kyoto", 35.0116, 135.7681, "Asia/Tokyo"},

		// Australia
		{"Australia", "Sydney", -33.8688, 151.2093, "Australia/Sydney"},
		{"Australia", "Melbourne", -37.8136, 144.9631, "Australia/Melbourne"},
		{"Australia", "Perth", -31.9505, 115.8605, "Australia/Perth"},

		// Brazil
		{"Brazil", "São Paulo", -23.5505, -46.6333, "America/Sao_Paulo"},
		{"Brazil", "Rio de Janeiro", -22.9068, -43.1729, "America/Sao_Paulo"},
		{"Brazil", "Brasília", -15.7975, -47.8919, "America/Sao_Paulo"},

		// Canada
		{"Canada", "Toronto", 43.6532, -79.3832, "America/Toronto"},
		{"Canada", "Vancouver", 49.2827, -123.1207, "America/Vancouver"},
		{"Canada", "Montreal", 45.5017, -73.5673, "America/Toronto"},

		// China
		{"China", "Beijing", 39.9042, 116.4074, "Asia/Shanghai"},
		{"China", "Shanghai", 31.2304, 121.4737, "Asia/Shanghai"},
		{"China", "Hong Kong", 22.3193, 114.1694, "Asia/Hong_Kong"},

		// Russia
		{"Russia", "Moscow", 55.7558, 37.6173, "Europe/Moscow"},
		{"Russia", "Saint Petersburg", 59.9311, 30.3609, "Europe/Moscow"},
		{"Russia", "Novosibirsk", 55.0084, 82.9357, "Asia/Novosibirsk"},

		// South Africa
		{"South Africa", "Cape Town", -33.9249, 18.4241, "Africa/Johannesburg"},
		{"South Africa", "Johannesburg", -26.2041, 28.0473, "Africa/Johannesburg"},
		{"South Africa", "Durban", -29.8587, 31.0218, "Africa/Johannesburg"},

		// Egypt
		{"Egypt", "Cairo", 30.0444, 31.2357, "Africa/Cairo"},
		{"Egypt", "Alexandria", 31.2001, 29.9187, "Africa/Cairo"},

		// Mexico
		{"Mexico", "Mexico City", 19.4326, -99.1332, "America/Mexico_City"},
		{"Mexico", "Guadalajara", 20.6597, -103.3496, "America/Mexico_City"},
		{"Mexico", "Monterrey", 25.6866, -100.3161, "America/Monterrey"},

		// Italy
		{"Italy", "Rome", 41.9028, 12.4964, "Europe/Rome"},
		{"Italy", "Milan", 45.4642, 9.1900, "Europe/Rome"},
		{"Italy", "Naples", 40.8518, 14.2681, "Europe/Rome"},

		// Spain
		{"Spain", "Madrid", 40.4168, -3.7038, "Europe/Madrid"},
		{"Spain", "Barcelona", 41.3851, 2.1734, "Europe/Madrid"},
		{"Spain", "Valencia", 39.4699, -0.3763, "Europe/Madrid"},

		// Argentina
		{"Argentina", "Buenos Aires", -34.6037, -58.3816, "America/Argentina/Buenos_Aires"},
		{"Argentina", "Córdoba", -31.4201, -64.1888, "America/Argentina/Cordoba"},

		// Turkey
		{"Turkey", "Istanbul", 41.0082, 28.9784, "Europe/Istanbul"},
		{"Turkey", "Ankara", 39.9334, 32.8597, "Europe/Istanbul"},
		{"Turkey", "Izmir", 38.4237, 27.1428, "Europe/Istanbul"},

		// South Korea
		{"South Korea", "Seoul", 37.5665, 126.9780, "Asia/Seoul"},
		{"South Korea", "Busan", 35.1796, 129.0756, "Asia/Seoul"},

		// Indonesia
		{"Indonesia", "Jakarta", -6.2088, 106.8456, "Asia/Jakarta"},
		{"Indonesia", "Surabaya", -7.2575, 112.7521, "Asia/Jakarta"},
		{"Indonesia", "Bali", -8.4095, 115.1889, "Asia/Makassar"},

		// Saudi Arabia
		{"Saudi Arabia", "Riyadh", 24.7136, 46.6753, "Asia/Riyadh"},
		{"Saudi Arabia", "Jeddah", 21.4858, 39.1925, "Asia/Riyadh"},
		{"Saudi Arabia", "Mecca", 21.3891, 39.8579, "Asia/Riyadh"},

		// UAE
		{"UAE", "Dubai", 25.2048, 55.2708, "Asia/Dubai"},
		{"UAE", "Abu Dhabi", 24.4539, 54.3773, "Asia/Dubai"},

		// Singapore
		{"Singapore", "Singapore", 1.3521, 103.8198, "Asia/Singapore"},

		// Thailand
		{"Thailand", "Bangkok", 13.7563, 100.5018, "Asia/Bangkok"},
		{"Thailand", "Chiang Mai", 18.7883, 98.9853, "Asia/Bangkok"},

		// Malaysia
		{"Malaysia", "Kuala Lumpur", 3.1390, 101.6869, "Asia/Kuala_Lumpur"},
		{"Malaysia", "Penang", 5.4164, 100.3327, "Asia/Kuala_Lumpur"},

		// Philippines
		{"Philippines", "Manila", 14.5995, 120.9842, "Asia/Manila"},
		{"Philippines", "Cebu", 10.3157, 123.8854, "Asia/Manila"},

		// New Zealand
		{"New Zealand", "Auckland", -36.8485, 174.7633, "Pacific/Auckland"},
		{"New Zealand", "Wellington", -41.2865, 174.7762, "Pacific/Auckland"},

		// Chile
		{"Chile", "Santiago", -33.4489, -70.6693, "America/Santiago"},
		{"Chile", "Valparaíso", -33.0472, -71.6127, "America/Santiago"},

		// Netherlands
		{"Netherlands", "Amsterdam", 52.3676, 4.9041, "Europe/Amsterdam"},
		{"Netherlands", "Rotterdam", 51.9244, 4.4777, "Europe/Amsterdam"},

		// Switzerland
		{"Switzerland", "Zurich", 47.3769, 8.5417, "Europe/Zurich"},
		{"Switzerland", "Geneva", 46.2044, 6.1432, "Europe/Zurich"},

		// Sweden
		{"Sweden", "Stockholm", 59.3293, 18.0686, "Europe/Stockholm"},
		{"Sweden", "Gothenburg", 57.7089, 11.9746, "Europe/Stockholm"},

		// Norway
		{"Norway", "Oslo", 59.9139, 10.7522, "Europe/Oslo"},
		{"Norway", "Bergen", 60.3913, 5.3221, "Europe/Oslo"},

		// Denmark
		{"Denmark", "Copenhagen", 55.6761, 12.5683, "Europe/Copenhagen"},
		{"Denmark", "Aarhus", 56.1629, 10.2039, "Europe/Copenhagen"},

		// Finland
		{"Finland", "Helsinki", 60.1699, 24.9384, "Europe/Helsinki"},

		// Poland
		{"Poland", "Warsaw", 52.2297, 21.0122, "Europe/Warsaw"},
		{"Poland", "Krakow", 50.0647, 19.9450, "Europe/Warsaw"},

		// Austria
		{"Austria", "Vienna", 48.2082, 16.3738, "Europe/Vienna"},
		{"Austria", "Salzburg", 47.8095, 13.0550, "Europe/Vienna"},

		// Belgium
		{"Belgium", "Brussels", 50.8476, 4.3572, "Europe/Brussels"},

		// Portugal
		{"Portugal", "Lisbon", 38.7223, -9.1393, "Europe/Lisbon"},
		{"Portugal", "Porto", 41.1579, -8.6291, "Europe/Lisbon"},

		// Greece
		{"Greece", "Athens", 37.9838, 23.7275, "Europe/Athens"},
		{"Greece", "Thessaloniki", 40.6401, 22.9444, "Europe/Athens"},

		// Czech Republic
		{"Czech Republic", "Prague", 50.0755, 14.4378, "Europe/Prague"},

		// Hungary
		{"Hungary", "Budapest", 47.4979, 19.0402, "Europe/Budapest"},

		// Ireland
		{"Ireland", "Dublin", 53.3498, -6.2603, "Europe/Dublin"},
		{"Ireland", "Cork", 51.8985, -8.4756, "Europe/Dublin"},

		// Israel
		{"Israel", "Tel Aviv", 32.0853, 34.7818, "Asia/Jerusalem"},
		{"Israel", "Jerusalem", 31.7683, 35.2137, "Asia/Jerusalem"},

		// Qatar
		{"Qatar", "Doha", 25.2854, 51.5310, "Asia/Qatar"},

		// Kuwait
		{"Kuwait", "Kuwait City", 29.3759, 47.9774, "Asia/Kuwait"},

		// Nigeria
		{"Nigeria", "Lagos", 6.5244, 3.3792, "Africa/Lagos"},
		{"Nigeria", "Abuja", 9.0765, 7.3986, "Africa/Lagos"},

		// Kenya
		{"Kenya", "Nairobi", -1.2921, 36.8219, "Africa/Nairobi"},
		{"Kenya", "Mombasa", -4.0435, 39.6682, "Africa/Nairobi"},

		// Ethiopia
		{"Ethiopia", "Addis Ababa", 9.1450, 40.4897, "Africa/Addis_Ababa"},

		// Morocco
		{"Morocco", "Casablanca", 33.5731, -7.5898, "Africa/Casablanca"},
		{"Morocco", "Marrakech", 31.6295, -7.9811, "Africa/Casablanca"},

		// Peru
		{"Peru", "Lima", -12.0464, -77.0428, "America/Lima"},
		{"Peru", "Cusco", -13.1631, -72.5450, "America/Lima"},

		// Colombia
		{"Colombia", "Bogotá", 4.7110, -74.0721, "America/Bogota"},
		{"Colombia", "Medellín", 6.2476, -75.5658, "America/Bogota"},

		// Venezuela
		{"Venezuela", "Caracas", 10.4806, -66.9036, "America/Caracas"},

		// Ecuador
		{"Ecuador", "Quito", -0.1807, -78.4678, "America/Guayaquil"},

		// Uruguay
		{"Uruguay", "Montevideo", -34.9011, -56.1645, "America/Montevideo"},

		// Paraguay
		{"Paraguay", "Asunción", -25.2637, -57.5759, "America/Asuncion"},

		// Bolivia
		{"Bolivia", "La Paz", -16.5000, -68.1500, "America/La_Paz"},

		// Costa Rica
		{"Costa Rica", "San José", 9.9281, -84.0907, "America/Costa_Rica"},

		// Panama
		{"Panama", "Panama City", 8.9824, -79.5199, "America/Panama"},

		// Cuba
		{"Cuba", "Havana", 23.1136, -82.3666, "America/Havana"},

		// Dominican Republic
		{"Dominican Republic", "Santo Domingo", 18.4861, -69.9312, "America/Santo_Domingo"},

		// Jamaica
		{"Jamaica", "Kingston", 17.9712, -76.7926, "America/Jamaica"},

		// Trinidad and Tobago
		{"Trinidad and Tobago", "Port of Spain", 10.6549, -61.5019, "America/Port_of_Spain"},

		// Iceland
		{"Iceland", "Reykjavik", 64.1466, -21.9426, "Atlantic/Reykjavik"},

		// Croatia
		{"Croatia", "Zagreb", 45.8150, 15.9819, "Europe/Zagreb"},
		{"Croatia", "Split", 43.5081, 16.4402, "Europe/Zagreb"},

		// Serbia
		{"Serbia", "Belgrade", 44.7866, 20.4489, "Europe/Belgrade"},

		// Romania
		{"Romania", "Bucharest", 44.4268, 26.1025, "Europe/Bucharest"},

		// Bulgaria
		{"Bulgaria", "Sofia", 42.6977, 23.3219, "Europe/Sofia"},

		// Ukraine
		{"Ukraine", "Kyiv", 50.4501, 30.5234, "Europe/Kyiv"},
		{"Ukraine", "Lviv", 49.8397, 24.0297, "Europe/Kyiv"},

		// Belarus
		{"Belarus", "Minsk", 53.9045, 27.5615, "Europe/Minsk"},

		// Kazakhstan
		{"Kazakhstan", "Almaty", 43.2220, 76.8512, "Asia/Almaty"},
		{"Kazakhstan", "Astana", 51.1699, 71.4491, "Asia/Almaty"},

		// Uzbekistan
		{"Uzbekistan", "Tashkent", 41.2995, 69.2401, "Asia/Tashkent"},

		// Vietnam
		{"Vietnam", "Ho Chi Minh City", 10.8231, 106.6297, "Asia/Ho_Chi_Minh"},
		{"Vietnam", "Hanoi", 21.0285, 105.8542, "Asia/Bangkok"},

		// Cambodia
		{"Cambodia", "Phnom Penh", 11.5564, 104.9282, "Asia/Phnom_Penh"},

		// Myanmar
		{"Myanmar", "Yangon", 16.8661, 96.1951, "Asia/Yangon"},

		// Bangladesh
		{"Bangladesh", "Dhaka", 23.8103, 90.4125, "Asia/Dhaka"},
		{"Bangladesh", "Chittagong", 22.3569, 91.7832, "Asia/Dhaka"},

		// Pakistan
		{"Pakistan", "Karachi", 24.8607, 67.0011, "Asia/Karachi"},
		{"Pakistan", "Lahore", 31.5204, 74.3587, "Asia/Karachi"},
		{"Pakistan", "Islamabad", 33.6844, 73.0479, "Asia/Karachi"},

		// Sri Lanka
		{"Sri Lanka", "Colombo", 6.9271, 79.8612, "Asia/Colombo"},

		// Nepal
		{"Nepal", "Kathmandu", 27.7172, 85.3240, "Asia/Kathmandu"},

		// Bhutan
		{"Bhutan", "Thimphu", 27.4728, 89.6390, "Asia/Thimphu"},

		// Maldives
		{"Maldives", "Malé", 4.1755, 73.5093, "Indian/Maldives"},
	}

	stmt, err := db.Prepare(`
		INSERT INTO cities (country, city, latitude, longitude, timezone)
		VALUES (?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, c := range cities {
		if _, err := stmt.Exec(c.country, c.city, c.latitude, c.longitude, c.timezone); err != nil {
			return err
		}
	}

	return nil
}
