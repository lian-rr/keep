package sqlite

// table creation
const (
	CommandTableQuery = `
	CREATE TABLE IF NOT EXISTS commands (
		id VARCHAR(16) PRIMARY KEY,
		name VARCHAR(64) NOT NULL,
		description VARCHAR(255),
		command VARCHAR(255) NOT NULL
	)`

	ParametersTableQuery = `
	CREATE TABLE IF NOT EXISTS parameters (
		id VARCHAR(16) PRIMARY KEY,
		command VARCHAR(16),
		name VARCHAR(64) NOT NULL,
		description VARCHAR(255),
		value VARCHAR(32)
	)`

	SearchTableQuery = `
	CREATE VIRTUAL TABLE IF NOT EXISTS commands_fts
	USING fts5(id UNINDEXED, name, command, description);
	`

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

// triggers
const (
	InsertCommandFtsTrigger = `
	CREATE TRIGGER IF NOT EXISTS insert_command_fts_trigger
		AFTER INSERT ON commands
	BEGIN
		INSERT INTO commands_fts (id, name, command, description)
		VALUES (NEW.id, NEW.name, NEW.command, NEW.description);
	END`

	UpdateCommandFtsTrigger = `
	CREATE TRIGGER IF NOT EXISTS update_command_fts_trigger
		AFTER UPDATE ON commands
	BEGIN
		UPDATE commands_fts
		SET
			name = NEW.name,
			command = NEW.command,
			description = NEW.description
		WHERE id = NEW.id;
	END`

	DeleteCommandFtsTrigger = `
	CREATE TRIGGER IF NOT EXISTS delete_command_fts_trigger
		AFTER DELETE ON commands
	BEGIN
		DELETE from commands
		WHERE id = OLD.id;
	END`
)

// queries
const (
	InsertCommandQuery = `
	INSERT INTO 
		commands(id, name, description, command) 
	VALUES (?, ?, ?, ?)`

	InsertParameterPartialQuery = `
	INSERT INTO 
		parameters(id, command, name, description, value)
	VALUES %s`

	GetAllCommandsQuery = `
	SELECT id, name, description, command 
	FROM commands`

	GetCommandbyIDQuery = `
	SELECT 
		id, name, description, command 
	FROM commands
	WHERE id = ?`

	GetParametersByCommandID = `
	SELECT 
		id, name, description, value 
	FROM parameters
	WHERE command = ?`

	SearchCommandQuery = `
	SELECT
		c.id, c.name, c.description, c.command
	FROM commands c
	INNER JOIN commands_fts fts 
		ON c.id = fts.id
	WHERE commands_fts MATCH ?
	ORDER BY bm25(commands_fts, 0, 10, 5, 2)`
)
