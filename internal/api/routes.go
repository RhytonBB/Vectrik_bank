package api

import "net/http"

func RegisterRoutes() {
	http.HandleFunc("/login", LoginHandler)
	http.HandleFunc("/dashboard", DashboardHandler)
}
