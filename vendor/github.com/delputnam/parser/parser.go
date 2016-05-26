// Package parser is an extensible parser that unmarshalls
// multiple data formats and extracts it into go maps
package parser

import (
	"encoding/json"
	"errors"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v2"
)

type (
	//ParseFunc is an interface for a function that parses the input
	ParseFunc func(string) (map[string]interface{}, error)
)

//Parser contains the available parsers
type Parser struct {
	parsers map[string]ParseFunc
}

var (
	// ErrUnknownParser is an error indicating that a parser for the requested
	// inputType is not registered.
	errUnknownParser = errors.New("parser: no parser for requested input type")
)

//NewParser creates a new Parser instance
func NewParser() *Parser {
	p := &Parser{parsers: make(map[string]ParseFunc)}

	// add the default parsers
	p.Handle("json", JSONHandler)
	p.Handle("toml", TOMLHandler)
	p.Handle("tml", TOMLHandler)
	p.Handle("yaml", YAMLHandler)
	p.Handle("yml", YAMLHandler)

	return p
}

//Handle registers a handler for the given ext
//This is public so new parsers can be registered
func (p *Parser) Handle(inputType string, fn ParseFunc) {
	p.parsers[inputType] = fn
}

// Parse deserializes the data contained in input based on the ext
func (p *Parser) Parse(inputType string, input string) (map[string]interface{}, error) {
	return p.parse(inputType, input)
}

func (p *Parser) parse(inputType string, input string) (data map[string]interface{}, err error) {
	if h, exists := p.parsers[inputType]; exists {
		data, err = h(input)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, errUnknownParser
	}

	return data, nil

}

//JSONHandler decodes json intput into a go map[string]interface{}
func JSONHandler(input string) (map[string]interface{}, error) {
	var out interface{}
	err := json.Unmarshal([]byte(input), &out)
	if err != nil {
		return nil, err
	}
	return out.(map[string]interface{}), nil
}

//YAMLHandler decodes yaml input into a go map[string]interface{}
func YAMLHandler(input string) (map[string]interface{}, error) {
	out := make(map[string]interface{})
	err := yaml.Unmarshal([]byte(input), out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

//TOMLHandler decodes toml imput into a go map[string]interface{}
func TOMLHandler(input string) (map[string]interface{}, error) {
	out := make(map[string]interface{})
	_, err := toml.Decode(input, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}
