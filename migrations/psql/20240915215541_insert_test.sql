-- migrate:up
-- all passwords are "1234"
INSERT INTO appuser (username, hpassword, surname)
VALUES
    ('test_1', '$2a$14$on9lMaLI1vHrnHhU.bf8aOOe9Bq1cIjNWTcjUJbbXTHjeDQt5ZU7K', NULL),
    ('test_2', '$2a$14$on9lMaLI1vHrnHhU.bf8aOOe9Bq1cIjNWTcjUJbbXTHjeDQt5ZU7K', NULL),
    ('test_3', '$2a$14$on9lMaLI1vHrnHhU.bf8aOOe9Bq1cIjNWTcjUJbbXTHjeDQt5ZU7K', NULL),
    ('test_4', '$2a$14$on9lMaLI1vHrnHhU.bf8aOOe9Bq1cIjNWTcjUJbbXTHjeDQt5ZU7K', NULL),
    ('test_5', '$2a$14$on9lMaLI1vHrnHhU.bf8aOOe9Bq1cIjNWTcjUJbbXTHjeDQt5ZU7K', NULL),
    ('pak', '$2a$14$on9lMaLI1vHrnHhU.bf8aOOe9Bq1cIjNWTcjUJbbXTHjeDQt5ZU7K', 'Pak'),
    ('smith', '$2a$14$on9lMaLI1vHrnHhU.bf8aOOe9Bq1cIjNWTcjUJbbXTHjeDQt5ZU7K', 'Smith'),
    ('bow', '$2a$14$on9lMaLI1vHrnHhU.bf8aOOe9Bq1cIjNWTcjUJbbXTHjeDQt5ZU7K', 'Bow'),
    ('gannick', '$2a$14$on9lMaLI1vHrnHhU.bf8aOOe9Bq1cIjNWTcjUJbbXTHjeDQt5ZU7K', 'Gannick'),
    ('smalltown', '$2a$14$on9lMaLI1vHrnHhU.bf8aOOe9Bq1cIjNWTcjUJbbXTHjeDQt5ZU7K', 'Smalltown');

-- migrate:down

