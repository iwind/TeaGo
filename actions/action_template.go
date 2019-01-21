package actions

import (
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/files"
	"github.com/iwind/TeaGo/gohtml"
	"github.com/iwind/TeaGo/gohtml/atom"
	"github.com/iwind/TeaGo/logs"
	"github.com/iwind/TeaGo/maps"
	"github.com/iwind/TeaGo/utils/string"
	"github.com/pquerna/ffjson/ffjson"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"text/template"
	"time"
)

type TemplateCache struct {
	template      *Template
	watchingFiles map[string]int64 // file => modifiedAt
}

var templateCaches = sync.Map{}
var templateCacheTime = time.Now().Unix()
var templateFileStatCache = sync.Map{} // filename => html

// 渲染模板
func (this *ActionObject) render(dir string) error {
	module := this.Module
	filename := this.viewTemplate
	data := this.Data
	viewFuncMap := this.viewFuncMap

	// 去除末尾的.html
	tailReplacer, err := stringutil.RegexpCompile("\\.html")
	if err != nil {
		panic(err)
	}
	filename = tailReplacer.ReplaceAllString(filename, "")

	filename = dir + "/" + filename
	cache, ok := templateCaches.Load(filename)
	if ok {
		// 生产环境直接使用缓存
		if Tea.Env == Tea.EnvProd {
			teaFuncMap := createTeaFuncMap(cache.(*TemplateCache).template, viewFuncMap, module, dir, filename, data)
			t := cache.(*TemplateCache).template.Funcs(teaFuncMap)
			if this.writer != nil {
				return t.Execute(this.writer, data)
			} else {
				return t.Execute(this.ResponseWriter, data)
			}
		}

		var isChanged = false
		for watchingFile, modifiedAt := range cache.(*TemplateCache).watchingFiles {
			stat, err := os.Stat(watchingFile)
			if err != nil {
				return err
			}
			if stat.ModTime().Unix() != modifiedAt {
				isChanged = true
				break
			}
		}

		if !isChanged {
			teaFuncMap := createTeaFuncMap(cache.(*TemplateCache).template, viewFuncMap, module, dir, filename, data)

			t := cache.(*TemplateCache).template.Funcs(teaFuncMap)
			if this.writer != nil {
				return t.Execute(this.writer, data)
			} else {
				return t.Execute(this.ResponseWriter, data)
			}
		}
	}

	watchingFiles := map[string]int64{}

	_, err = os.Stat(filename + ".html")
	if err != nil {
		return err
	}

	bodyBytes, err := ioutil.ReadFile(filename + ".html")

	if err != nil {
		return err
	}
	body := string(bodyBytes)

	// 布局模板
	{
		reg, err := stringutil.RegexpCompile("\\{\\s*\\$(layout|TEA\\.LAYOUT)\\s*\\}")
		if err != nil {
			return err
		}
		hasLayout := false
		body = reg.ReplaceAllStringFunc(body, func(s string) string {
			hasLayout = true
			return ""
		})

		if hasLayout {
			layoutFile := dir + "/@layout.html"
			addFileToWatchingFiles(&watchingFiles, layoutFile)

			_, err := os.Stat(layoutFile)
			if err == nil {
				layoutBytes, err := ioutil.ReadFile(layoutFile)
				if err == nil {
					layoutBody := string(layoutBytes)

					// 支持{$TEA.VIEW}
					reg, err = stringutil.RegexpCompile("\\{\\s*\\$TEA\\s*\\.\\s*VIEW\\s*\\}")
					if err == nil {
						body = reg.ReplaceAllStringFunc(layoutBody, func(s string) string {
							return body
						})
					}
				}
			}
		}
	}

	// 支持{$var "varName"}var value{$end}
	reg, _ := stringutil.RegexpCompile("(?U)\\{\\s*\\$var\\s+\"(\\w+)\"\\s*\\}((.|\n)+){\\s*\\$end\\s*}(\n|$)")
	varMaps := []maps.Map{}
	body = reg.ReplaceAllStringFunc(body, func(s string) string {
		matches := reg.FindStringSubmatch(s)
		varMaps = append(varMaps, maps.Map{
			matches[1]: formatHTML(matches[2]),
		})
		return ""
	})

	// 支持 {$TEA.VUE}
	reg, _ = stringutil.RegexpCompile("\\{\\s*\\$TEA\\s*\\.\\s*VUE\\s*\\}")
	body = reg.ReplaceAllString(body, "{$$TEA_VUE}")

	// 支持 {$TEA.DATA}
	reg, _ = stringutil.RegexpCompile("\\{\\s*\\$TEA\\s*\\.\\s*DATA\\s*\\}")
	body = reg.ReplaceAllString(body, "{$$TEA_DATA}")

	// 支持 {$TEA.SEMANTIC}
	reg, _ = stringutil.RegexpCompile("\\{\\s*\\$TEA\\s*\\.\\s*SEMANTIC\\s*\\}")
	body = reg.ReplaceAllString(body, "{$$TEA_SEMANTIC}")

	// 去掉 {$TEA.VIEW}
	reg, _ = stringutil.RegexpCompile("\\{\\s*\\$TEA\\s*\\.\\s*VIEW\\s*\\}")
	body = reg.ReplaceAllString(body, "{$$TEA_VIEW}")

	// 分析模板
	body = formatHTML(body)

	// 内部自定义函数
	tpl := NewTemplate(filename)
	teaFuncMap := createTeaFuncMap(tpl, viewFuncMap, module, dir, filename, data)
	newTemplate, err := tpl.Delims("{$", "}").Funcs(teaFuncMap).Parse(body)
	if err != nil {
		logs.Errorf("Template parse error:%s", err.Error())
		return err
	}

	for _, varMap := range varMaps {
		newTemplate.SetVars(varMap)
	}

	addFileToWatchingFiles(&watchingFiles, filename+".html")

	// 子模板
	{
		reg, err := stringutil.RegexpCompile("\\{\\$template\\s+\"(.+)\"\\}")
		if err != nil {
			return err
		}
		matches := reg.FindAllStringSubmatch(body, -1)
		for _, match := range matches {
			err = loadChildTemplate(&watchingFiles, newTemplate, dir, filename, match[1])
			if err != nil {
				return err
			}
		}
	}

	cache = &TemplateCache{
		template:      newTemplate,
		watchingFiles: watchingFiles,
	}
	templateCaches.Store(filename, cache)

	if this.writer != nil {
		return newTemplate.ExecuteTemplate(this.writer, filename, data)
	} else {
		return newTemplate.ExecuteTemplate(this.ResponseWriter, filename, data)
	}
}

func loadChildTemplate(watchingFiles *map[string]int64, tpl *Template, dir string, filename string, childTemplateName string) error {
	viewPath := pathRelative(dir, filename, childTemplateName)
	childBytes, err := ioutil.ReadFile(viewPath)
	if err != nil {
		return err
	}
	body := string(childBytes)
	body = formatHTML(body)

	// 支持{$var "varName"}var value{$end}
	reg, _ := stringutil.RegexpCompile("(?U)\\{\\s*\\$var\\s+\"(\\w+)\"\\s*\\}((.|\n)+){\\s*\\$end\\s*}(\n|$)")
	body = reg.ReplaceAllStringFunc(body, func(s string) string {
		matches := reg.FindStringSubmatch(s)
		tpl.SetVars(maps.Map{
			matches[1]: matches[2],
		})
		return ""
	})

	// 子模板
	reg, err = stringutil.RegexpCompile("\\{\\$template\\s+\"(.+)\"\\}")
	matches := reg.FindAllStringSubmatch(body, -1)
	for _, match := range matches {
		err = loadChildTemplate(watchingFiles, tpl, dir, filename, match[1])
		if err != nil {
			return err
		}
	}

	_, err = tpl.NewChild(childTemplateName).Delims("{$", "}").Parse(body)
	if err != nil {
		logs.Errorf("Template parse error:%s", err.Error())
		return err
	}
	addFileToWatchingFiles(watchingFiles, viewPath)
	return nil
}

func addFileToWatchingFiles(watchingFiles *map[string]int64, filename string) {
	stat, err := os.Stat(filename)
	if err != nil {
		(*watchingFiles)[filename] = 0
	} else {
		(*watchingFiles)[filename] = stat.ModTime().Unix()
	}
}

func pathRelative(dir string, filename string, path string) string {
	childDir := filepath.Dir(path)
	childName := filepath.Base(path)

	if childDir == "." {
		return filepath.Dir(filename) + "/@" + childName + ".html"
	} else if childDir == ".." {
		return filepath.Dir(filepath.Dir(filename)) + "/@" + childName + ".html"
	} else if strings.HasPrefix(childDir, "/") {
		return dir + childDir + "/@" + childName + ".html"
	} else {
		return dir + "/" + childDir + "/@" + childName + ".html"
	}
	return path + ".html"
}

func createTeaFuncMap(tpl *Template, funcMap template.FuncMap, module string, dir string, filename string, data map[string]interface{}) template.FuncMap {
	parent := filepath.Dir(strings.TrimPrefix(filename, dir))
	if runtime.GOOS == "windows" {
		parent = strings.Replace(parent, "\\", "/", -1)
	}
	if module == "default" {
		module = ""
	}

	actionData := map[string]interface{}{
		"data":        data,
		"base":        "",
		"module":      module,
		"parent":      parent,
		"actionParam": false,
	}
	dataBytes, err := ffjson.Marshal(actionData)
	jsonString := ""
	if err != nil {
		logs.Errorf("%s", err.Error())
		jsonString = "null"
	} else {
		jsonString = string(dataBytes)
	}

	funcMap["TEA_DATA"] = func() string {
		var teaHtml = `<script type="text/javascript">
window.TEA = {
	"ACTION": ` + jsonString + `
}
</script>`
		return string(teaHtml)
	}
	funcMap["TEA_VUE"] = func() string {
		var teaHtml = `<script type="text/javascript">
window.TEA = {
	"ACTION": ` + jsonString + `
}
</script>` + "\n"

		includeHTML, found := templateFileStatCache.Load(filename)
		if Tea.Env == Tea.EnvProd && found {
			teaHtml += includeHTML.(string)
		} else {
			pieces := []string{}

			{
				if Tea.Env == Tea.EnvProd {
					jsFile := "js/vue.min.js"
					pieces = append(pieces, "<script type=\"text/javascript\" src=\"/"+jsFile+"?v="+stringutil.ConvertID(templateCacheTime)+"\"></script>")
				} else {
					jsFile := "js/vue.js"
					stat, err := os.Stat(Tea.PublicFile(jsFile))
					if err == nil {
						pieces = append(pieces, "<script type=\"text/javascript\" src=\"/"+jsFile+"?v="+stringutil.ConvertID(stat.ModTime().Unix())+"\"></script>")
					} else {
						pieces = append(pieces, "<!-- warning: "+jsFile+" not appeared in public/ -->")
					}
				}
			}

			{
				jsFile := "js/vue.tea.js"
				if Tea.Env == Tea.EnvProd {
					pieces = append(pieces, "<script type=\"text/javascript\" src=\"/"+jsFile+"?v="+stringutil.ConvertID(templateCacheTime)+"\"></script>")
				} else {
					stat, err := os.Stat(Tea.PublicFile(jsFile))
					if err == nil {
						pieces = append(pieces, "<script type=\"text/javascript\" src=\"/"+jsFile+"?v="+stringutil.ConvertID(stat.ModTime().Unix())+"\"></script>")
					} else {
						pieces = append(pieces, "<!-- warning: "+jsFile+" not appeared in public/ -->")
					}
				}
			}

			{
				jsFile := filename + ".js"
				if Tea.Env == Tea.EnvProd {
					stat, err := os.Stat(jsFile)
					if err == nil {
						pieces = append(pieces, "<script type=\"text/javascript\" src=\"/_/"+strings.TrimPrefix(jsFile, Tea.ViewsDir()+"/")+"?v="+stringutil.ConvertID(stat.ModTime().Unix())+"\"></script>")
					}
				} else {
					stat, err := os.Stat(jsFile)
					if err == nil {
						pieces = append(pieces, "<script type=\"text/javascript\" src=\"/_/"+strings.TrimPrefix(jsFile, Tea.ViewsDir()+"/")+"?v="+stringutil.ConvertID(stat.ModTime().Unix())+"\"></script>")
					} else {
						pieces = append(pieces, "<!-- warning: "+strings.TrimPrefix(jsFile, Tea.ViewsDir()+"/")+" not appeared in views/ -->")
					}
				}
			}

			{
				cssFile := filename + ".css"
				if Tea.Env == Tea.EnvProd {
					stat, err := os.Stat(cssFile)
					if err == nil {
						pieces = append(pieces, "<link rel=\"stylesheet\" type=\"text/css\" href=\"/_/"+strings.TrimPrefix(cssFile, Tea.ViewsDir()+"/")+"?v="+stringutil.ConvertID(stat.ModTime().Unix())+"\" media=\"all\"/>")
					}
				} else {
					stat, err := os.Stat(cssFile)
					if err == nil {
						pieces = append(pieces, "<link rel=\"stylesheet\" type=\"text/css\" href=\"/_/"+strings.TrimPrefix(cssFile, Tea.ViewsDir()+"/")+"?v="+stringutil.ConvertID(stat.ModTime().Unix())+"\" media=\"all\"/>")
					} else {
						pieces = append(pieces, "<!-- warning: "+strings.TrimPrefix(cssFile, Tea.ViewsDir()+"/")+" not appeared in views/ -->")
					}
				}
			}

			includeHTML := strings.Join(pieces, "\n")

			if Tea.Env == Tea.EnvProd {
				templateFileStatCache.Store(filename, includeHTML)
			}

			teaHtml += includeHTML
		}

		return teaHtml
	}

	funcMap["TEA_VIEW"] = func() string {
		return ""
	}

	funcMap["TEA_SEMANTIC"] = func() string {
		cssFile := Tea.PublicFile("css/semantic.min.css")
		if Tea.Env == Tea.EnvProd {
			includeHTML, found := templateFileStatCache.Load(filename + "_TEA_SEMANTIC")
			if found {
				return includeHTML.(string)
			} else {
				stat, err := os.Stat(cssFile)
				if err == nil {
					s := "<link rel=\"stylesheet\" type=\"text/css\" href=\"/css/semantic.min.css?v=" + stringutil.ConvertID(stat.ModTime().Unix()) + "\" media=\"all\"/>"
					templateFileStatCache.Store(filename+"_TEA_SEMANTIC", s)
					return s
				} else {
					return "<!-- warning: css/semantic.min.css not appeared in public/ -->"
				}
			}
		} else {
			stat, err := os.Stat(cssFile)
			if err == nil {
				return "<link rel=\"stylesheet\" type=\"text/css\" href=\"/css/semantic.min.css?v=" + stringutil.ConvertID(stat.ModTime().Unix()) + "\" media=\"all\"/>"
			}
		}

		return "<!-- warning: css/semantic.min.css not appeared in public/ -->"
	}

	funcMap["echo"] = func(s string) string {
		return tpl.VarValue(s)
	}

	funcMap["hasVar"] = func(s string) bool {
		return tpl.HasVar(s)
	}

	return funcMap
}

func formatHTML(htmlString string) string {
	reader := strings.NewReader(htmlString)
	tokenizer := gohtml.NewTokenizer(reader)
	result := ""
	hasDocType := false
	for {
		if tokenizer.Err() != nil {
			break
		}
		tokenType := tokenizer.Next()
		if tokenType != gohtml.StartTagToken && tokenType != gohtml.SelfClosingTagToken {
			result += string(tokenizer.Raw())

			if tokenType == gohtml.DoctypeToken {
				hasDocType = true
			}
			continue
		}

		token := tokenizer.Token()
		tagType := token.DataAtom

		// 自动增加 doctype
		if tagType == atom.Html && !hasDocType {
			result += "<!DOCTYPE html>\n"
		}

		tagHTML := string(tokenizer.Raw())
		if tagType == atom.Img {
			for _, attr := range token.Attr {
				if attr.Key == "src" {
					pattern := attr.Val
					pattern = strings.Replace(pattern, "?", "\\?", -1)
					reg, err := regexp.Compile("(['\"])" + pattern + "(['\"])")
					if err == nil {
						tagHTML = reg.ReplaceAllString(tagHTML, "${1}"+getResourceVersion(attr.Val)+"${2}")
					}
					break
				}
			}
		} else if tagType == atom.Script {
			for _, attr := range token.Attr {
				if attr.Key == "src" {
					pattern := attr.Val
					pattern = strings.Replace(pattern, "?", "\\?", -1)
					reg, err := regexp.Compile("(['\"])" + pattern + "(['\"])")
					if err == nil {
						tagHTML = reg.ReplaceAllString(tagHTML, "${1}"+getResourceVersion(attr.Val)+"${2}")
					}
					break
				}
			}
		} else if tagType == atom.Link {
			for _, attr := range token.Attr {
				if attr.Key == "href" {
					pattern := attr.Val
					pattern = strings.Replace(pattern, "?", "\\?", -1)
					reg, err := regexp.Compile("(['\"])" + pattern + "(['\"])")
					if err == nil {
						tagHTML = reg.ReplaceAllString(tagHTML, "${1}"+getResourceVersion(attr.Val)+"${2}")
					}
					break
				}
			}
		} else if tagType == atom.Source {
			for _, attr := range token.Attr {
				if attr.Key == "src" {
					pattern := attr.Val
					pattern = strings.Replace(pattern, "?", "\\?", -1)
					reg, err := regexp.Compile("(['\"])" + pattern + "(['\"])")
					if err == nil {
						tagHTML = reg.ReplaceAllString(tagHTML, "${1}"+getResourceVersion(attr.Val)+"${2}")
					}
					break
				}
			}
		}

		result += tagHTML
	}

	return result
}

func getResourceVersion(resourceURL string) string {
	if !strings.HasPrefix(resourceURL, "/") {
		return resourceURL
	}

	uri, err := url.ParseRequestURI(resourceURL)
	if err != nil {
		return resourceURL
	}

	query := uri.Query()

	var file *files.File
	if strings.HasPrefix(uri.Path, "/_/") {
		file = files.NewFile(Tea.ViewsDir() + uri.Path[2:])
	} else {
		file = files.NewFile(Tea.PublicFile(uri.Path))
	}
	if !file.Exists() || !file.IsFile() {
		return resourceURL
	}

	modifiedTime, err := file.LastModified()
	if err != nil {
		return resourceURL
	}

	version := stringutil.ConvertID(modifiedTime.Unix())
	if len(query) == 0 {
		resourceURL = resourceURL + "?v=" + version
	} else {
		_, found := query["v"]
		if !found {
			resourceURL += "&v=" + version
		} else {
			resourceURL += "&rv=" + version
		}
	}

	return resourceURL
}
