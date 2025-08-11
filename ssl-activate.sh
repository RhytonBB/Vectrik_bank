#!/bin/bash

# Скрипт активации SSL после получения сертификатов
set -e

echo "🔒 Активируем SSL конфигурацию..."

# 1. Проверить что сертификаты есть
if [ ! -f "/etc/letsencrypt/live/vectrikbank.ru/fullchain.pem" ]; then
    echo "❌ SSL сертификаты не найдены!"
    echo "Сначала запустите: ./ssl-setup-temp.sh"
    exit 1
fi

echo "✅ SSL сертификаты найдены!"

# 2. Остановить контейнеры
echo "🛑 Останавливаем контейнеры..."
docker-compose down

# 3. Остановить системный nginx
echo "🛑 Останавливаем системный nginx..."
systemctl stop nginx

# 4. Скопировать SSL конфигурацию
echo "📝 Копируем SSL конфигурацию..."
cp nginx/conf.d/vectrikbank.ru.ssl.conf nginx/conf.d/vectrikbank.ru.conf

# 5. Запустить с SSL
echo "🚀 Запускаем с SSL..."
docker-compose up -d

# 6. Проверить статус
echo "🔍 Проверяем статус..."
docker-compose ps

# 7. Проверить доступность
echo "🌐 Проверяем доступность..."
sleep 5
curl -I http://vectrikbank.ru
curl -I https://vectrikbank.ru

echo "✅ SSL активирован!"
echo "🌐 Сайт доступен по адресу: https://vectrikbank.ru" 