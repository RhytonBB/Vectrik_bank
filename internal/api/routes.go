package api

import (
	"net/http"
	"github.com/gorilla/mux"
)

func RegisterRoutes() {
	router := mux.NewRouter()

	router.HandleFunc("/login", LoginHandler)
	router.HandleFunc("/dashboard", DashboardHandler)

	router.HandleFunc("/admin/reward", AdminRewardHandler)
	router.HandleFunc("/admin/users", AdminUsersHandler)
	router.HandleFunc("/admin/users/register", AdminUserRegisterHandler)
	router.HandleFunc("/admin/users/{id}", AdminUserDetailHandler)

	router.HandleFunc("/admin/banners", AdminBannersHandler)
	router.HandleFunc("/admin/products", AdminProductsHandler)
	router.HandleFunc("/admin/analytics", AdminAnalyticsHandler)

	router.HandleFunc("/logout", LogoutHandler)

	http.Handle("/", router)
}
