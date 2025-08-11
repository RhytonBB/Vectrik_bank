# 🔒 Настройка SSL сертификатов для Vektrik Bank

## 📋 Пошаговая инструкция

### 🎯 Шаг 1: Скопировать файлы на сервер

```bash
# На локальном компьютере
scp ssl-setup-temp.sh root@147.45.159.81:/root/vectrik-bank/
scp ssl-activate.sh root@147.45.159.81:/root/vectrik-bank/
scp nginx/conf.d/vectrikbank.ru.ssl.conf root@147.45.159.81:/root/vectrik-bank/nginx/conf.d/
```

### 🚀 Шаг 2: Сделать скрипты исполняемыми

```bash
# На сервере
cd /root/vectrik-bank
chmod +x ssl-setup-temp.sh ssl-activate.sh
```

### 🔐 Шаг 3: Получить SSL сертификаты

```bash
# Запустить скрипт получения сертификатов
./ssl-setup-temp.sh
```

### ✅ Шаг 4: Активировать SSL

```bash
# Активировать SSL конфигурацию
./ssl-activate.sh
```

## 🔍 Проверка работы

### Проверить сертификаты:

```bash
# Проверить что сертификаты получены
ls -la /etc/letsencrypt/live/vectrikbank.ru/

# Проверить статус сертификатов
certbot certificates
```

### Проверить доступность:

```bash
# HTTP редирект
curl -I http://vectrikbank.ru

# HTTPS доступ
curl -I https://vectrikbank.ru

# API
curl -I https://api.vectrikbank.ru/api/me
```

### Проверить контейнеры:

```bash
# Статус контейнеров
docker-compose ps

# Логи nginx
docker-compose logs nginx
```

## 🚨 Устранение неполадок

### Проблема: Сертификаты не получены

```bash
# Проверить DNS
nslookup vectrikbank.ru

# Проверить что порт 80 открыт
netstat -tlnp | grep :80

# Проверить файрвол
ufw status
```

### Проблема: Nginx не запускается

```bash
# Проверить синтаксис конфигурации
docker run --rm -v $(pwd)/nginx/conf.d:/etc/nginx/conf.d nginx:alpine nginx -t

# Проверить логи
docker-compose logs nginx
```

### Проблема: SSL ошибки

```bash
# Пересоздать сертификаты
certbot delete --cert-name vectrikbank.ru
./ssl-setup-temp.sh
./ssl-activate.sh
```

## 🔄 Автоматическое обновление

### Настроить автообновление сертификатов:

```bash
# Добавить в crontab
crontab -e

# Добавить строку:
0 12 * * * /usr/bin/certbot renew --quiet && docker-compose restart nginx
```

## 📞 Поддержка

При возникновении проблем:

1. Проверьте логи: `docker-compose logs -f`
2. Проверьте DNS: `nslookup vectrikbank.ru`
3. Проверьте файрвол: `ufw status`
4. Проверьте порты: `netstat -tlnp`

## 🎯 Ожидаемый результат

После успешной настройки:

- ✅ `http://vectrikbank.ru` → редирект на HTTPS
- ✅ `https://vectrikbank.ru` → основной сайт
- ✅ `https://api.vectrikbank.ru` → API
- ✅ Зеленый замок в браузере
- ✅ Безопасное соединение 