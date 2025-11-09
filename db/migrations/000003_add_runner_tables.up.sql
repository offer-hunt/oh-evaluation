BEGIN;

-- заранее создаём нужные схемы
CREATE SCHEMA IF NOT EXISTS course;
CREATE SCHEMA IF NOT EXISTS evaluation;

-- заглушка внешней таблицы, на которую мы потом ссылаемся
-- если она уже есть у "курса" — этот блок просто ничего не сделает
CREATE TABLE IF NOT EXISTS course.course_question_test_cases (
                                                                 id              UUID PRIMARY KEY,
                                                                 question_id     UUID NOT NULL,
                                                                 input_data      TEXT NOT NULL,
                                                                 expected_output TEXT NOT NULL,
                                                                 timeout_ms      INT NULL,
                                                                 memory_limit_mb INT NULL
);

-- 1. Таблица с отправками
CREATE TABLE IF NOT EXISTS evaluation.runner_submissions (
                                                             id              UUID PRIMARY KEY,
                                                             user_id         UUID NOT NULL,
                                                             course_id       UUID NULL,
                                                             lesson_id       UUID NULL,
                                                             page_id         UUID NOT NULL,
                                                             question_id     UUID NOT NULL,
                                                             submission_type VARCHAR NOT NULL CHECK (submission_type IN ('TEXT', 'CODE')),
    language        VARCHAR(32) NULL,
    runtime_image   VARCHAR(128) NULL,
    time_limit_ms   INT NULL,
    memory_limit_mb INT NULL,
    code            TEXT NOT NULL,
    status          VARCHAR NOT NULL,
    verdict         VARCHAR NOT NULL CHECK (verdict IN ('PENDING', 'ACCEPTED', 'REJECTED')),
    tests_passed    INT NOT NULL DEFAULT 0,
    tests_total     INT NOT NULL DEFAULT 0,
    time_ms         INT NULL,
    memory_kb       INT NULL,
    result          JSONB NULL,
    -- это просто ссылка-ид, без FK, чтобы не делать кольцевую зависимость
    ai_review_id    UUID NULL,
    submitted_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    started_at      TIMESTAMPTZ NULL,
    finished_at     TIMESTAMPTZ NULL
    );

-- Индексы
CREATE INDEX IF NOT EXISTS runner_submissions_user_id_idx ON evaluation.runner_submissions (user_id);
CREATE INDEX IF NOT EXISTS runner_submissions_question_id_idx ON evaluation.runner_submissions (question_id);
CREATE INDEX IF NOT EXISTS runner_submissions_page_id_idx ON evaluation.runner_submissions (page_id);
CREATE INDEX IF NOT EXISTS runner_submissions_verdict_idx ON evaluation.runner_submissions (verdict);
CREATE INDEX IF NOT EXISTS runner_submissions_submitted_at_idx ON evaluation.runner_submissions (submitted_at);

-- Права сервисному пользователю
GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE evaluation.runner_submissions TO evaluation_user;


-- 2. Таблица с AI-review результатами
CREATE TABLE IF NOT EXISTS evaluation.runner_ai_reviews (
                                                            id            UUID PRIMARY KEY,
                                                            submission_id UUID NOT NULL REFERENCES evaluation.runner_submissions (id) ON DELETE CASCADE,
    status        VARCHAR NOT NULL CHECK (status IN ('PENDING', 'DONE', 'ERROR')),
    model         VARCHAR(64) NULL,
    score         NUMERIC(5,2) NULL,
    summary       TEXT NULL,
    suggestions   JSONB NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    completed_at  TIMESTAMPTZ NULL,
    -- один review на submission — логично держать уникальность
    UNIQUE (submission_id)
    );

CREATE INDEX IF NOT EXISTS runner_ai_reviews_submission_id_idx ON evaluation.runner_ai_reviews (submission_id);

GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE evaluation.runner_ai_reviews TO evaluation_user;


-- 3. Таблица с результатами по каждому тест-кейсу
CREATE TABLE IF NOT EXISTS evaluation.runner_test_case_results (
                                                                   id             UUID PRIMARY KEY,
                                                                   submission_id  UUID NOT NULL REFERENCES evaluation.runner_submissions (id) ON DELETE CASCADE,
    -- внешний кейс из другой схемы
    test_case_id   UUID NULL REFERENCES course.course_question_test_cases (id),
    status         VARCHAR NOT NULL CHECK (status IN ('PASS', 'FAIL', 'TLE', 'MLE', 'RE')),
    time_ms        INT NULL,
    memory_kb      INT NULL,
    stderr_snippet TEXT NULL,
    diff_snippet   TEXT NULL
    );

CREATE INDEX IF NOT EXISTS runner_test_case_results_submission_id_idx ON evaluation.runner_test_case_results (submission_id);
CREATE INDEX IF NOT EXISTS runner_test_case_results_test_case_id_idx ON evaluation.runner_test_case_results (test_case_id);

GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE evaluation.runner_test_case_results TO evaluation_user;

COMMIT;
