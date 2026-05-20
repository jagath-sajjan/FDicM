package spell

import "github.com/sahilm/fuzzy"

var corpus = []string{
	"algorithm", "application", "api", "architecture", "binary", "boolean",
	"buffer", "cache", "compiler", "computer", "communication", "database",
	"debugging", "design", "development", "dictionary", "encryption", "function",
	"framework", "golang", "hardware", "interface", "iteration", "json", "language",
	"library", "memory", "network", "object", "program", "programming", "pointer",
	"runtime", "server", "structure", "syntax", "system", "terminal", "variable",
}

func Suggest(input string) []string {
	if input == "" {
		return []string{}
	}
	matches := fuzzy.Find(input, corpus)
	out := []string{}
	for i, m := range matches {
		if i >= 8 {
			break
		}
		out = append(out, corpus[m.Index])
	}
	return out
}
