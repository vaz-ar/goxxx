
CREATE TABLE quote_backup (
    id integer NOT NULL PRIMARY KEY,
    user TEXT,
    content TEXT,
    date DATETIME DEFAULT CURRENT_TIMESTAMP);

INSERT INTO quote_backup SELECT user, content, date FROM Quote;

DROP TABLE Quote;

ALTER TABLE quote_backup RENAME TO Quote;

