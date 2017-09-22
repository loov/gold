package gold

import (
	"fmt"
	"html"
	"io"
	"text/tabwriter"
)

const Epsilon = 0.0001

type Renderer struct {
	Sources   []string
	Variables []string

	LevelLimit int
	ShowZero   bool
	ShowSource bool
	Epsilon    float64
}

func NewRenderer() *Renderer {
	r := &Renderer{}
	r.Epsilon = Epsilon
	r.ShowZero = false
	r.ShowSource = false
	return r
}

func (r *Renderer) HTML(w io.Writer, tree *Tree) {
	sources := r.Sources
	variables := r.Variables

	if len(sources) == 0 {
		sources = tree.Sources
	}
	if len(variables) == 0 {
		variables = tree.Variables
	}

	escape := func(s string) string {
		return html.EscapeString(s)
	}
	write := func(format string, args ...interface{}) {
		fmt.Fprintf(w, format, args...)
	}

	write("<table>")
	write("<style>")
	write(".zero { color: #888; } ")
	write(".source { color: #888; } ")
	write("</style>")
	defer write("</table>")

	{
		write("<thead>")
		write("<tr>")
		write("<th></th>")
		for _, variable := range variables {
			write("<th>%v</th>", escape(variable))
		}
		write("</tr>")
		write("</thead>")
	}

	var render func(node *Node, level int)
	render = func(node *Node, level int) {
		if !r.ShowZero && !node.HasErrors() {
			return
		}
		if level > r.LevelLimit && r.LevelLimit >= 0 {
			return
		}

		{ // print error
			write("<tr>")
			write("<td>%v</td>", escape(node.QualifiedName))
			for _, variable := range variables {
				if value, ok := node.Error[variable]; ok {
					if -r.Epsilon < float64(value) && float64(value) < r.Epsilon {
						write(`<td class="zero">`)
					} else {
						write(`<td>`)
					}
					write("%.2f</td>", value)
				} else {
					write("<td></td>")
				}
			}
			write("</tr>")
		}

		if r.ShowSource {
			cross := make([][]string, len(sources))
			for _, variableName := range variables {
				variable, ok := node.State.Get(variableName)
				if !ok {
					continue
				}
				for k, source := range sources {
					val, ok := variable.Get(source)
					if !ok {
						cross[k] = append(cross[k], "")
						continue
					}
					cross[k] = append(cross[k], fmt.Sprintf("<td>%.2f</td>", val))
				}
			}

			for sourceIndex, values := range cross {
				if empty(values) {
					continue
				}
				write(`<tr class="source">`)
				write("<td>%v</td>", escape(node.QualifiedName+":"+sources[sourceIndex]))
				for _, value := range values {
					if value != "" {
						write(value)
					} else {
						write("<td></td>")
					}
				}
				write("</tr>")
			}
		}

		for _, childName := range node.ChildrenOrder {
			child := node.Children[childName]
			render(child, level+1)
		}
	}

	write("<tbody>")
	render(tree.Node, 0)
	write("</tbody>")
}

func (r *Renderer) Console(w io.Writer, tree *Tree) {
	sources := r.Sources
	variables := r.Variables

	if len(sources) == 0 {
		sources = tree.Sources
	}
	if len(variables) == 0 {
		variables = tree.Variables
	}

	tw := new(tabwriter.Writer)
	tw.Init(w, 4, 8, 4, ' ', 0)
	defer tw.Flush()

	write := func(format string, args ...interface{}) {
		fmt.Fprintf(tw, format, args...)
	}

	{
		write("")
		for _, variable := range variables {
			write("\t%v", variable)
		}
		write("\n")
	}

	var render func(node *Node, level int)
	render = func(node *Node, level int) {
		if !r.ShowZero && !node.HasErrors() {
			return
		}
		if level > r.LevelLimit && r.LevelLimit >= 0 {
			return
		}

		{ // print error
			write("%v", node.QualifiedName)
			for _, variable := range variables {
				if value, ok := node.Error[variable]; ok {
					if -r.Epsilon < float64(value) && float64(value) < r.Epsilon {
						write(``)
					} else {
						write(``)
					}
					write("\t%.2f", value)
				} else {
					write("\t")
				}
			}
			write("\n")
		}

		if r.ShowSource {
			cross := make([][]string, len(sources))
			for _, variableName := range variables {
				variable, ok := node.State.Get(variableName)
				if !ok {
					continue
				}
				for k, source := range sources {
					val, ok := variable.Get(source)
					if !ok {
						cross[k] = append(cross[k], "")
						continue
					}
					cross[k] = append(cross[k], fmt.Sprintf("%.2f", val))
				}
			}

			for sourceIndex, values := range cross {
				if empty(values) {
					continue
				}
				write("%v", node.QualifiedName+":"+sources[sourceIndex])
				for _, value := range values {
					if value != "" {
						write("\t" + value)
					} else {
						write("\t")
					}
				}
				write("\n")
			}
		}

		for _, childName := range node.ChildrenOrder {
			child := node.Children[childName]
			render(child, level+1)
		}
	}

	render(tree.Node, 0)
}
