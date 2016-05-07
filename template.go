// Copyright 2016 goweb Author. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package goweb

import (
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/cooleo/goweb/utils"
)

var (
	gowebTplFuncMap = make(template.FuncMap)
	// beeTemplates caching map and supported template file extensions.
	beeTemplates  = make(map[string]*template.Template)
	templatesLock sync.RWMutex
	// beeTemplateExt stores the template extension which will build
	beeTemplateExt = []string{"tpl", "html"}
)

func executeTemplate(wr io.Writer, name string, data interface{}) error {
	if BConfig.RunMode == DEV {
		templatesLock.RLock()
		defer templatesLock.RUnlock()
	}
	if t, ok := beeTemplates[name]; ok {
		err := t.ExecuteTemplate(wr, name, data)
		if err != nil {
			Trace("template Execute err:", err)
		}
		return err
	}
	panic("can't find templatefile in the path:" + name)
}

func init() {
	gowebTplFuncMap["dateformat"] = DateFormat
	gowebTplFuncMap["date"] = Date
	gowebTplFuncMap["compare"] = Compare
	gowebTplFuncMap["compare_not"] = CompareNot
	gowebTplFuncMap["not_nil"] = NotNil
	gowebTplFuncMap["not_null"] = NotNil
	gowebTplFuncMap["substr"] = Substr
	gowebTplFuncMap["html2str"] = HTML2str
	gowebTplFuncMap["str2html"] = Str2html
	gowebTplFuncMap["htmlquote"] = Htmlquote
	gowebTplFuncMap["htmlunquote"] = Htmlunquote
	gowebTplFuncMap["renderform"] = RenderForm
	gowebTplFuncMap["assets_js"] = AssetsJs
	gowebTplFuncMap["assets_css"] = AssetsCSS
	gowebTplFuncMap["config"] = GetConfig
	gowebTplFuncMap["map_get"] = MapGet

	// Comparisons
	gowebTplFuncMap["eq"] = eq // ==
	gowebTplFuncMap["ge"] = ge // >=
	gowebTplFuncMap["gt"] = gt // >
	gowebTplFuncMap["le"] = le // <=
	gowebTplFuncMap["lt"] = lt // <
	gowebTplFuncMap["ne"] = ne // !=

	gowebTplFuncMap["urlfor"] = URLFor // build a URL to match a Controller and it's method
}

// AddFuncMap let user to register a func in the template.
func AddFuncMap(key string, fn interface{}) error {
	gowebTplFuncMap[key] = fn
	return nil
}

type templateFile struct {
	root  string
	files map[string][]string
}

// visit will make the paths into two part,the first is subDir (without tf.root),the second is full path(without tf.root).
// if tf.root="views" and
// paths is "views/errors/404.html",the subDir will be "errors",the file will be "errors/404.html"
// paths is "views/admin/errors/404.html",the subDir will be "admin/errors",the file will be "admin/errors/404.html"
func (tf *templateFile) visit(paths string, f os.FileInfo, err error) error {
	if f == nil {
		return err
	}
	if f.IsDir() || (f.Mode()&os.ModeSymlink) > 0 {
		return nil
	}
	if !HasTemplateExt(paths) {
		return nil
	}

	replace := strings.NewReplacer("\\", "/")
	file := strings.TrimLeft(replace.Replace(paths[len(tf.root):]), "/")
	subDir := filepath.Dir(file)

	tf.files[subDir] = append(tf.files[subDir], file)
	return nil
}

// HasTemplateExt return this path contains supported template extension of goweb or not.
func HasTemplateExt(paths string) bool {
	for _, v := range beeTemplateExt {
		if strings.HasSuffix(paths, "."+v) {
			return true
		}
	}
	return false
}

// AddTemplateExt add new extension for template.
func AddTemplateExt(ext string) {
	for _, v := range beeTemplateExt {
		if v == ext {
			return
		}
	}
	beeTemplateExt = append(beeTemplateExt, ext)
}

// BuildTemplate will build all template files in a directory.
// it makes goweb can render any template file in view directory.
func BuildTemplate(dir string, files ...string) error {
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return errors.New("dir open err")
	}
	self := &templateFile{
		root:  dir,
		files: make(map[string][]string),
	}
	err := filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
		return self.visit(path, f, err)
	})
	if err != nil {
		fmt.Printf("filepath.Walk() returned %v\n", err)
		return err
	}
	for _, v := range self.files {
		for _, file := range v {
			if len(files) == 0 || utils.InSlice(file, files) {
				templatesLock.Lock()
				t, err := getTemplate(self.root, file, v...)
				if err != nil {
					Trace("parse template err:", file, err)
				} else {
					beeTemplates[file] = t
				}
				templatesLock.Unlock()
			}
		}
	}
	return nil
}

func getTplDeep(root, file, parent string, t *template.Template) (*template.Template, [][]string, error) {
	var fileAbsPath string
	if filepath.HasPrefix(file, "../") {
		fileAbsPath = filepath.Join(root, filepath.Dir(parent), file)
	} else {
		fileAbsPath = filepath.Join(root, file)
	}
	if e := utils.FileExists(fileAbsPath); !e {
		panic("can't find template file:" + file)
	}
	data, err := ioutil.ReadFile(fileAbsPath)
	if err != nil {
		return nil, [][]string{}, err
	}
	t, err = t.New(file).Parse(string(data))
	if err != nil {
		return nil, [][]string{}, err
	}
	reg := regexp.MustCompile(BConfig.WebConfig.TemplateLeft + "[ ]*template[ ]+\"([^\"]+)\"")
	allSub := reg.FindAllStringSubmatch(string(data), -1)
	for _, m := range allSub {
		if len(m) == 2 {
			tl := t.Lookup(m[1])
			if tl != nil {
				continue
			}
			if !HasTemplateExt(m[1]) {
				continue
			}
			t, _, err = getTplDeep(root, m[1], file, t)
			if err != nil {
				return nil, [][]string{}, err
			}
		}
	}
	return t, allSub, nil
}

func getTemplate(root, file string, others ...string) (t *template.Template, err error) {
	t = template.New(file).Delims(BConfig.WebConfig.TemplateLeft, BConfig.WebConfig.TemplateRight).Funcs(gowebTplFuncMap)
	var subMods [][]string
	t, subMods, err = getTplDeep(root, file, "", t)
	if err != nil {
		return nil, err
	}
	t, err = _getTemplate(t, root, subMods, others...)

	if err != nil {
		return nil, err
	}
	return
}

func _getTemplate(t0 *template.Template, root string, subMods [][]string, others ...string) (t *template.Template, err error) {
	t = t0
	for _, m := range subMods {
		if len(m) == 2 {
			tpl := t.Lookup(m[1])
			if tpl != nil {
				continue
			}
			//first check filename
			for _, otherFile := range others {
				if otherFile == m[1] {
					var subMods1 [][]string
					t, subMods1, err = getTplDeep(root, otherFile, "", t)
					if err != nil {
						Trace("template parse file err:", err)
					} else if subMods1 != nil && len(subMods1) > 0 {
						t, err = _getTemplate(t, root, subMods1, others...)
					}
					break
				}
			}
			//second check define
			for _, otherFile := range others {
				fileAbsPath := filepath.Join(root, otherFile)
				data, err := ioutil.ReadFile(fileAbsPath)
				if err != nil {
					continue
				}
				reg := regexp.MustCompile(BConfig.WebConfig.TemplateLeft + "[ ]*define[ ]+\"([^\"]+)\"")
				allSub := reg.FindAllStringSubmatch(string(data), -1)
				for _, sub := range allSub {
					if len(sub) == 2 && sub[1] == m[1] {
						var subMods1 [][]string
						t, subMods1, err = getTplDeep(root, otherFile, "", t)
						if err != nil {
							Trace("template parse file err:", err)
						} else if subMods1 != nil && len(subMods1) > 0 {
							t, err = _getTemplate(t, root, subMods1, others...)
						}
						break
					}
				}
			}
		}

	}
	return
}

// SetViewsPath sets view directory path in goweb application.
func SetViewsPath(path string) *App {
	BConfig.WebConfig.ViewsPath = path
	return BeeApp
}

// SetStaticPath sets static directory path and proper url pattern in goweb application.
// if goweb.SetStaticPath("static","public"), visit /static/* to load static file in folder "public".
func SetStaticPath(url string, path string) *App {
	if !strings.HasPrefix(url, "/") {
		url = "/" + url
	}
	if url != "/" {
		url = strings.TrimRight(url, "/")
	}
	BConfig.WebConfig.StaticDir[url] = path
	return BeeApp
}

// DelStaticPath removes the static folder setting in this url pattern in goweb application.
func DelStaticPath(url string) *App {
	if !strings.HasPrefix(url, "/") {
		url = "/" + url
	}
	if url != "/" {
		url = strings.TrimRight(url, "/")
	}
	delete(BConfig.WebConfig.StaticDir, url)
	return BeeApp
}
