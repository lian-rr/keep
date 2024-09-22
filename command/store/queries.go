package store

// table creation
const (
	commandTableQuery = `
	CREATE TABLE IF NOT EXISTS commands (
		uuid VARCHAR(16) PRIMARY KEY,
		name VARCHAR(64) NOT NULL,
		description VARCHAR(255),
		command VARCHAR(255) NOT NULL
	)`

	parametersTableQuery = `
	CREATE TABLE IF NOT EXISTS parameters (
		uuid VARCHAR(16) PRIMARY KEY,
		command VARCHAR(16),
		name VARCHAR(64) NOT NULL,
		description VARCHAR(255),
		value VARCHAR(32)
	)`

	// TODO: thing more regarding this part
	tagsTableQuery = `
	CREATE TABLE IF NOT EXISTS tags (
		tag VARCHAR(16) PRIMARY KEY
	)`

	tagsAndCommandsTableQuery = `
	CREATE TABLE IF NOT EXISTS tags_commands (
		tag VARCHAR(16),
		command VARCHAR(16),
		PRIMARY KEY (tag, command)
	)`
)

// insert queries
const (
	insertCommandQuery = `
	INSERT INTO 
		commands(uuid, name, description, command) 
	VALUES (?, ?, ?, ?)`

	insertParameterPartialQuery = `
	INSERT INTO 
		parameters(uuid, command, name, description, value)
	VALUES %s`
)
