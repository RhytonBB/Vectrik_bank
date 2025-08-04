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
	router.HandleFunc("/admin/analytics", AdminAnalyticsAPIHandler)
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
	
	// NEWS (ADMIN)
	router.HandleFunc("/api/admin/news", GetNewsList).Methods("GET")
	router.HandleFunc("/api/admin/news", CreateNews).Methods("POST")
	router.HandleFunc("/api/admin/news/{id}", GetNewsById).Methods("GET")
	router.HandleFunc("/api/admin/news/{id}", UpdateNews).Methods("PUT")
	router.HandleFunc("/api/admin/news/{id}/publish", ToggleNewsPublish).Methods("POST")
	router.HandleFunc("/api/admin/news/{id}", DeleteNews).Methods("DELETE")
	
	// LOGIN
	router.HandleFunc("/api/login", LoginHandler).Methods("POST")
	router.HandleFunc("/api/me", MeAPIHandler).Methods("GET")

	router.HandleFunc("/logout", LogoutHandler)
	
	// CLIENT API
	router.HandleFunc("/api/client/banners", ClientBannersAPIHandler).Methods("GET")
	router.HandleFunc("/api/client/products", ClientProductsAPIHandler).Methods("GET")
	router.HandleFunc("/api/client/purchase", ClientPurchaseProductAPIHandler).Methods("POST")
	router.HandleFunc("/api/client/purchases", ClientUserPurchasesAPIHandler).Methods("GET")
	router.HandleFunc("/api/client/transactions", ClientUserTransactionsAPIHandler).Methods("GET")
	router.HandleFunc("/api/client/theme", ClientThemeGetHandler).Methods("GET")
	
	// NEWS (CLIENT)
	router.HandleFunc("/api/news", GetPublishedNews).Methods("GET")
	router.HandleFunc("/api/news/latest", GetLatestNews).Methods("GET")
	router.HandleFunc("/api/news/{id}", GetPublicNewsById).Methods("GET")
	router.HandleFunc("/api/news/{id}/rate", RateNews).Methods("POST")
	router.HandleFunc("/api/news/{id}/rating", GetUserRating).Methods("GET")

	router.HandleFunc("/api/admin/theme", AdminThemeGetHandler).Methods("GET")
	router.HandleFunc("/api/admin/theme", AdminThemeSetHandler).Methods("POST")
	router.HandleFunc("/api/admin/groups", AdminGroupsAPIHandler).Methods("GET", "POST")
	router.HandleFunc("/api/admin/groups/{id}", AdminGroupsAPIHandler).Methods("DELETE")

	http.Handle("/", router)
}
