-- Миграция для добавления поля account_number в таблицу users
-- Добавляем новое поле account_number
ALTER TABLE users ADD COLUMN account_number VARCHAR(8) UNIQUE NOT NULL DEFAULT '00000000';

-- Обновляем существующие записи, генерируя уникальные номера счетов
-- Это нужно сделать для существующих пользователей
-- В Go коде уже есть логика для генерации уникальных номеров при создании новых пользователей

-- Создаем индекс для быстрого поиска по номеру счета
CREATE INDEX idx_users_account_number ON users(account_number); 