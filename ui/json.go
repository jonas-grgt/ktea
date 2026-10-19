package ui

import (
	"bytes"
	"encoding/json"
	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	chrome_styles "github.com/alecthomas/chroma/v2/styles"
	"ktea/styles"
	"strings"
)

func PrettyPrintJson(text string) string {
	builder := &strings.Builder{}
	formatter := formatters.TTY256
	style := chrome_styles.Get(styles.Current.Syntax)

	lexer := lexers.Fallback
	var prettyJSON bytes.Buffer
	var iterator chroma.Iterator
	if strings.Contains(text, "{") &&
		strings.Contains(text, "}") {
		json.Indent(&prettyJSON, []byte(text), "", "\t")
		lexer = lexers.Get("json")
		iterator, _ = lexer.Tokenise(nil, prettyJSON.String())
	} else if strings.Contains(text, "/>") {
		lexer = lexers.Get("XML")
		iterator, _ = lexer.Tokenise(nil, text)
	} else {
		lexer = lexers.Fallback
		iterator, _ = lexer.Tokenise(nil, text)
	}

	formatter.Format(builder, style, iterator)

	return builder.String()
}
