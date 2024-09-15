-- migrate:up
-- all passwords are "1234"
INSERT INTO appuser (id, username, hpassword, surname)
VALUES
    (1, 'test_1', '$2a$14$on9lMaLI1vHrnHhU.bf8aOOe9Bq1cIjNWTcjUJbbXTHjeDQt5ZU7K', NULL),
    (2, 'test_2', '$2a$14$on9lMaLI1vHrnHhU.bf8aOOe9Bq1cIjNWTcjUJbbXTHjeDQt5ZU7K', NULL),
    (3, 'test_3', '$2a$14$on9lMaLI1vHrnHhU.bf8aOOe9Bq1cIjNWTcjUJbbXTHjeDQt5ZU7K', NULL),
    (4, 'test_4', '$2a$14$on9lMaLI1vHrnHhU.bf8aOOe9Bq1cIjNWTcjUJbbXTHjeDQt5ZU7K', NULL),
    (5, 'test_5', '$2a$14$on9lMaLI1vHrnHhU.bf8aOOe9Bq1cIjNWTcjUJbbXTHjeDQt5ZU7K', NULL),
    (6, 'pak', '$2a$14$on9lMaLI1vHrnHhU.bf8aOOe9Bq1cIjNWTcjUJbbXTHjeDQt5ZU7K', 'Pak'),
    (7, 'smith', '$2a$14$on9lMaLI1vHrnHhU.bf8aOOe9Bq1cIjNWTcjUJbbXTHjeDQt5ZU7K', 'Smith'),
    (8, 'bow', '$2a$14$on9lMaLI1vHrnHhU.bf8aOOe9Bq1cIjNWTcjUJbbXTHjeDQt5ZU7K', 'Bow'),
    (9, 'gannick', '$2a$14$on9lMaLI1vHrnHhU.bf8aOOe9Bq1cIjNWTcjUJbbXTHjeDQt5ZU7K', 'Gannick'),
    (10, 'smalltown', '$2a$14$on9lMaLI1vHrnHhU.bf8aOOe9Bq1cIjNWTcjUJbbXTHjeDQt5ZU7K', 'Smalltown');

-- migrate:down
DELETE FROM appuser WHERE username LIKE 'test_%' OR username IN ('pak', 'smith', 'bow', 'gannick', 'smalltown')
