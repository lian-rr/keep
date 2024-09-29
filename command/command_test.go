package command

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	textParam := Parameter{
		Name:         "text",
		Description:  "text to echo",
		DefaultValue: "hello",
	}

	tests := []struct {
		name          string
		id            uuid.UUID
		commandName   string
		desc          string
		raw           string
		params        []Parameter
		expectedError string
		assertParams  func(actual []Parameter)
	}{
		{
			name: "missing expected param",
			raw:  "echo 'hello world'",
			params: []Parameter{
				{
					Name:        "missing",
					Description: "this param shouldn't exist",
				},
			},
			expectedError: "param 'missing' not found in the command",
		},
		{
			name:        "with params - happy path",
			commandName: "test command",
			desc:        "command for testing",
			raw:         "echo '{{.text}}'",
			params:      []Parameter{textParam},
			assertParams: func(actual []Parameter) {
				assert.Equal(t, 1, len(actual), "not same amount of params")
				assert.Equal(t, textParam, actual[0], "missing text param")
			},
		},
		{
			name:        "parsing params - happy path",
			commandName: "test command 2",
			desc:        "command for testing 2",
			raw:         "echo '{{.text}} - {{.text2}}'",
			assertParams: func(actual []Parameter) {
				assert.Equal(t, 2, len(actual), "not same amount of params")
				param1 := Parameter{
					Name: "text",
				}
				param2 := Parameter{
					Name: "text2",
				}

				assert.Equal(t, param1.Name, actual[0].Name, "missing text param")
				assert.Equal(t, param2.Name, actual[1].Name, "missing text1 param")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var opts []cmdOpt
			if len(tt.params) > 0 {
				opts = append(opts, WithParams(tt.params))
			}

			cmd, err := New(tt.commandName, tt.desc, tt.raw, opts...)
			if tt.expectedError != "" {
				assert.ErrorContains(t, err, tt.expectedError, "error not the expected")
				return
			}

			assert.NoError(t, err, "unexpected error")

			assert.Equal(t, tt.commandName, cmd.Name, "name not the expected")
			assert.Equal(t, tt.desc, cmd.Description, "description not the expected")
			assert.Equal(t, tt.raw, cmd.Command, "command not the expected")
			tt.assertParams(cmd.Params)
		})
	}
}

func TestCommand_Compile(t *testing.T) {
	id, err := uuid.NewV6()
	require.NoError(t, err)

	tests := []struct {
		name           string
		cmd            Command
		args           []Argument
		expectedOutput string
		expectedError  string
	}{
		{
			name: "invalid command",
			cmd: Command{
				ID:      id,
				Name:    "test command",
				Command: "invalid{{.test}",
			},
			expectedError: "invalid command",
		},
		{
			name: "invalid number of arguments",
			cmd: Command{
				ID:      id,
				Name:    "test command",
				Command: "echo {{.text}}",
				Params: []Parameter{
					{
						Name: "text",
					},
				},
			},
			args: []Argument{
				{
					Name:  "text",
					Value: "hello",
				},
				{
					Name:  ".text",
					Value: "bye",
				},
			},
			expectedError: ErrInvalidNumOfParams.Error(),
		},
		{
			name: "happy path",
			cmd: Command{
				ID:      id,
				Name:    "test command",
				Command: "echo '{{.text}} - {{.text2}}'",
				Params: []Parameter{
					{
						Name: "text",
					},
					{
						Name: "text2",
					},
				},
			},
			args: []Argument{
				{
					Name:  "text",
					Value: "hello",
				},
				{
					Name:  "text2",
					Value: "bye",
				},
			},
			expectedOutput: "echo 'hello - bye'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := tt.cmd.Compile(tt.args)
			if tt.expectedError != "" {
				assert.ErrorContains(t, err, tt.expectedError, "error message not the expected")
				return
			}

			assert.NoError(t, err, "error unexpected")
			assert.Equal(t, tt.expectedOutput, out, "out command not the expected")
		})
	}
}
