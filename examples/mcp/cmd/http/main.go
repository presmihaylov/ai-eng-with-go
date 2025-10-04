package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// API key for bearer token authentication
const apiKey = "Rd26ZK/TmNFinnc354BarseyF96sCT7PoFXhtAVVH6E="

type GetWeatherInput struct {
	City string `json:"city" jsonschema:"the name of the city to get weather for"`
}

type WeatherOutput struct {
	City        string `json:"city" jsonschema:"the city name"`
	Temperature string `json:"temperature" jsonschema:"current temperature"`
	Condition   string `json:"condition" jsonschema:"weather condition"`
	Humidity    string `json:"humidity" jsonschema:"humidity percentage"`
	WindSpeed   string `json:"wind_speed" jsonschema:"wind speed"`
	Description string `json:"description" jsonschema:"full weather description"`
}

// GeocodeResponse represents the response from Open-Meteo geocoding API
type GeocodeResponse struct {
	Results []struct {
		Name      string  `json:"name"`
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
		Country   string  `json:"country"`
	} `json:"results"`
}

// WeatherResponse represents the response from Open-Meteo weather API
type WeatherResponse struct {
	CurrentWeather struct {
		Temperature float64 `json:"temperature"`
		WindSpeed   float64 `json:"windspeed"`
		WeatherCode int     `json:"weathercode"`
		Time        string  `json:"time"`
	} `json:"current_weather"`
	CurrentUnits struct {
		Temperature string `json:"temperature"`
		WindSpeed   string `json:"windspeed"`
	} `json:"current_weather_units"`
	Hourly struct {
		RelativeHumidity []int `json:"relative_humidity_2m"`
	} `json:"hourly"`
}

func getWeatherCondition(code int) string {
	conditions := map[int]string{
		0: "Clear sky", 1: "Mainly clear", 2: "Partly cloudy", 3: "Overcast",
		45: "Foggy", 48: "Depositing rime fog",
		51: "Light drizzle", 53: "Moderate drizzle", 55: "Dense drizzle",
		61: "Slight rain", 63: "Moderate rain", 65: "Heavy rain",
		71: "Slight snow", 73: "Moderate snow", 75: "Heavy snow",
		77: "Snow grains", 80: "Slight rain showers", 81: "Moderate rain showers",
		82: "Violent rain showers", 85: "Slight snow showers", 86: "Heavy snow showers",
		95: "Thunderstorm", 96: "Thunderstorm with slight hail", 99: "Thunderstorm with heavy hail",
	}
	if condition, ok := conditions[code]; ok {
		return condition
	}
	return "Unknown"
}

func GetWeather(ctx context.Context, req *mcp.CallToolRequest, input GetWeatherInput) (*mcp.CallToolResult, WeatherOutput, error) {
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Step 1: Geocode the city name to get coordinates
	geocodeURL := fmt.Sprintf("https://geocoding-api.open-meteo.com/v1/search?name=%s&count=1&language=en&format=json", input.City)

	resp, err := httpClient.Get(geocodeURL)
	if err != nil {
		return nil, WeatherOutput{}, fmt.Errorf("failed to geocode city: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, WeatherOutput{}, fmt.Errorf("geocoding API returned status %d", resp.StatusCode)
	}

	var geocodeResp GeocodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&geocodeResp); err != nil {
		return nil, WeatherOutput{}, fmt.Errorf("failed to decode geocoding data: %w", err)
	}

	if len(geocodeResp.Results) == 0 {
		return nil, WeatherOutput{}, fmt.Errorf("city not found: %s", input.City)
	}

	location := geocodeResp.Results[0]

	// Step 2: Get weather data using coordinates
	weatherURL := fmt.Sprintf("https://api.open-meteo.com/v1/forecast?latitude=%.4f&longitude=%.4f&current_weather=true&hourly=relative_humidity_2m&timezone=auto",
		location.Latitude, location.Longitude)

	resp, err = httpClient.Get(weatherURL)
	if err != nil {
		return nil, WeatherOutput{}, fmt.Errorf("failed to fetch weather data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, WeatherOutput{}, fmt.Errorf("weather API returned status %d", resp.StatusCode)
	}

	var weatherResp WeatherResponse
	if err := json.NewDecoder(resp.Body).Decode(&weatherResp); err != nil {
		return nil, WeatherOutput{}, fmt.Errorf("failed to decode weather data: %w", err)
	}

	current := weatherResp.CurrentWeather
	condition := getWeatherCondition(current.WeatherCode)

	// Get current humidity (first hourly value)
	humidity := 0
	if len(weatherResp.Hourly.RelativeHumidity) > 0 {
		humidity = weatherResp.Hourly.RelativeHumidity[0]
	}

	// Convert temperature to Fahrenheit
	tempF := current.Temperature*9/5 + 32

	output := WeatherOutput{
		City:        fmt.Sprintf("%s, %s", location.Name, location.Country),
		Temperature: fmt.Sprintf("%.1f°C (%.1f°F)", current.Temperature, tempF),
		Condition:   condition,
		Humidity:    fmt.Sprintf("%d%%", humidity),
		WindSpeed:   fmt.Sprintf("%.1f km/h", current.WindSpeed),
		Description: fmt.Sprintf("The weather in %s is currently %s with a temperature of %.1f°C. Humidity is at %d%% and wind speed is %.1f km/h.",
			location.Name, condition, current.Temperature, humidity, current.WindSpeed),
	}

	return nil, output, nil
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			log.Printf("[ERROR] Unauthorized request from %s: missing Authorization header", r.RemoteAddr)
			http.Error(w, "Unauthorized: missing Authorization header", http.StatusUnauthorized)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			log.Printf("[ERROR] Unauthorized request from %s: invalid Authorization header format", r.RemoteAddr)
			http.Error(w, "Unauthorized: invalid Authorization header format", http.StatusUnauthorized)
			return
		}

		token := parts[1]
		if token != apiKey {
			log.Printf("[ERROR] Unauthorized request from %s: invalid token", r.RemoteAddr)
			http.Error(w, "Unauthorized: invalid token", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	port := flag.Int("port", 8080, "port to listen on")
	host := flag.String("host", "localhost", "host to listen on")
	flag.Parse()

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "weather-server",
		Version: "1.0.0",
	}, nil)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "getWeather",
		Description: "Get the current weather for a specified city",
	}, GetWeather)

	handler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server {
		return server
	}, nil)

	// Wrap handler with authentication middleware
	authHandler := authMiddleware(handler)

	addr := fmt.Sprintf("%s:%d", *host, *port)
	log.Printf("[INFO] Starting MCP weather HTTP server on %s with authentication...", addr)
	log.Printf("[INFO] API Key: %s", apiKey)

	if err := http.ListenAndServe(addr, authHandler); err != nil {
		log.Fatalf("[ERROR] Server failed: %v", err)
	}
}
