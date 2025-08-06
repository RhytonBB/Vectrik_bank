# 🚀 Деплой Vektrik Bank на сервер

## 📋 Предварительные требования

1. **DNS настройки** (уже выполнены):
   - `vectrikbank.ru` → `147.45.159.81`
   - `www.vectrikbank.ru` → `147.45.159.81`
   - `api.vectrikbank.ru` → `147.45.159.81`

2. **Доступ к серверу**:
   - IP: `147.45.159.81`
   - Пользователь: `root`
   - Пароль: `yg-aXS1?.HC@yx`

## 🏗️ Архитектура

```
Internet → Nginx (80/443) → Frontend (3000) + Backend (8080)
```

- **Nginx** - реверс прокси, SSL терминация
- **Frontend** - Vue.js приложение на порту 3000
- **Backend** - Go API на порту 8080

## 📁 Структура проекта

```
/var/www/vectrikbank/
├── backend/          # Go backend
├── frontend/         # Vue.js frontend
├── nginx/
│   └── conf.d/
│       └── vectrikbank.ru.conf
├── docker-compose.yml
├── deploy.sh
└── ssl-setup.sh
```

## 🚀 Быстрый деплой

### 1. Подготовка локальных файлов

```bash
# Сделать скрипты исполняемыми
chmod +x deploy.sh ssl-setup.sh
```

### 2. Запуск деплоя

```bash
# Основной деплой
./deploy.sh

# Настройка SSL (после деплоя)
./ssl-setup.sh
```

## 🔧 Ручной деплой

### 1. Подключение к серверу

```bash
ssh root@147.45.159.81
```

### 2. Подготовка сервера

```bash
# Обновление системы
apt update && apt upgrade -y

# Установка Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sh get-docker.sh
systemctl enable docker
systemctl start docker

# Установка docker-compose
curl -L "https://github.com/docker/compose/releases/download/v2.20.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
chmod +x /usr/local/bin/docker-compose

# Установка certbot
apt install -y certbot python3-certbot-nginx
```

### 3. Создание структуры папок

```bash
mkdir -p /var/www/vectrikbank/{backend,frontend,nginx/conf.d,nginx/ssl}
cd /var/www/vectrikbank
```

### 4. Копирование файлов

```bash
# На локальной машине
scp -r . root@147.45.159.81:/var/www/vectrikbank/backend/
scp -r vektrik-frontend/* root@147.45.159.81:/var/www/vectrikbank/frontend/
scp docker-compose.yml root@147.45.159.81:/var/www/vectrikbank/
scp nginx/conf.d/vectrikbank.ru.conf root@147.45.159.81:/var/www/vectrikbank/nginx/conf.d/
```

### 5. Запуск приложения

```bash
cd /var/www/vectrikbank
docker-compose up -d --build
```

### 6. Настройка SSL

```bash
# Остановка контейнеров
docker-compose down

# Временная конфигурация nginx для certbot
cat > /etc/nginx/sites-available/vectrikbank.ru << 'EOF'
server {
    listen 80;
    server_name vectrikbank.ru www.vectrikbank.ru api.vectrikbank.ru;
    
    location /.well-known/acme-challenge/ {
        root /var/www/certbot;
    }
    
    location / {
        return 301 https://$host$request_uri;
    }
}
EOF

ln -sf /etc/nginx/sites-available/vectrikbank.ru /etc/nginx/sites-enabled/
mkdir -p /var/www/certbot
systemctl start nginx

# Получение сертификатов
certbot certonly --webroot -w /var/www/certbot \
    -d vectrikbank.ru \
    -d www.vectrikbank.ru \
    -d api.vectrikbank.ru \
    --email admin@vectrikbank.ru \
    --agree-tos \
    --non-interactive

# Запуск с SSL
docker-compose up -d
```

## 🔍 Проверка работы

### Проверка контейнеров

```bash
docker-compose ps
docker-compose logs -f
```

### Проверка доступности

```bash
# HTTP редирект
curl -I http://vectrikbank.ru

# HTTPS доступ
curl -I https://vectrikbank.ru
curl -I https://api.vectrikbank.ru/api/me
```

### Проверка SSL

```bash
# Проверка сертификатов
certbot certificates

# Автоматическое обновление
crontab -e
# Добавить: 0 12 * * * /usr/bin/certbot renew --quiet
```

## 🛠️ Управление

### Перезапуск сервисов

```bash
# Все сервисы
docker-compose restart

# Конкретный сервис
docker-compose restart backend
docker-compose restart frontend
docker-compose restart nginx
```

### Обновление

```bash
# Остановка
docker-compose down

# Обновление кода
git pull

# Пересборка и запуск
docker-compose up -d --build
```

### Логи

```bash
# Все логи
docker-compose logs -f

# Конкретный сервис
docker-compose logs -f backend
docker-compose logs -f frontend
docker-compose logs -f nginx
```

## 🔧 Изменения в коде

### Backend изменения:

1. **Cookie domain** в `internal/api/auth_handlers.go`:
   ```go
   Domain: "vectrikbank.ru", // убрана точка в начале
   ```

### Frontend изменения:

1. **Убран прокси** в `vue.config.js` - теперь это делает Nginx
2. **Обновлен baseURL** в `main.js` для работы с реверс прокси
3. **Обновлен nginx.conf** для работы на порту 3000

## 🚨 Устранение неполадок

### Проблема: Контейнеры не запускаются

```bash
# Проверка логов
docker-compose logs

# Проверка портов
netstat -tlnp | grep :80
netstat -tlnp | grep :443
```

### Проблема: SSL не работает

```bash
# Проверка сертификатов
ls -la /etc/letsencrypt/live/vectrikbank.ru/

# Пересоздание сертификатов
certbot delete --cert-name vectrikbank.ru
certbot certonly --webroot -w /var/www/certbot -d vectrikbank.ru -d www.vectrikbank.ru -d api.vectrikbank.ru
```

### Проблема: DNS не обновился

```bash
# Проверка DNS
nslookup vectrikbank.ru
dig vectrikbank.ru

# Ожидание обновления (может занять до 30 минут)
```

## 📞 Поддержка

При возникновении проблем:

1. Проверьте логи: `docker-compose logs -f`
2. Проверьте статус контейнеров: `docker-compose ps`
3. Проверьте доступность портов: `netstat -tlnp`
4. Проверьте DNS: `nslookup vectrikbank.ru` 