CREATE TABLE IF NOT EXISTS posts (
	title			text,
	content			text,
	poster_id		text,
	id				text PRIMARY KEY,
	comment_ids		text ARRAY,
	date_created	timestamp
);

CREATE TABLE IF NOT EXISTS comments (
	content			text,
	post_id			text,
	poster_id		text,
	id				text PRIMARY KEY,
	date_created	timestamp
);

CREATE TABLE IF NOT EXISTS users (
	name			text UNIQUE,
	id				text,
	password		text PRIMARY KEY,
	date_joined		timestamp
);