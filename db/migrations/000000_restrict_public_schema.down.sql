-- Откат ограничений PUBLIC до стандартных настроек “как в чистой БД”.

-- 1) Вернём PUBLIC доступ к БД на уровень по умолчанию.
-- По умолчанию PUBLIC имеет CONNECT; TEMPORARY также часто разрешён.
GRANT CONNECT, TEMPORARY ON DATABASE evaluation_db TO PUBLIC;

-- 2) Вернём PUBLIC дефолтные права на схему public:
-- в типичной свежей БД PUBLIC имеет USAGE и CREATE.
GRANT USAGE, CREATE ON SCHEMA public TO PUBLIC;

-- 3) Сбросим принудительный search_path у пользователя сервиса.
ALTER ROLE evaluation_user RESET search_path;
