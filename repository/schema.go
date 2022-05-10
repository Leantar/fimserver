package repository

var schemas = [...]string{
	`CREATE TABLE endpoints (
		id BIGSERIAL PRIMARY KEY,
		name TEXT UNIQUE NOT NULL,
		kind VARCHAR(32) NOT NULL,
		roles TEXT[] NOT NULL,
		has_baseline BOOLEAN NOT NULL,
		baseline_is_current BOOLEAN NOT NULL,
		watched_paths TEXT[] NOT NULL
	);`,
	`CREATE TABLE baseline_fs_objects (
		id BIGSERIAL PRIMARY KEY,
		path TEXT NOT NULL,
		hash VARCHAR(64) NOT NULL,
		created BIGINT NOT NULL,
		modified BIGINT NOT NULL,
		uid INT NOT NULL,
		gid INT NOT NULL,
		mode BIGINT NOT NULL,
		fk_agent_id BIGINT NOT NULL,
		UNIQUE (path, fk_agent_id),
		FOREIGN KEY (fk_agent_id)
			REFERENCES endpoints(id)
			ON DELETE CASCADE);`,
	`CREATE TABLE alerts (
		id BIGSERIAL PRIMARY KEY,
		kind VARCHAR(32) NOT NULL,
		difference TEXT NOT NULL,
		issued_at BIGINT NOT NULL,
		path TEXT NOT NULL,
		modified BIGINT NOT NULL,
		fk_agent_id BIGINT NOT NULL,
		FOREIGN KEY (fk_agent_id)
			REFERENCES endpoints(id)
			ON DELETE CASCADE);`,
	`CREATE TABLE rules (
		id BIGSERIAL PRIMARY KEY,
		p_type VARCHAR(100) NOT NULL,
		v0 VARCHAR(100) NOT NULL DEFAULT '',
		v1 VARCHAR(100) NOT NULL DEFAULT '',
		v2 VARCHAR(100) NOT NULL DEFAULT '',
		v3 VARCHAR(100) NOT NULL DEFAULT '',
		v4 VARCHAR(100) NOT NULL DEFAULT '',
		v5 VARCHAR(100) NOT NULL DEFAULT '',
		UNIQUE (p_type, v0, v1, v2, v3, v4, v5));`,
}
