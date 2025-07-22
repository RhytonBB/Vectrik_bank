package services

import (
	"sync"
	"github.com/RhytonBB/vectrik-webbank/internal/db"
)

var (
	cachedTheme     = "green"
	cachedThemeOnce sync.Once
	cachedThemeMu   sync.RWMutex
)

func initThemeCache() {
	var setting db.Setting
	err := db.DB.First(&setting, "key = ?", "user_theme").Error
	if err == nil {
		cachedTheme = setting.Value
	}
}

func GetTheme() string {
	cachedThemeOnce.Do(initThemeCache)
	cachedThemeMu.RLock()
	theme := cachedTheme
	cachedThemeMu.RUnlock()
	return theme
}

func SetTheme(theme string) {
	cachedThemeMu.Lock()
	cachedTheme = theme
	cachedThemeMu.Unlock()
} 