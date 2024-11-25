-- Covered objects:
-- - Table
-- - Index

CREATE DATABASE devsecdb_test_todo;

-- Table and Index
CREATE TABLE devsecdb_test_todo.author (
	id INTEGER PRIMARY KEY AUTO_INCREMENT,
	name VARCHAR(255) COMMENT 'name of the author'
);

CREATE TABLE devsecdb_test_todo.todo (
	id INTEGER PRIMARY KEY AUTO_INCREMENT,
	name VARCHAR(255) COMMENT 'name of the todo',
	content TEXT,
	author_id INTEGER,
	created_ts BIGINT NOT NULL,
	updated_ts BIGINT NOT NULL,
	FOREIGN KEY (author_id) REFERENCES author (id) ON DELETE CASCADE,
	INDEX (name) COMMENT 'index on todo.name',
	INDEX (created_ts) COMMENT 'index on todo.created_ts'
);
