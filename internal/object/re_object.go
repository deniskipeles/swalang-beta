// pylearn/internal/object/re_object.go
package object

import (
	"fmt"
	"regexp"

	"github.com/deniskipeles/pylearn/internal/constants"
)

const (
	REGEX_PATTERN_OBJ ObjectType = constants.OBJECT_TYPE_REGEX_PATTERN
	REGEX_MATCH_OBJ   ObjectType = constants.OBJECT_TYPE_REGEX_MATCH
	REGEX_ITERATOR_OBJ ObjectType = constants.OBJECT_TYPE_REGEX_ITERATOR
)

// --- Regex Pattern Object ---
// Represents a compiled regular expression.
type Pattern struct {
	Regex   *regexp.Regexp
	Pattern string // Original pattern string
}

func (p *Pattern) Type() ObjectType { return REGEX_PATTERN_OBJ }
func (p *Pattern) Inspect() string {
	return fmt.Sprintf("re.compile(%q)", p.Pattern)
}

// --- Regex Match Object ---
// Represents a single match of a regular expression.
type Match struct {
	matchStrings []string       // The full match and all subgroups. matchStrings[0] is the full match.
	namedGroups  map[string]int // Map of group name to its index in matchStrings
	sourceString string         // The original string searched
	span         [2]int         // [start, end) of the full match in sourceString
}

func (m *Match) Type() ObjectType { return REGEX_MATCH_OBJ }
func (m *Match) Inspect() string {
	return fmt.Sprintf("<re.Match object; span=(%d, %d), match=%q>", m.span[0], m.span[1], m.matchStrings[0])
}

// RegexIterator is the object returned by pattern.finditer()
type RegexIterator struct {
	pattern      *Pattern
	subject      string
	allMatches   [][]int // Slice of start/end indices from FindAllStringSubmatchIndex
	currentIndex int
}

func (ri *RegexIterator) Type() ObjectType { return REGEX_ITERATOR_OBJ }
func (ri *RegexIterator) Inspect() string {
	return fmt.Sprintf("<callable_iterator object for re.Pattern at %p>", ri)
}

// Next implements the object.Iterator interface
func (ri *RegexIterator) Next() (Object, bool) {
	if ri.currentIndex >= len(ri.allMatches) {
		return nil, true // Stop iteration
	}

	loc := ri.allMatches[ri.currentIndex]
	ri.currentIndex++

	// Recreate the submatch strings for this specific match
	submatchStrings := make([]string, ri.pattern.Regex.NumSubexp()+1)
	for i := 0; i < len(submatchStrings); i++ {
		start, end := loc[i*2], loc[i*2+1]
		if start != -1 { // -1 indicates an optional group that didn't match
			submatchStrings[i] = ri.subject[start:end]
		}
	}

	namedGroups := make(map[string]int)
	for i, name := range ri.pattern.Regex.SubexpNames() {
		if i > 0 && name != "" {
			namedGroups[name] = i
		}
	}

	// Create and return a new Match object
	return &Match{
		matchStrings: submatchStrings,
		namedGroups:  namedGroups,
		sourceString: ri.subject,
		span:         [2]int{loc[0], loc[1]},
	}, false
}

var _ Iterator = (*RegexIterator)(nil)
// --- Go functions for Pattern methods ---

func patternSearch(ctx ExecutionContext, args ...Object) Object {
	self, ok := args[0].(*Pattern)
	if !ok {
		return NewError(constants.TypeError, "search() must be called on a compiled re.Pattern object")
	}
	if len(args) != 2 {
		return NewError(constants.TypeError, "search() takes exactly 1 argument (string)")
	}
	str, ok := args[1].(*String)
	if !ok {
		return NewError(constants.TypeError, "search() argument must be a string")
	}

	loc := self.Regex.FindStringSubmatchIndex(str.Value)
	if loc == nil {
		return NULL
	}

	submatchStrings := self.Regex.FindStringSubmatch(str.Value)
	namedGroups := make(map[string]int)
	for i, name := range self.Regex.SubexpNames() {
		if i > 0 && name != "" {
			namedGroups[name] = i
		}
	}

	return &Match{
		matchStrings: submatchStrings,
		namedGroups:  namedGroups,
		sourceString: str.Value,
		span:         [2]int{loc[0], loc[1]},
	}
}

func patternMatch(ctx ExecutionContext, args ...Object) Object {
	self, ok := args[0].(*Pattern)
	if !ok {
		return NewError(constants.TypeError, "match() must be called on a compiled re.Pattern object")
	}
	if len(args) != 2 {
		return NewError(constants.TypeError, "match() takes exactly 1 argument (string)")
	}
	str, ok := args[1].(*String)
	if !ok {
		return NewError(constants.TypeError, "match() argument must be a string")
	}

	// Enforce matching only at the beginning of the string
	loc := self.Regex.FindStringSubmatchIndex(str.Value)
	if loc == nil || loc[0] != 0 {
		return NULL
	}

	submatchStrings := self.Regex.FindStringSubmatch(str.Value)
	namedGroups := make(map[string]int)
	for i, name := range self.Regex.SubexpNames() {
		if i > 0 && name != "" {
			namedGroups[name] = i
		}
	}

	return &Match{
		matchStrings: submatchStrings,
		namedGroups:  namedGroups,
		sourceString: str.Value,
		span:         [2]int{loc[0], loc[1]},
	}
}

func patternFindAll(ctx ExecutionContext, args ...Object) Object {
	self, ok := args[0].(*Pattern)
	if !ok {
		return NewError(constants.TypeError, "findall() must be called on a compiled re.Pattern object")
	}
	if len(args) != 2 {
		return NewError(constants.TypeError, "findall() takes exactly 1 argument (string)")
	}
	str, ok := args[1].(*String)
	if !ok {
		return NewError(constants.TypeError, "findall() argument must be a string")
	}

	matches := self.Regex.FindAllStringSubmatch(str.Value, -1)
	if matches == nil {
		return &List{Elements: []Object{}}
	}

	numGroups := len(self.Regex.SubexpNames()) - 1

	results := &List{Elements: []Object{}}

	if numGroups == 0 { // No capturing groups
		for _, match := range self.Regex.FindAllString(str.Value, -1) {
			results.Elements = append(results.Elements, &String{Value: match})
		}
	} else if numGroups == 1 { // One capturing group
		for _, match := range matches {
			results.Elements = append(results.Elements, &String{Value: match[1]})
		}
	} else { // Multiple capturing groups
		for _, match := range matches {
			tupleElements := make([]Object, numGroups)
			for i := 0; i < numGroups; i++ {
				tupleElements[i] = &String{Value: match[i+1]}
			}
			results.Elements = append(results.Elements, &Tuple{Elements: tupleElements})
		}
	}

	return results
}

func patternFindIter(ctx ExecutionContext, args ...Object) Object {
	self, ok := args[0].(*Pattern)
	if !ok {
		return NewError(constants.TypeError, "finditer() must be called on a compiled re.Pattern object")
	}
	if len(args) != 2 {
		return NewError(constants.TypeError, "finditer() takes exactly 1 argument (string)")
	}
	str, ok := args[1].(*String)
	if !ok {
		return NewError(constants.TypeError, "finditer() argument must be a string")
	}

	// Find all match indices at once. This is efficient.
	allMatchesIndices := self.Regex.FindAllStringSubmatchIndex(str.Value, -1)

	// Return a new iterator object that will yield Match objects on demand.
	return &RegexIterator{
		pattern:      self,
		subject:      str.Value,
		allMatches:   allMatchesIndices,
		currentIndex: 0,
	}
}
// GetObjectAttribute for Pattern
func (p *Pattern) GetObjectAttribute(ctx ExecutionContext, name string) (Object, bool) {
	makeMethod := func(methodName string, goFn BuiltinFunction) *Builtin {
		return &Builtin{
			Name: "re.Pattern." + methodName,
			Fn: func(callCtx ExecutionContext, scriptProvidedArgs ...Object) Object {
				methodArgs := make([]Object, 0, 1+len(scriptProvidedArgs))
				methodArgs = append(methodArgs, p) // Prepend self (the Pattern 'p')
				methodArgs = append(methodArgs, scriptProvidedArgs...)
				return goFn(callCtx, methodArgs...)
			},
		}
	}

	switch name {
	case "search":
		return makeMethod("search", patternSearch), true
	case "match":
		return makeMethod("match", patternMatch), true
	case "findall":
		return makeMethod("findall", patternFindAll), true
	case "finditer":
		return makeMethod("finditer", patternFindIter), true
	case "pattern":
		return &String{Value: p.Pattern}, true
	}
	return nil, false
}

// --- Go functions for Match methods ---

func matchGroup(ctx ExecutionContext, args ...Object) Object {
	self, ok := args[0].(*Match)
	if !ok {
		return NewError(constants.TypeError, "group() must be called on a re.Match object")
	}

	if len(args) == 1 { // group() or group(0)
		return &String{Value: self.matchStrings[0]}
	}

	results := []Object{}
	for i := 1; i < len(args); i++ {
		groupArg := args[i]
		var groupStr string
		found := false

		switch arg := groupArg.(type) {
		case *Integer:
			idx := int(arg.Value)
			if idx >= 0 && idx < len(self.matchStrings) {
				groupStr = self.matchStrings[idx]
				found = true
			}
		case *String:
			if idx, ok := self.namedGroups[arg.Value]; ok {
				if idx < len(self.matchStrings) {
					groupStr = self.matchStrings[idx]
					found = true
				}
			}
		default:
			return NewError(constants.TypeError, "group() arguments must be int or str")
		}

		if !found {
			return NewError(constants.IndexError, "no such group")
		}
		results = append(results, &String{Value: groupStr})
	}

	if len(results) == 1 {
		return results[0]
	}

	return &Tuple{Elements: results}
}

func matchGroups(ctx ExecutionContext, args ...Object) Object {
	self, ok := args[0].(*Match)
	if !ok {
		return NewError(constants.TypeError, "groups() must be called on a re.Match object")
	}

	if len(args) != 1 {
		return NewError(constants.TypeError, "groups() takes no arguments")
	}

	if len(self.matchStrings) <= 1 {
		return &Tuple{Elements: []Object{}}
	}

	elements := make([]Object, len(self.matchStrings)-1)
	for i, s := range self.matchStrings[1:] {
		elements[i] = &String{Value: s}
	}
	return &Tuple{Elements: elements}
}

func matchSpan(ctx ExecutionContext, args ...Object) Object {
	self, ok := args[0].(*Match)
	if !ok {
		return NewError(constants.TypeError, "span() must be called on a re.Match object")
	}
	if len(args) != 1 {
		return NewError(constants.TypeError, "span() takes no arguments")
	}

	return &Tuple{Elements: []Object{
		&Integer{Value: int64(self.span[0])},
		&Integer{Value: int64(self.span[1])},
	}}
}

// GetObjectAttribute for Match
// GetObjectAttribute for Match is needed for token.lastgroup and token.group()
func (m *Match) GetObjectAttribute(ctx ExecutionContext, name string) (Object, bool) {
	makeMethod := func(methodName string, goFn BuiltinFunction) *Builtin {
		return &Builtin{
			Name: "re.Match." + methodName,
			Fn: func(callCtx ExecutionContext, scriptProvidedArgs ...Object) Object {
				methodArgs := make([]Object, 0, 1+len(scriptProvidedArgs))
				methodArgs = append(methodArgs, m) // Prepend self
				methodArgs = append(methodArgs, scriptProvidedArgs...)
				return goFn(callCtx, methodArgs...)
			},
		}
	}
	switch name {
	case "group":
		return makeMethod("group", matchGroup), true
	case "groups":
		return makeMethod("groups", matchGroups), true
	case "span":
		return makeMethod("span", matchSpan), true
	case "lastgroup":
		// Find the rightmost group that matched.
		lastMatchIndex := -1
		var lastName string
		for name, index := range m.namedGroups {
			if index > lastMatchIndex && index < len(m.matchStrings) && m.matchStrings[index] != "" {
				lastMatchIndex = index
				lastName = name
			}
		}
		if lastName == "" {
			return NULL, true
		}
		return &String{Value: lastName}, true
	}
	return nil, false
}

var _ Object = (*Pattern)(nil)
var _ AttributeGetter = (*Pattern)(nil)
var _ Object = (*Match)(nil)
var _ AttributeGetter = (*Match)(nil)