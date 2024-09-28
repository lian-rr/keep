package sqlite

// table creation
const (
	CommandTableQuery = `
	CREATE TABLE IF NOT EXISTS commands (
		uuid VARCHAR(16) PRIMARY KEY,
		name VARCHAR(64) NOT NULL,
		description VARCHAR(255),
		command VARCHAR(255) NOT NULL
	)`

	ParametersTableQuery = `
	CREATE TABLE IF NOT EXISTS parameters (
		uuid VARCHAR(16) PRIMARY KEY,
		command VARCHAR(16),
		name VARCHAR(64) NOT NULL,
		description VARCHAR(255),
		value VARCHAR(32)
	)`

	// TODO: thing more regarding this part
	TagsTableQuery = `
	CREATE TABLE IF NOT EXISTS tags (
		tag VARCHAR(16) PRIMARY KEY
	)`

	TagsAndCommandsTableQuery = `
	CREATE TABLE IF NOT EXISTS tags_commands (
		tag VARCHAR(16),
		command VARCHAR(16),
		PRIMARY KEY (tag, command)
	)`
)

// queries
const (
	InsertCommandQuery = `
	INSERT INTO 
		commands(uuid, name, description, command) 
	VALUES (?, ?, ?, ?)`

	InsertParameterPartialQuery = `
	INSERT INTO 
		parameters(uuid, command, name, description, value)
	VALUES %s`

	GetAllCommandsQuery = `
	SELECT uuid, name, description, command 
	FROM commands`

	GetCommandbyIDQuery = `
	SELECT 
		uuid, name, description, command 
	FROM commands
	WHERE uuid = ?`

	GetParametersByCommandID = `
	SELECT 
		uuid, name, description, value 
	FROM commands
	WHERE command = ?`
)
