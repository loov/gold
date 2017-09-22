package gold

type JSON struct {
	Flat []struct {
		Path      string                 `json:"path"`
		Variables map[string]interface{} `json:"variables"`
	} `json:"flat"`
}

func (json *JSON) AddInto(tree *Tree, source string) {
	for _, row := range json.Flat {
		tree.Add(source, row.Path, row.Variables)
	}
}
