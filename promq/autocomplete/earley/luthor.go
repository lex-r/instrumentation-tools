/*
Copyright 2020 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package earley

// luthor is our lexer, of course.

import (
	"fmt"
	"strings"

	"github.com/lex-r/promq/notstdlib/sets"

	"github.com/prometheus/prometheus/promql"

	"github.com/lex-r/promq/debug"
)

type Tokens []Tokhan

func (ws Tokens) Vals() []string {
	v := make([]string, len(ws))
	for i, w := range ws {
		v[i] = w.Val
	}
	return v
}

func (ws Tokens) Types() []string {
	v := make([]string, len(ws))
	for i, w := range ws {
		v[i] = string(w.Type)
	}
	return v
}
func (ws Tokens) Print() {
	for _, w := range ws {
		debug.Debugln(w.String())
	}
}
func (ws Tokens) PrintVals() {
	b := strings.Builder{}

	for _, w := range ws {
		b.WriteString(w.Val)
		b.WriteString("|")
	}
	fmt.Println(b.String())
}

func (ws Tokens) Last() Tokhan {
	return ws[len(ws)-2]
}

func (ws Tokens) Compare(tks2 Tokens) int {
	i := 0
	for i, t := range ws {
		// return positive i if ws and tks2 have partial overlap
		if i >= len(tks2) || !t.equals(tks2[i]) {
			return i
		}
	}
	// return negative i if tks2 covers ws
	if len(tks2) > len(ws) {
		return 0 - i
	}
	// return 0 if they are equal
	return 0
}

type TypedToken interface {
	GetTokenType() TokenType
}

type TokenType string

func newStringSet(items ...TokenType) sets.String {
	ss := sets.String{}
	for _, item := range items {
		ss.Insert(string(item))
	}
	return ss
}

const (
	ID                   TokenType = "identifier"
	METRIC_ID            TokenType = "metric-identifier"
	METRIC_LABEL_SUBTYPE TokenType = "metric-label-identifier"
	FUNCTION_SCALAR_ID   TokenType = "function-scalar-identifier"
	FUNCTION_VECTOR_ID   TokenType = "function-vector-identifier"

	OPERATOR TokenType = "operator"
	//binary operators
	ARITHMETIC  TokenType = "arithmetic"
	COMPARISION TokenType = "comparision"
	SET         TokenType = "set"
	//label match operator
	LABELMATCH TokenType = "label-match"
	// unary operators
	UNARY_OP TokenType = "unary-op"

	AGGR_OP TokenType = "aggregator_operation"

	//keywords
	KEYWORD    TokenType = "keyword"
	AGGR_KW    TokenType = "aggregator_keyword"
	BOOL_KW    TokenType = "bool-keyword"
	OFFSET_KW  TokenType = "offset-keyword"
	GROUP_SIDE TokenType = "group-side"
	GROUP_KW   TokenType = "group-keyword"

	LEFT_BRACE    TokenType = "leftbrace"
	RIGHT_BRACE   TokenType = "rightbrace"
	LEFT_PAREN    TokenType = "leftparen"
	RIGHT_PAREN   TokenType = "rightparen"
	LEFT_BRACKET  TokenType = "leftbracket"
	RIGHT_BRACKET TokenType = "rightbracket"
	COMMA         TokenType = "comma"
	COLON         TokenType = "colon"
	STRING        TokenType = "string"
	NUM           TokenType = "number"
	DURATION      TokenType = "duration"
	EOF           TokenType = "EOF"
	UNKNOWN       TokenType = "unknown"
)

// Tokhan contains the essential bits of data we need
// for processing a single lexical unit.
type Tokhan struct {
	StartPos int
	EndPos   int
	Type     TokenType
	ItemType promql.ItemType
	Val      string
	_index   int
}

func (t Tokhan) isEof() bool {
	return t.ItemType == promql.EOF
}

func (t Tokhan) String() string {
	return fmt.Sprintf("Tokhan.Val(%v) Type(%v) StartEnd[%v:%v]",
		t.Val,
		t.Type,
		t.StartPos,
		t.EndPos,
	)
}

func (t Tokhan) equals(t2 Tokhan) bool {
	return t.Val == t2.Val && t.StartPos == t2.StartPos && t.EndPos == t2.EndPos
}

func extractWords(query string) Tokens {
	words := extractTokensWithOffset(query, 0)
	words.Print()
	return words
}

// todo(rant):  we should probably just hand-roll a parser for this. I am
// todo(cont):  not fond of the way this lexer encodes random syntactical
// todo(cont):  rules during lexing
func extractTokensWithOffset(query string, offset int) (words Tokens) {
	l := promql.Lex(query)
	i := 0
	for {
		currItem := promql.Item{}
		l.NextItem(&currItem)
		if currItem.Typ == promql.EOF {
			words = append(words, createTokenFromItem(currItem, offset))
			break
		}

		// recurse our lexer on a sub-query string. We do this specifically to accommodate
		// strings like `start(label='value)end` where we want as output:
		// a linked list of tokens like this:
		// "start" <-> "(" <-> "label" <-> "=" <-> "'value" <-> ")" <-> "end"
		if currItem.Typ == promql.ERROR {
			substring := query[currItem.Pos:]
			// we're recursing and found an error already abort
			if i == 0 {
				break
			}
			subWords := extractTokensWithOffset(substring, int(currItem.Pos))
			if len(subWords) > 0 {
				words = append(words, subWords...)
			}
			return
		}
		words = append(words, createTokenFromItem(currItem, offset))
		i++
	}
	return
}

func createTokenFromItem(item promql.Item, offset int) Tokhan {
	return Tokhan{
		Val:      item.Val,
		ItemType: item.Typ,
		Type:     mapParserItemTypeToTokhanType(item),
		StartPos: int(item.Pos) + offset,
		EndPos:   int(item.PositionRange().End),
	}
}

func mapParserItemTypeToTokhanType(item promql.Item) TokenType {
	t := item.Typ
	switch {
	case t == promql.BY, t == promql.WITHOUT:
		return AGGR_KW
	case t == promql.OFFSET:
		return OFFSET_KW
	case t == promql.BOOL:
		return BOOL_KW
	case t == promql.GROUP_LEFT, t == promql.GROUP_RIGHT:
		return GROUP_SIDE
	case t == promql.IGNORING, t == promql.ON:
		return GROUP_KW
	case t == promql.EOF:
		return EOF
	case t == promql.STRING:
		return STRING
	case isAggregator(t):
		return AGGR_OP
	case t == promql.METRIC_IDENTIFIER:
		return METRIC_ID
	case isScalarFunction(item):
		return FUNCTION_SCALAR_ID
	case isVectorFunction(item):
		return FUNCTION_VECTOR_ID
	case t == promql.IDENTIFIER:
		return ID
	case t == promql.LEFT_BRACE:
		return LEFT_BRACE
	case t == promql.RIGHT_BRACE:
		return RIGHT_BRACE
	case t == promql.LEFT_PAREN:
		return LEFT_PAREN
	case t == promql.RIGHT_PAREN:
		return RIGHT_PAREN
	case t == promql.LEFT_BRACKET:
		return LEFT_BRACKET
	case t == promql.RIGHT_BRACKET:
		return RIGHT_BRACKET
	case t == promql.DURATION:
		return DURATION
	case t == promql.ADD, t == promql.SUB, t == promql.MUL, t == promql.DIV, t == promql.MOD, t == promql.POW:
		return ARITHMETIC
	case t == promql.LAND, t == promql.LOR, t == promql.LUNLESS:
		return SET
	case isOperator(t):
		return OPERATOR
	case t == promql.COMMA:
		return COMMA
	case t == promql.COLON:
		return COLON
	case t == promql.NUMBER:
		return NUM
	default:
		return UNKNOWN
	}
}

// need to explicitly extract this function since it's private
// in prometheus 2.16
func isAggregator(item promql.ItemType) bool {
	return item > 57386 && item < 57398
}

func isOperator(item promql.ItemType) bool {
	return item > 57367 && item < 57385
}

func isScalarFunction(item promql.Item) bool {
	for _, v := range sets.StringKeySet(scalarFunctions).List() {
		if v == item.Val {
			return true
		}
	}
	return false
}

func isVectorFunction(item promql.Item) bool {
	for _, v := range sets.StringKeySet(vectorFunctions).List() {
		if v == item.Val {
			return true
		}
	}
	return false
}
