package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

const dadataAPIkey = `d9e0649452a137b73d941aa4fb4fcac859372c8c`

type RequestAddressSearch struct {
	Query string `json:"query"`
}

type ResponseAddress struct {
	Addresses []*Address `json:"addresses"`
}

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	// Middleware для перенаправления
	r.Use(proxyMiddleware)

	r.Route("/api", func(r chi.Router) {
		r.Post("/address/geocode", addressGeocode)
		r.Post("/address/search", addressSearch)
	})

	http.ListenAndServe(":8080", r)
}

func proxyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api") { // Проверяем префикс
			w.WriteHeader(http.StatusOK)      // Устанавливаем статус 200 OK
			w.Write([]byte("Hello from API")) // Отправляем текст
			return
		}
		// Перенаправление на hugo
		proxyURL, _ := url.Parse("http://hugo:1313")
		proxy := httputil.NewSingleHostReverseProxy(proxyURL)
		proxy.ServeHTTP(w, r)
	})
}

func addressSearch(w http.ResponseWriter, r *http.Request) {
	var req RequestAddressSearch
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	addresses, err := searchAddress(req.Query)
	if err != nil {
		http.Error(w, "Service unavailable: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ResponseAddress{Addresses: addresses})
}

func addressGeocode(w http.ResponseWriter, r *http.Request) {
	var geocode Address
	if err := json.NewDecoder(r.Body).Decode(&geocode); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	addresses, err := geocodeAddress(geocode.Lat, geocode.Lon)
	if err != nil {
		http.Error(w, "Service unavailable: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"data": addresses})
}

func searchAddress(query string) ([]*Address, error) {
	url := "https://suggestions.dadata.ru/suggestions/api/4_1/rs/suggest/address"
	body := map[string]string{"query": query}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Token "+dadataAPIkey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Dadata service returned status: %s", resp.Status)
	}

	var result struct {
		Suggestions []struct {
			Value string  `json:"value"`
			Data  Address `json:"data"`
		} `json:"suggestions"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var addresses []*Address
	for _, suggestion := range result.Suggestions {
		addresses = append(addresses, &suggestion.Data)
	}

	return addresses, nil
}

func geocodeAddress(lat, lon string) ([]*Address, error) {
	url := "https://dadata.ru/api/v2/geocode"
	body := map[string]string{"lat": lat, "lon": lon}
	jsonBody, err := json.Marshal(body) // Обработка ошибки
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Token "+dadataAPIkey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Проверяем статус ответа
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Dadata service returned status: %s", resp.Status)
	}
	var result struct {
		Addresses []Address `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	// Декодируем ответ

	var addresses []*Address
	for i := range result.Addresses {
		addresses = append(addresses, &result.Addresses[i]) // Добавляем указатели на адреса
	}

	// Возвращаем адреса
	return addresses, nil // Исправлено на result.Addresses
}
