#!/bin/bash

# Скрипт настройки SSL сертификатов для Vektrik Bank
set -e

echo "🔒 Настраиваем SSL сертификаты..."

# Подключаемся к серверу и настраиваем SSL
ssh root@147.45.159.81 << 'EOF'
cd /var/www/vectrikbank

echo "📋 Проверяем доступность доменов..."
nslookup vectrikbank.ru
nslookup www.vectrikbank.ru
nslookup api.vectrikbank.ru

echo "🔧 Останавливаем контейнеры для настройки SSL..."
docker-compose down

echo "📝 Создаем временную конфигурацию nginx для certbot..."
cat > /etc/nginx/sites-available/vectrikbank.ru << 'NGINX'
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
NGINX

# Создаем символическую ссылку
ln -sf /etc/nginx/sites-available/vectrikbank.ru /etc/nginx/sites-enabled/
rm -f /etc/nginx/sites-enabled/default

# Создаем папку для certbot
mkdir -p /var/www/certbot

# Запускаем nginx
systemctl start nginx
systemctl enable nginx

echo "🔐 Получаем SSL сертификаты..."
certbot certonly --webroot -w /var/www/certbot \
    -d vectrikbank.ru \
    -d www.vectrikbank.ru \
    -d api.vectrikbank.ru \
    --email admin@vectrikbank.ru \
    --agree-tos \
    --non-interactive

echo "🔄 Перезапускаем приложение с SSL..."
docker-compose up -d

echo "✅ SSL сертификаты настроены!"
echo "🌐 Сайт доступен по адресу: https://vectrikbank.ru"
EOF

echo "✅ Настройка SSL завершена!" 