-- Цель: запретить всему PUBLIC что-либо создавать в public-схеме
-- и ограничить доступ к БД только нужному пользователю сервиса.
-- Предполагается, что БД называется evaluation_db и пользователь evaluation_user
-- (как в docker-compose).

-- 1) Ограничим доступ к самой БД.
REVOKE ALL ON DATABASE evaluation_db FROM PUBLIC;

-- Разрешим подключаться и создавать временные таблицы только нашему пользователю.
GRANT CONNECT, TEMPORARY ON DATABASE evaluation_db TO evaluation_user;

-- 2) Запретим создавать объекты в схеме public всем.
REVOKE CREATE ON SCHEMA public FROM PUBLIC;

-- По умолчанию в новой БД на схему public у PUBLIC есть USAGE и CREATE.
-- Чтобы быть строже, снимем любые права:
REVOKE ALL ON SCHEMA public FROM PUBLIC;

-- 3) (Опционально) Настроим search_path пользователю сервиса на свою схему.
-- Сама схема evaluation создаётся следующей миграцией (000001), поэтому здесь
-- просто фиксируем будущий search_path, чтобы по подключению он уже был корректным.
ALTER ROLE evaluation_user SET search_path = evaluation;
