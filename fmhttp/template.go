package fmhttp

import (
	"html/template"
	"io"
	"io/fs"
	"strings"
	"sync"
	"text/template/parse"
)

type TemplateRender struct {
	DevMode    bool
	TemplateFS fs.FS

	cachedTemplatesMu sync.RWMutex
	cachedTemplates   map[string]*template.Template
}

func NewTemplateRender(templateFS fs.FS) *TemplateRender {
	return &TemplateRender{
		TemplateFS:      templateFS,
		cachedTemplates: map[string]*template.Template{},
	}
}

func (r *TemplateRender) parseTemplateOne(name string) (*template.Template, error) {
	f, err := r.TemplateFS.Open(name)
	if err != nil {
		return nil, err
	}
	tmplBytes, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	t := template.New(name)
	_, err = t.Parse(string(tmplBytes))
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (r *TemplateRender) parseTemplate(parsedMap map[string]*template.Template, name string) error {
	tmpl, err := r.parseTemplateOne(name)
	if err != nil {
		return err
	}

	parsedMap[name] = tmpl
	for _, node := range tmpl.Tree.Root.Nodes {
		if tmplNode, ok := node.(*parse.TemplateNode); ok {
			if _, ok := parsedMap[tmplNode.Name]; ok {
				continue
			}
			if !strings.HasSuffix(tmplNode.Name, ".tmpl") {
				continue
			}
			err = r.parseTemplate(parsedMap, tmplNode.Name)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *TemplateRender) loadTemplate(tmplName string) (*template.Template, error) {
	if !r.DevMode {
		r.cachedTemplatesMu.RLock()
		t, ok := r.cachedTemplates[tmplName]
		r.cachedTemplatesMu.RUnlock()
		if ok {
			return t, nil
		}
	}
	r.cachedTemplatesMu.Lock()
	defer r.cachedTemplatesMu.Unlock()

	t, ok := r.cachedTemplates[tmplName]
	if r.DevMode || !ok {
		parsedMap := map[string]*template.Template{}
		err := r.parseTemplate(parsedMap, tmplName)
		if err != nil {
			return nil, err
		}

		t = parsedMap[tmplName]
		for parsedName, parsedTmpl := range parsedMap {
			_, err = t.AddParseTree(parsedName, parsedTmpl.Tree)
			if err != nil {
				return nil, err
			}
		}

		r.cachedTemplates[tmplName] = t
	}
	return t, nil
}

func (r *TemplateRender) Render(wr io.Writer, tmplName string, data map[string]any) error {
	t, err := r.loadTemplate(tmplName)
	if err != nil {
		return err
	}
	return t.Execute(wr, data)
}
