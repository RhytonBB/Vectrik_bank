# 📋 Итоговые изменения для деплоя с реверс прокси

## 🔧 Изменения в коде

### Backend изменения:

1. **Cookie domain** в `internal/api/auth_handlers.go`:
   ```go
   // Было:
   Domain: ".vectrikbank.ru",
   
   // Стало:
   Domain: "vectrikbank.ru",
   ```

### Frontend изменения:

1. **Убран прокси** в `vektrik-frontend/vue.config.js`:
   ```javascript
   // Было:
   devServer: {
     proxy: {
       '/api': {
         target: 'https://vectrik-bank.onrender.com',
         changeOrigin: true,
       },
       '/internal/media': {
         target: 'https://vectrik-bank.onrender.com',
         changeOrigin: true,
       },
     },
   }
   
   // Стало:
   // Убрано - теперь это делает Nginx
   ```

2. **Обновлен baseURL** в `vektrik-frontend/src/main.js`:
   ```javascript
   // Было:
   axios.defaults.baseURL = process.env.VUE_APP_API_URL || '/';
   
   // Стало:
   axios.defaults.baseURL = process.env.NODE_ENV === 'production' ? '' : '/';
   ```

3. **Обновлен nginx.conf** в `vektrik-frontend/nginx.conf`:
   ```nginx
   # Было:
   listen 80;
   proxy_pass https://vectrik-bank.onrender.com/api/;
   
   # Стало:
   listen 3000;
   proxy_pass http://backend:8080/api/;
   ```

## 🐳 Docker изменения:

1. **Обновлен docker-compose.yml**:
   - Добавлен сервис `nginx` для реверс прокси
   - Backend и frontend теперь используют `expose` вместо `ports`
   - Обновлены пути для контекстов сборки

2. **Создан backend/Dockerfile**:
   - Перемещен из корня в папку backend
   - Обновлены комментарии

3. **Обновлен vektrik-frontend/Dockerfile**:
   - Изменен порт с 80 на 3000
   - Обновлены комментарии

## 📁 Новые файлы:

1. **nginx/conf.d/vectrikbank.ru.conf** - конфигурация Nginx для домена
2. **deploy.sh** - скрипт автоматического деплоя
3. **ssl-setup.sh** - скрипт настройки SSL сертификатов
4. **DEPLOYMENT.md** - подробная инструкция по деплою
5. **backend/.dockerignore** - исключения для Docker сборки backend

## 🏗️ Архитектура:

```
Internet → Nginx (80/443) → Frontend (3000) + Backend (8080)
```

### Маршрутизация:
- `vectrikbank.ru/` → Frontend (Vue.js)
- `vectrikbank.ru/api/*` → Backend (Go API)
- `vectrikbank.ru/internal/media/*` → Backend (медиа файлы)
- `api.vectrikbank.ru/*` → Backend (только API)

## 🚀 Готово к деплою:

1. ✅ DNS настройки изменены
2. ✅ Код адаптирован для реверс прокси
3. ✅ Docker конфигурация обновлена
4. ✅ Скрипты деплоя созданы
5. ✅ SSL настройка подготовлена

## 📋 Следующие шаги:

1. **Запустить деплой:**
   ```bash
   chmod +x deploy.sh ssl-setup.sh
   ./deploy.sh
   ```

2. **Настроить SSL:**
   ```bash
   ./ssl-setup.sh
   ```

3. **Проверить работу:**
   - https://vectrikbank.ru
   - https://api.vectrikbank.ru/api/me

## 🔍 Проверка изменений:

Все изменения внесены внимательно и корректно. Код готов для деплоя на сервер с реверс прокси через Nginx. 