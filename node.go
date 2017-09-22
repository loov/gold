package gold

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

const PathSeparator = "/"

type Tree struct {
	Sources   []string
	Variables []string

	*Node
}

func NewTree() *Tree {
	tree := &Tree{}
	tree.Node = NewNode("", "")
	return tree
}

// Add adds a single data point
func (tree *Tree) Add(source, path string, variables map[string]interface{}) {
	if strings.HasPrefix(path, PathSeparator) {
		path = path[len(PathSeparator):]
	}

	if !contains(source, tree.Sources) {
		tree.Sources = append(tree.Sources, source)
	}

	newvars := []string{}
	for variable := range variables {
		if !contains(variable, tree.Variables) {
			newvars = append(newvars, variable)
		}
	}
	sort.Strings(newvars)
	tree.Variables = append(tree.Variables, newvars...)

	paths := strings.Split(path, PathSeparator)
	tree.Include(source, nil, paths, variables)
}

// Error is a metric how much values differ
type Error float64

// Node stores information about a Node and different states recursively
type Node struct {
	QualifiedName string
	Name          string
	State         State
	Error         map[string]Error
	Children      map[string]*Node
	ChildrenOrder []string
}

// NewNode returns a new node
func NewNode(qualifiedName, name string) *Node {
	node := &Node{}
	node.QualifiedName = qualifiedName
	node.Name = name
	node.Error = make(map[string]Error)
	node.Children = make(map[string]*Node)
	return node
}

// Include creates child nodes and puts the variables to the deepest level
func (node *Node) Include(source string, prefix, path []string, variables map[string]interface{}) {
	if len(path) == 0 {
		node.State.Add(source, variables)
		return
	}
	var name string
	name, path = path[0], path[1:]
	prefix = append(prefix, name)
	child, ok := node.Children[name]
	if !ok {
		child = NewNode(strings.Join(prefix, PathSeparator), name)
		node.Children[name] = child
		node.ChildrenOrder = append(node.ChildrenOrder, name)
	}
	child.Include(source, prefix, path, variables)
}

// UpdateError calculates recursively error in state and children
func (node *Node) UpdateError() {
	all := node.State.Clone()
	for _, child := range node.Children {
		child.UpdateError()
		all.addError(child.Name, child.Error)
	}
	node.Error = all.CalculateError()
}

func (node *Node) HasErrors() bool {
	for _, err := range node.Error {
		if err != 0 {
			return true
		}
	}
	return false
}

// State is a collection of variables at a given time
type State struct {
	Variables []Variable
}

// Clone creates a copy of all entries.
func (state *State) Clone() State {
	clone := State{}
	clone.Variables = make([]Variable, 0, len(state.Variables))
	for _, variable := range state.Variables {
		clone.Variables = append(clone.Variables, variable.Clone())
	}
	return clone
}

func (state *State) Get(name string) (*Variable, bool) {
	for i := range state.Variables {
		if state.Variables[i].Name == name {
			return &state.Variables[i], true
		}
	}
	return nil, false
}

// mustGet either finds a particular variable, or if one does not exist, creates one
func (state *State) mustGet(name string) *Variable {
	if variable, ok := state.Get(name); ok {
		return variable
	}
	state.Variables = append(state.Variables, Variable{})
	variable := &state.Variables[len(state.Variables)-1]
	variable.Name = name
	return variable
}

// Add adds variables from a particular source to the state
func (state *State) Add(source string, variables map[string]interface{}) {
	for name, value := range variables {
		val := state.mustGet(name)
		val.Add(source, value)
	}
}

// addError is a convenience func for adding a map of Errors
func (state *State) addError(source string, variables map[string]Error) {
	for name, variable := range variables {
		val := state.mustGet(name)
		val.Add(source, variable)
	}
}

// CalculateError calculates error for each variable
func (state *State) CalculateError() map[string]Error {
	r := map[string]Error{}
	for _, variable := range state.Variables {
		r[variable.Name] = variable.CalculateError()
	}
	return r
}

// Variable represents a named data-point from multiple sources
type Variable struct {
	Name    string
	Entries []Entry
}

// Entry remembers a variable values from different sources
type Entry struct {
	Source string
	Value  interface{}
}

func (variable *Variable) Clone() Variable {
	clone := Variable{}
	clone.Name = variable.Name
	clone.Entries = append([]Entry{}, variable.Entries...)
	return clone
}

// Add adds a new data-point from a particular source
func (variable *Variable) Add(source string, value interface{}) {
	variable.Entries = append(variable.Entries, Entry{source, value})
}

// Get a value from a particular source
func (variable *Variable) Get(source string) (interface{}, bool) {
	for _, entry := range variable.Entries {
		if entry.Source == source {
			return entry.Value, true
		}
	}
	return nil, false
}

// CalculateError calculates a value how much the source values differ
func (variable *Variable) CalculateError() Error {
	err := 0.0

	mean := 0.0
	n := 0
	for _, entry := range variable.Entries {
		if v, ok := nativeAsFloat(entry.Value); ok {
			mean += v
			n++
		} else if v, ok := entry.Value.(Error); ok {
			err += float64(v)
		} else {
			panic(fmt.Sprintf("do not know how to calculate error for %v", entry))
		}
	}

	if n >= 2 {
		mean /= float64(n)
		sum2 := 0.0
		for _, entry := range variable.Entries {
			if v, ok := nativeAsFloat(entry.Value); ok {
				d := v - mean
				sum2 += d * d
			}
		}
		err += math.Sqrt(sum2 / float64(n-1))
	}

	return Error(err)
}

func nativeAsFloat(v interface{}) (r float64, ok bool) {
	switch v := v.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case int32:
		return float64(v), true
	case int8:
		return float64(v), true
	case uint:
		return float64(v), true
	case uint64:
		return float64(v), true
	case uint32:
		return float64(v), true
	case uint8:
		return float64(v), true
	}
	return 0, false
}
