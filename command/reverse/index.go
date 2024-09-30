package reverse

import "github.com/lian_rr/keep/command"

type set map[string]struct{}

type tokenizer interface {
	tokenize(str string, exs set) set
}

// Index contains the reverse index data.
type Index struct {
	memory    map[string]set
	tokenizer basicTokenizer
}

// New returns a new Index
func New() *Index {
	return &Index{
		memory: make(map[string]set),
		tokenizer: basicTokenizer{
			minLenght: 3,
		},
	}
}

// Search returns the Command IDs for the search term.
func (idx *Index) Search(terms []string) ([]string, error) {
	tokens := idx.tokenizer.tokenize(terms, nil)
	tokens, err := idx.tokenizer.stemsFilter(tokens)
	if err != nil {
		return nil, err
	}

	foundIDs := make(set)
	for token := range tokens {
		if ids, ok := idx.memory[token]; ok {
			for id := range ids {
				if _, ok := foundIDs[id]; !ok {
					foundIDs[id] = struct{}{}
				}
			}
		}
	}

	results := make([]string, 0, len(foundIDs))
	for id := range foundIDs {
		results = append(results, id)
	}

	return results, nil
}

// Add adds a new command to the index.
// Returns an error if there were an issue adding the command.
// Command description does not get indexed.
func (idx *Index) Add(cmd command.Command) error {
	tokens, err := idx.tokenize(cmd)
	if err != nil {
		return err
	}

	for token := range tokens {
		if ids, ok := idx.memory[token]; ok {
			ids[cmd.ID.String()] = struct{}{}
		} else {
			idx.memory[token] = set{cmd.ID.String(): {}}
		}
	}

	return nil
}

// Remove removes a command from the index.
// Returns an error if there were an issue removing the command.
func (idx *Index) Remove(cmd command.Command) error {
	tokens, err := idx.tokenize(cmd)
	if err != nil {
		return err
	}

	for token := range tokens {
		if ids, ok := idx.memory[token]; ok {
			delete(ids, cmd.ID.String())
		}
	}

	return nil
}

// Update updates the index based on the changes in the command.
func (idx *Index) Update(oldCmd, newCmd command.Command) error {
	oldTokens, err := idx.tokenize(oldCmd)
	if err != nil {
		return err
	}

	newTokens, err := idx.tokenize(newCmd)
	if err != nil {
		return err
	}

	id := oldCmd.ID.String()
	for token := range oldTokens {
		if _, ok := newTokens[token]; !ok {
			idx.removeFromIndex(token, id)
		}
	}

	for token := range newTokens {
		if _, ok := oldTokens[token]; !ok {
			idx.addToIndex(token, id)
		}
	}

	return nil
}

func (idx *Index) tokenize(cmd command.Command) (set, error) {
	exl := make(set)
	for _, param := range cmd.Params {
		exl[param.Name] = struct{}{}
	}

	tokens := idx.tokenizer.tokenizeStr(cmd.Name, exl)
	tokens, err := idx.tokenizer.stemsFilter(tokens)
	if err != nil {
		return nil, err
	}

	tokens = idx.tokenizer.tokenizeStr(cmd.Command, nil)
	tokens, err = idx.tokenizer.stemsFilter(tokens)
	if err != nil {
		return nil, err
	}

	return tokens, nil
}

func (idx *Index) addToIndex(token, id string) {
	if ids, ok := idx.memory[token]; ok {
		ids[id] = struct{}{}
	} else {
		idx.memory[token] = set{id: {}}
	}
}

func (idx *Index) removeFromIndex(token, id string) {
	if ids, ok := idx.memory[token]; ok {
		delete(ids, id)
	}
}
