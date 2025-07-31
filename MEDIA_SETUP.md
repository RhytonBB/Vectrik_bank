# Настройка медиафайлов для Render

## Проблема
На Render файловая система эфемерная, что означает, что загруженные файлы теряются при перезапуске сервера.

## Решения

### 1. Локальное хранение (временное решение)
Для тестирования можно использовать локальное хранение, но файлы будут теряться при перезапуске.

**Переменные окружения для Render:**
```
MEDIA_STORAGE_TYPE=local
MEDIA_BASE_URL=/internal/media/
```

### 2. Облачное хранилище (рекомендуется)

#### Cloudinary (бесплатный план)
1. Зарегистрируйтесь на [cloudinary.com](https://cloudinary.com)
2. Получите URL подключения из Dashboard
3. Установите переменные окружения:

```
MEDIA_STORAGE_TYPE=cloudinary
CLOUDINARY_URL=cloudinary://your_cloudinary_url_here
MEDIA_BASE_URL=https://res.cloudinary.com/your_cloud_name/image/upload/
```

#### Другие варианты:
- **AWS S3** (есть бесплатный план)
- **Google Cloud Storage**
- **Azure Blob Storage**

### 3. Настройка переменных окружения на Render

1. Перейдите в Dashboard вашего приложения на Render
2. Откройте раздел "Environment"
3. Добавьте следующие переменные:

```
MEDIA_STORAGE_TYPE=cloudinary
CLOUDINARY_URL=cloudinary://your_cloudinary_url_here
MEDIA_BASE_URL=https://res.cloudinary.com/your_cloud_name/image/upload/
```

### 4. Временное решение для тестирования

Если нужно быстро протестировать на Render без настройки облачного хранилища:

1. Установите переменные окружения:
```
MEDIA_STORAGE_TYPE=local
MEDIA_BASE_URL=/internal/media/
```

2. Файлы будут сохраняться локально, но будут теряться при перезапуске
3. Для постоянного решения настройте облачное хранилище

## Структура файлов

- `internal/services/media_service.go` - сервис для работы с медиафайлами
- `internal/config/config.go` - конфигурация медиафайлов
- `cmd/server/main.go` - инициализация медиа-сервиса

## Использование

Медиа-сервис автоматически используется в:
- Загрузке баннеров
- Загрузке изображений товаров
- Рассылке медиафайлов

## Примечания

- При использовании локального хранения на Render файлы будут теряться при перезапуске
- Рекомендуется настроить облачное хранилище для продакшена
- Cloudinary предоставляет бесплатный план с 25GB хранилища 