package gold

// Output is a set of snapshots from a program
type Output struct {
	Source    string     `json:"source"`
	Snapshots []Snapshot `json:"snapshots"`
}

// Snapshot is variables from a particular point in time and space
type Snapshot struct {
	Path      string                 `json:"path"`
	Variables map[string]interface{} `json:"variables"`
}

func (output *Output) AddInto(tree *Tree) {
	for _, snapshot := range output.Snapshots {
		tree.Add(output.Source, snapshot.Path, snapshot.Variables)
	}
}
