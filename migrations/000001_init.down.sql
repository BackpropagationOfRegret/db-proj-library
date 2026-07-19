DROP INDEX IF EXISTS book_genres_genre_id_idx;
DROP INDEX IF EXISTS book_authors_author_id_idx;
DROP INDEX IF EXISTS authors_full_name_idx;
DROP INDEX IF EXISTS books_title_trgm_idx;
DROP INDEX IF EXISTS loans_active_due_idx;
DROP INDEX IF EXISTS loans_reader_status_idx;
DROP INDEX IF EXISTS copies_book_status_idx;
DROP INDEX IF EXISTS reservations_active_reader_book_idx;
DROP INDEX IF EXISTS loans_active_copy_idx;

DROP TABLE IF EXISTS fines;
DROP TABLE IF EXISTS reservations;
DROP TABLE IF EXISTS loans;
DROP TABLE IF EXISTS readers;
DROP TABLE IF EXISTS copies;
DROP TABLE IF EXISTS book_genres;
DROP TABLE IF EXISTS book_authors;
DROP TABLE IF EXISTS books;
DROP TABLE IF EXISTS genres;
DROP TABLE IF EXISTS authors;

DROP TYPE IF EXISTS reservation_status;
DROP TYPE IF EXISTS loan_status;
DROP TYPE IF EXISTS reader_status;
DROP TYPE IF EXISTS copy_status;
