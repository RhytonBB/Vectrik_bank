#!/bin/bash

# Скрипт деплоя Vektrik Bank на сервер
set -e

echo "🚀 Начинаем деплой Vektrik Bank..."

# Проверяем подключение к серверу
echo "📡 Проверяем подключение к серверу..."
ssh root@147.45.159.81 "echo 'Подключение успешно'"

# Создаем структуру папок на сервере
echo "📁 Создаем структуру папок..."
ssh root@147.45.159.81 << 'EOF'
mkdir -p /var/www/vectrikbank/{backend,frontend,nginx/conf.d,nginx/ssl}
cd /var/www/vectrikbank
EOF

# Копируем файлы backend (исключаем frontend и другие файлы)
echo "📦 Копируем backend..."
rsync -av --exclude='vektrik-frontend/' --exclude='.git/' --exclude='*.md' --exclude='*.txt' --exclude='*.xlsx' --exclude='*.csv' --exclude='deploy.sh' --exclude='ssl-setup.sh' --exclude='DEPLOYMENT.md' --exclude='docker-compose.yml' --exclude='nginx/' . root@147.45.159.81:/var/www/vectrikbank/backend/

# Копируем файлы frontend
echo "📦 Копируем frontend..."
scp -r vektrik-frontend/* root@147.45.159.81:/var/www/vectrikbank/frontend/

# Копируем конфигурацию nginx
echo "⚙️ Копируем конфигурацию nginx..."
scp nginx/conf.d/vectrikbank.ru.conf root@147.45.159.81:/var/www/vectrikbank/nginx/conf.d/

# Копируем docker-compose
echo "🐳 Копируем docker-compose..."
scp docker-compose.yml root@147.45.159.81:/var/www/vectrikbank/

# Устанавливаем Docker и зависимости на сервере
echo "🔧 Устанавливаем зависимости..."
ssh root@147.45.159.81 << 'EOF'
cd /var/www/vectrikbank

# Обновляем систему
apt update && apt upgrade -y

# Устанавливаем Docker если не установлен
if ! command -v docker &> /dev/null; then
    curl -fsSL https://get.docker.com -o get-docker.sh
    sh get-docker.sh
    systemctl enable docker
    systemctl start docker
fi

# Устанавливаем docker-compose если не установлен
if ! command -v docker-compose &> /dev/null; then
    curl -L "https://github.com/docker/compose/releases/download/v2.20.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
    chmod +x /usr/local/bin/docker-compose
fi

# Устанавливаем certbot для SSL
apt install -y certbot python3-certbot-nginx

# Останавливаем стандартный nginx если запущен
systemctl stop nginx || true
systemctl disable nginx || true
EOF

# Запускаем приложение
echo "🚀 Запускаем приложение..."
ssh root@147.45.159.81 << 'EOF'
cd /var/www/vectrikbank

# Останавливаем существующие контейнеры
docker-compose down || true

# Собираем и запускаем
docker-compose up -d --build

# Проверяем статус
docker-compose ps
EOF

echo "✅ Деплой завершен!"
echo "🌐 Сайт будет доступен по адресу: https://vectrikbank.ru"
echo "🔧 Для получения SSL сертификата выполните:"
echo "   ssh root@147.45.159.81"
echo "   cd /var/www/vectrikbank"
echo "   certbot --nginx -d vectrikbank.ru -d www.vectrikbank.ru -d api.vectrikbank.ru" 