CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS users
(
    id         UUID         NOT NULL PRIMARY KEY,
    login      varchar(128) NOT NULL,
    password   text         NOT NULL,

    created_at TIMESTAMP    NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS users_login_idx ON users (login);

DO
$$
    BEGIN
        IF NOT EXISTS(SELECT 1 FROM pg_type WHERE typname = 'order_statuses') THEN
            CREATE TYPE order_statuses AS ENUM ('NEW','PROCESSING','INVALID','PROCESSED');
        END IF;
    END
$$;

CREATE TABLE IF NOT EXISTS orders
(
    number      varchar        NOT NULL,
    user_id     uuid,
    status      order_statuses NOT NULL DEFAULT 'NEW',
    accrual     decimal,
    uploaded_at TIMESTAMP      NOT NULL DEFAULT now(),
    FOREIGN KEY (user_id) REFERENCES users (id)
        ON DELETE CASCADE
        ON UPDATE CASCADE
        DEFERRABLE INITIALLY DEFERRED
);

CREATE TABLE IF NOT EXISTS withdrawals
(
    user_id      uuid,
    number       varchar   NOT NULL,
    sum          decimal,
    processed_at TIMESTAMP NOT NULL DEFAULT now(),
    FOREIGN KEY (user_id) REFERENCES users (id)
        ON DELETE CASCADE
        ON UPDATE CASCADE
        DEFERRABLE INITIALLY DEFERRED
);