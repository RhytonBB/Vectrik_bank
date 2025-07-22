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
	// USERS
	router.HandleFunc("/api/admin/users", AdminUsersAPIHandler).Methods("GET", "POST")
	router.HandleFunc("/api/admin/users/{id}", AdminUserDetailAPIHandler).Methods("GET")
	router.HandleFunc("/api/admin/users/{id}/activate", AdminUserActivateAPIHandler).Methods("POST")
	router.HandleFunc("/api/admin/users/{id}/admin", AdminUserAdminAPIHandler).Methods("POST")
	router.HandleFunc("/api/admin/users/{id}/ban", AdminUserBanAPIHandler).Methods("POST")
	router.HandleFunc("/api/admin/users/{id}/unban", AdminUserUnbanAPIHandler).Methods("POST")
	router.HandleFunc("/api/admin/users/{id}", AdminUserEditDeleteAPIHandler).Methods("PUT", "DELETE")
	router.HandleFunc("/api/admin/users/adjust_balance", AdminUserAdjustBalanceAPIHandler).Methods("POST")
	// BULK USERS
	router.HandleFunc("/api/admin/users/bulk/preview", AdminBulkUsersPreviewAPIHandler).Methods("POST")
	router.HandleFunc("/api/admin/users/bulk", AdminBulkUsersRegisterAPIHandler).Methods("POST")

	// PRODUCTS
	router.HandleFunc("/api/admin/products", AdminProductsAPIHandler).Methods("GET", "POST")
	router.HandleFunc("/api/admin/products/{id}", AdminProductDetailAPIHandler).Methods("GET", "PUT", "DELETE")
	router.HandleFunc("/api/admin/products/upload", AdminProductUploadAPIHandler).Methods("POST")

	// BANNERS
	router.HandleFunc("/api/admin/banners", AdminBannersAPIHandler).Methods("GET", "POST")
	router.HandleFunc("/api/admin/banners/{id}/activate", AdminBannerActivateAPIHandler).Methods("POST")
	router.HandleFunc("/api/admin/banners/{id}", AdminBannerDetailAPIHandler).Methods("PUT", "DELETE")

	// ANALYTICS
	router.HandleFunc("/api/admin/analytics", AdminAnalyticsAPIHandler).Methods("GET")
	// BALANCE
	router.HandleFunc("/api/admin/balance", AdminBalanceAPIHandler).Methods("GET")
	// REWARD USERS
	router.HandleFunc("/api/admin/reward/users", AdminRewardUsersAPIHandler).Methods("GET")
	router.HandleFunc("/api/admin/reward", AdminRewardAPIHandler).Methods("POST")
	
	// BROADCAST
	router.HandleFunc("/api/admin/broadcast", AdminBroadcastAPIHandler).Methods("GET", "POST")
	router.HandleFunc("/api/admin/broadcast/{id}/history", AdminBroadcastHistoryAPIHandler).Methods("GET")
	router.HandleFunc("/api/admin/broadcast/users", AdminBroadcastUsersAPIHandler).Methods("GET")
	
	// LOGIN
	router.HandleFunc("/api/login", LoginHandler).Methods("POST")
	router.HandleFunc("/api/me", MeAPIHandler).Methods("GET")

	router.HandleFunc("/logout", LogoutHandler)

	router.HandleFunc("/api/admin/theme", AdminThemeGetHandler).Methods("GET")
	router.HandleFunc("/api/admin/theme", AdminThemeSetHandler).Methods("POST")

	http.Handle("/", router)
}
