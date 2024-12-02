package main

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/go-chi/chi"
)

func handle(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello from API"))
}

func main() {
	r := chi.NewRouter()

	// Middleware для обработки запросов
	r.Use(middleware)

	// Определяем маршрут для API
	r.Get("/api/", handle)

	// Запуск сервера на порту 8080
	http.ListenAndServe(":8080", r)
}

func middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/" {
			// Прокси для всех остальных запросов
			proxyURL, _ := url.Parse("http://hugo:1313")
			proxy := httputil.NewSingleHostReverseProxy(proxyURL)
			proxy.ServeHTTP(w, r)
			return
		}
		// Если это запрос к /api/, передаем его дальше
		next.ServeHTTP(w, r)
	})
}
