package command

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"text/template"

	"github.com/google/uuid"
)

const paramsRegex = `{{\s?\.\w+\s?}}`

var reg = regexp.MustCompile(paramsRegex)

var ErrInvalidNumOfParams = errors.New("invalid number of params provided")

type (
	// Command represents a shell.
	Command struct {
		ID          uuid.UUID
		Name        string
		Description string
		Command     string
		Params      []Parameter
	}

	// Parameter represents the Command Parameter
	Parameter struct {
		ID           uuid.UUID
		Name         string
		Description  string
		DefaultValue string
	}
	// Argument represents the command arguments to place in the params
	Argument struct {
		Name  string
		Value string
	}
)

type cmdOpt func(*Command) error

// New returns a new Command.
func New(name string, desc string, cmd string, opts ...cmdOpt) (Command, error) {
	id, err := uuid.NewV6()
	if err != nil {
		return Command{}, err
	}

	cont := Command{
		ID:          id,
		Name:        name,
		Description: desc,
		Command:     cmd,
	}

	for _, opt := range opts {
		if err := opt(&cont); err != nil {
			return Command{}, err
		}
	}

	if len(cont.Params) == 0 {
		cont.Params = parseParams(cmd)
	}

	return cont, nil
}

// Compile returns the command with the arguments applied.
func (c Command) Compile(args []Argument) (string, error) {
	tmp, err := template.New("command").Parse(c.Command)
	if err != nil {
		return "", fmt.Errorf("invalid command: %w", err)
	}

	if len(args) != len(c.Params) {
		return "", ErrInvalidNumOfParams
	}

	arguments := make(map[string]string, len(args))
	for _, arg := range args {
		arguments[arg.Name] = arg.Value
	}

	var buffer bytes.Buffer
	if err := tmp.Execute(&buffer, arguments); err != nil {
		return "", err
	}

	return buffer.String(), nil
}

func parseParams(raw string) []Parameter {
	rawParams := reg.FindAllString(raw, -1)

	params := make([]Parameter, 0, len(rawParams))
	for _, rp := range rawParams {
		id, _ := uuid.NewV6()
		param := Parameter{
			ID:   id,
			Name: rp[3 : len(rp)-2],
		}

		params = append(params, param)
	}

	return params
}

// WithParams used to pass the params to the Command.
// Returns an error if the param is not found.
func WithParams(params []Parameter) cmdOpt {
	return func(c *Command) error {
		for _, param := range params {
			if !strings.Contains(c.Command, fmt.Sprintf("{{.%s}}", param.Name)) {
				return fmt.Errorf("param '%s' not found in the command", param.Name)
			}

			c.Params = append(c.Params, param)
		}

		return nil
	}
}
