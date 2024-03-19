package template

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"os"

	"github.com/labstack/echo/v4"
)

type Template struct {
	templates map[string]*template.Template
}

type CompRecipe struct {
	Name      string `json:"name"`
	Component string `json:"component"`
}

type Recipe struct {
	Root                string       `json:"root"`
	PageDefinition      []string     `json:"page_definition"`
	ComponentDefinition []string     `json:"component_definition"`
	Recipes             []CompRecipe `json:"recipes"`
}

func New() *Template {
	templates := make(map[string]*template.Template)
	// open "public/recipe.json"
	file, err := os.Open("public/recipe.json")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	// decode json
	decoder := json.NewDecoder(file)
	recipe := Recipe{}
	err = decoder.Decode(&recipe)
	if err != nil {
		log.Fatal(err)
	}
	if recipe.Root == "" {
		log.Fatal("root is required")
	}
	if recipe.PageDefinition == nil || len(recipe.PageDefinition) == 0 {
		log.Fatal("page_definition is required")
	}
	if recipe.ComponentDefinition == nil || len(recipe.ComponentDefinition) == 0 {
		log.Fatal("component_definition is required")
	}

	for i, comp := range recipe.ComponentDefinition {
		recipe.ComponentDefinition[i] = fmt.Sprintf("%s/%s.go.html", recipe.Root, comp)
	}

	for i, comp := range recipe.PageDefinition {
		recipe.PageDefinition[i] = fmt.Sprintf("%s/%s.go.html", recipe.Root, comp)
	}

	for _, comp := range recipe.Recipes {
		compName := fmt.Sprintf("%s/%s.go.html", recipe.Root, comp.Component)
		// parse component
		componentList := []string{
			compName,
		}
		componentList = append(componentList, recipe.ComponentDefinition...)
		component := template.Must(template.ParseFiles(componentList...))
		if err != nil {
			log.Printf("Error parsing component %s: %s", comp.Name, err)
			continue
		}
		templates[fmt.Sprintf("%s-comp", comp.Name)] = component
		// parse page
		pageList := []string{
			compName,
		}
		pageList = append(pageList, recipe.PageDefinition...)
		page, err := template.ParseFiles(pageList...)
		if err != nil {
			log.Printf("Error parsing page %s: %s", comp.Name, err)
			continue
		}
		templates[fmt.Sprintf("%s-page", comp.Name)] = page
	}
	return &Template{
		templates: templates,
	}
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	var ctx map[string]any = map[string]any{}
	var ok bool
	if data != nil {
		ctx, ok = data.(map[string]any)
		if !ok {
			log.Printf("no pongo context found: %v\n", data)
			return errors.New("no pongo context found")
		}
	}
	// check if the templates exists
	var tpl *template.Template
	if tpl, ok = t.templates[name]; !ok {
		log.Printf("template not found: %s\n", name)
		return errors.New("template not found")
	}
	err := tpl.ExecuteTemplate(w, "html", ctx)
	if err != nil {
		log.Printf("error executing template: %s\n", err)
		return err
	}
	return nil
}
