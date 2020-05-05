CREATE TABLE IF NOT EXISTS books (
	id varchar(36) NOT NULL,
	title varchar(200) NOT NULL,
	isbn varchar(18) NOT NULL,
	created_at timestamp NOT NULL,
	PRIMARY KEY(id)
);

