package template

import (
	"errors"
	"io"
	"log"
	"os"

	"github.com/flosch/pongo2/v6"
	"github.com/labstack/echo/v4"
)

type Template struct {
	templates map[string]*pongo2.Template
}

func New() *Template {
	templates := make(map[string]*pongo2.Template)
	// check public/views directory for templates
	files, err := os.ReadDir("public/views")
	if err != nil {
		// create the directory if it doesn't exist, then exit
		if os.IsNotExist(err) {
			os.Mkdir("public/views", 0755)
			log.Println("public/views directory created")
			os.Exit(0)
		} else {
			panic(err)
		}
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		// load the template
		tpl, err := pongo2.FromFile("public/views/" + file.Name())
		if err != nil {
			panic(err)
		}
		// add the template to the map
		templates[file.Name()] = tpl
	}
	for k := range templates {
		log.Println("template loaded: ", k)
	}
	return &Template{
		templates: templates,
	}
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	var ctx pongo2.Context
	var ok bool
	c.Logger().Info("Rendering template: ", name)
	if data != nil {
		ctx, ok = data.(pongo2.Context)
		if !ok {
			return errors.New("no pongo context found")
		}
	}
	// check if the template exists
	tpl, ok := t.templates[name]
	if !ok {
		return errors.New("template not found")
	}
	// render the template
	err := tpl.ExecuteWriter(ctx, w)
	if err != nil {
		return err
	}
	return nil
}
