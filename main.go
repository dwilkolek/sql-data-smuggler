package main

import (
	"log"
	"os"
	"path"
	"regexp"
	"strings"
)

func main() {
	var example = "example-models"
	rootModel := readModelDir(example)
	log.Default().Printf("Root model: %v", rootModel)
	prepareExecutionPlan(rootModel)
}

type model struct {
	name       string
	path       string
	children   []model
	sqlSources []string
}

func readModelDir(dirPath string) model {

	info, err := os.Lstat(dirPath)

	model := model{
		name:       dirPath,
		path:       dirPath,
		children:   make([]model, 0),
		sqlSources: make([]string, 0),
	}
	if !info.IsDir() {
		model.name = model.name[:len(model.name)-strings.LastIndex(model.name, ".")]
	}
	files, err := os.ReadDir(dirPath)
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range files {
		if file.IsDir() {
			subModel := readModelDir(path.Join(dirPath, file.Name()))
			model.children = append(model.children, subModel)
		} else {
			model.sqlSources = append(model.sqlSources, file.Name())
		}
	}

	return model
}

type step struct {
	id      int
	file    string
	parents []string
}
type executionPlan struct {
	model model
	steps []step
}

func (m model) files() []string {
	files := make([]string, len(m.sqlSources))
	for i, file := range m.sqlSources {
		files[i] = path.Join(m.path, file)
	}
	for _, child := range m.children {
		files = append(files, child.files()...)
	}
	return files
}

var placeholderRegexp = regexp.MustCompile(`{{([\s]*)([a-z]+)\((.*)\)([\s]*)}}`)

func prepareExecutionPlan(model model) executionPlan {
	files := model.files()
	executionPlan := executionPlan{
		model: model,
		steps: make([]step, len(files)),
	}
	dependencies := make(map[string][]string)
	log.Default().Printf("files %v", files)
	for _, file := range files {
		contentBytes, _ := os.ReadFile(file)
		content := string(contentBytes)
		log.Default().Printf("-----file before %s ------\n%s\n----------", file, content)
		placeholders := placeholderRegexp.FindAllString(content, -1)

		log.Default().Printf("placeholders %v", placeholders)
		for _, placeholder := range placeholders {
			replace, dependencyList := findReplacements(placeholder)
			content = strings.Replace(content, placeholder, replace, 1)
			dependencies[file] = dependencyList
		}

		log.Default().Printf("-----file after %s ------\n%s\n----------", file, content)
	}

	return executionPlan
}

func findReplacements(placeholder string) (string, []string) {
	subMatches := placeholderRegexp.FindStringSubmatch(placeholder)
	log.Default().Printf("findReplacements %v", subMatches)
	function := subMatches[2]
	args := subMatches[3]
	dependencies := make([]string, 0)
	replacement := ""
	log.Default().Printf("function %s", function)
	if function == "ref" {
		replacement = strings.Trim(args, "\"'")
		dependencies = append(dependencies, replacement)
	} else if function == "source" {
		argList := strings.Split(args, ",")
		argsNorm := make([]string, len(argList))
		for i, arg := range argList {
			argsNorm[i] = strings.Trim(arg, "\"' ")
		}
		replacement = strings.ToLower(strings.Join(argsNorm, "."))
		dependencies = append(dependencies, replacement)
	}
	log.Default().Printf("Replaceing %s with %s", subMatches[0], replacement)
	return replacement, dependencies
}
