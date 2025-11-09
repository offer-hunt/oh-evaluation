CREATE TABLE IF NOT EXISTS evaluation.__migration_probe
(
    id         BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
    );

-- Устанавливаем права на новую таблицу
GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE evaluation.__migration_probe TO evaluation_user;
GRANT USAGE, SELECT ON SEQUENCE evaluation.__migration_probe_id_seq TO evaluation_user;