CREATE TYPE copy_status AS ENUM ('available', 'on_loan', 'reserved', 'lost', 'maintenance');
CREATE TYPE reader_status AS ENUM ('active', 'blocked', 'inactive');
CREATE TYPE loan_status AS ENUM ('active', 'returned', 'overdue');
CREATE TYPE reservation_status AS ENUM ('pending', 'fulfilled', 'cancelled', 'expired');

CREATE TABLE authors (
    id          BIGSERIAL PRIMARY KEY,
    full_name   TEXT NOT NULL,
    birth_date  DATE,
    bio         TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE genres (
    id          BIGSERIAL PRIMARY KEY,
    name        TEXT NOT NULL UNIQUE
);

CREATE TABLE books (
    id               BIGSERIAL PRIMARY KEY,
    title            TEXT NOT NULL,
    isbn             TEXT NOT NULL UNIQUE,
    publication_year INT NOT NULL CHECK (publication_year BETWEEN 1000 AND 2100),
    pages            INT NOT NULL CHECK (pages > 0),
    description      TEXT NOT NULL DEFAULT '',
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE book_authors (
    book_id   BIGINT NOT NULL REFERENCES books (id) ON DELETE CASCADE,
    author_id BIGINT NOT NULL REFERENCES authors (id) ON DELETE CASCADE,
    PRIMARY KEY (book_id, author_id)
);

CREATE TABLE book_genres (
    book_id  BIGINT NOT NULL REFERENCES books (id) ON DELETE CASCADE,
    genre_id BIGINT NOT NULL REFERENCES genres (id) ON DELETE CASCADE,
    PRIMARY KEY (book_id, genre_id)
);

CREATE TABLE copies (
    id               BIGSERIAL PRIMARY KEY,
    book_id          BIGINT NOT NULL REFERENCES books (id) ON DELETE CASCADE,
    inventory_number TEXT NOT NULL UNIQUE,
    status           copy_status NOT NULL DEFAULT 'available',
    condition        TEXT NOT NULL DEFAULT 'good',
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE readers (
    id            BIGSERIAL PRIMARY KEY,
    first_name    TEXT NOT NULL,
    last_name     TEXT NOT NULL,
    email         TEXT NOT NULL UNIQUE,
    phone         TEXT NOT NULL DEFAULT '',
    status        reader_status NOT NULL DEFAULT 'active',
    registered_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE loans (
    id          BIGSERIAL PRIMARY KEY,
    copy_id     BIGINT NOT NULL REFERENCES copies (id),
    reader_id   BIGINT NOT NULL REFERENCES readers (id),
    loaned_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    due_at      TIMESTAMPTZ NOT NULL,
    returned_at TIMESTAMPTZ,
    status      loan_status NOT NULL DEFAULT 'active',
    CONSTRAINT loans_dates_check CHECK (
        due_at > loaned_at
        AND (returned_at IS NULL OR returned_at >= loaned_at)
    )
);

CREATE TABLE reservations (
    id          BIGSERIAL PRIMARY KEY,
    book_id     BIGINT NOT NULL REFERENCES books (id) ON DELETE CASCADE,
    reader_id   BIGINT NOT NULL REFERENCES readers (id) ON DELETE CASCADE,
    reserved_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at  TIMESTAMPTZ NOT NULL,
    status      reservation_status NOT NULL DEFAULT 'pending',
    CONSTRAINT reservations_dates_check CHECK (expires_at > reserved_at)
);

CREATE TABLE fines (
    id         BIGSERIAL PRIMARY KEY,
    loan_id    BIGINT NOT NULL UNIQUE REFERENCES loans (id) ON DELETE CASCADE,
    amount     NUMERIC(12, 2) NOT NULL CHECK (amount >= 0),
    paid       BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX loans_active_copy_idx
    ON loans (copy_id) WHERE status = 'active';

CREATE UNIQUE INDEX reservations_active_reader_book_idx
    ON reservations (reader_id, book_id) WHERE status = 'pending';

CREATE INDEX copies_book_status_idx ON copies (book_id, status);
CREATE INDEX loans_reader_status_idx ON loans (reader_id, status);
CREATE INDEX loans_active_due_idx ON loans (due_at) WHERE status = 'active';
CREATE INDEX books_title_trgm_idx ON books USING gin (to_tsvector('simple', title));
CREATE INDEX authors_full_name_idx ON authors (full_name);
CREATE INDEX book_authors_author_id_idx ON book_authors (author_id);
CREATE INDEX book_genres_genre_id_idx ON book_genres (genre_id);
