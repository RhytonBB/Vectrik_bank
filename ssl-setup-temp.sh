#!/bin/bash

# Временный скрипт для настройки SSL сертификатов
set -e

echo "🔒 Настраиваем SSL сертификаты..."

# 1. Остановить контейнеры
echo "🛑 Останавливаем контейнеры..."
docker-compose down

# 2. Установить certbot
echo "📦 Устанавливаем certbot..."
apt install -y certbot python3-certbot-nginx

# 3. Создать временную конфигурацию для certbot
echo "📝 Создаем временную конфигурацию nginx..."
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

# 4. Активировать конфигурацию
echo "🔗 Активируем конфигурацию..."
ln -sf /etc/nginx/sites-available/vectrikbank.ru /etc/nginx/sites-enabled/
rm -f /etc/nginx/sites-enabled/default

# 5. Создать папку для certbot
echo "📁 Создаем папку для certbot..."
mkdir -p /var/www/certbot

# 6. Запустить nginx
echo "🚀 Запускаем nginx..."
systemctl start nginx
systemctl enable nginx

# 7. Получить SSL сертификаты
echo "🔐 Получаем SSL сертификаты..."
certbot certonly --webroot -w /var/www/certbot \
    -d vectrikbank.ru \
    -d www.vectrikbank.ru \
    -d api.vectrikbank.ru \
    --email admin@vectrikbank.ru \
    --agree-tos \
    --non-interactive

echo "✅ SSL сертификаты получены!"
echo "🔄 Теперь запустите: docker-compose up -d" 