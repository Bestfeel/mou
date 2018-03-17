package markdown

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	md "github.com/russross/blackfriday"
	"fmt"
	"encoding/base64"
	"github.com/elazarl/go-bindata-assetfs"
	"github.com/Bestfeel/mou/asset"
	"gopkg.in/russross/blackfriday.v2"
	"path"
)

const tpl = `
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="utf-8">
    <meta http-equiv="Cache-Control" content="no-store" />
	<meta http-equiv="cache-control" content="max-age=0" />
	<meta http-equiv="cache-control" content="no-cache" />
	<meta http-equiv="expires" content="0" />
	<meta http-equiv="pragma" content="no-cache" />
    <meta http-equiv="x-ua-compatible" content="ie=edge">
    <title>mou</title>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link rel="stylesheet" href="/static/styles/bulma.min.css">
    <link rel="stylesheet" href="/static/styles/font-awesome.min.css">
    <link rel="stylesheet" href="/static/styles/prism.css">
    <link rel="stylesheet" href="/static/styles/echo.css">
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/9.12.0/styles/monokai.min.css">
    <script src="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/9.12.0/highlight.min.js"></script>
    <script>hljs.initHighlightingOnLoad();</script>

</head>
<body>
<section class="hero hero-echo">
    <div class="hero-head">
        <header class="nav">
            <div class="container">
                <div class="nav-left">
                    <span class="nav-item title" style="color:#fff;font-weight:bold;">{{.name}}
                    </span>
                </div>
                <div class="nav-right nav-menu">
                    <div class="nav-item">
                    </div>
                    <span class="nav-item">
                  </span>
                </div>
                <span class="nav-toggle"></span>
            </div>
        </header>
    </div>
</section>
<div class="container">
    <div class="columns">
        <div id="search-menu" class="column is-half is-offset-one-quarter"></div>
    </div>
</div>

<section style="margin-top: 40px;">
    <div class="container">
        <div class="columns">
            <div class="column is-3">
                <aside class="menu">
                    <ul class="menu-list">
                    {{range .files}}
                        <li>
                            <a href="{{.Path}}">
                            {{ .Name }}
                            </a>
                        </li>
                    {{end}}
                    </ul>
                </aside>
            </div>
            <div class="column is-9">
                <div class="content">
                {{.body}}
                </div>
            </div>
        </div>
    </div>
</section>
<script src="/static/scripts/prism.js"></script>
//<script src="/static/scripts/echo.js"></script>
</body>
</html>
`
const LOGO = `
                                                  _
                                          _______| |
                                         |_________|
                                          _________
                                         |  _______|   万物互联
                                         | |   ____
                                         | |  |__  |   机智云
                                         | |_____| |
                                         |_________|   Gizwits
                                          机智云只为硬件而生的云服务


`

const ICO = "data:image/vnd.microsoft.icon;base64,AAABAAEAICAAAAEAIACoEAAAFgAAACgAAAAgAAAAQAAAAAEAIAAAAAAAABAAACUWAAAlFgAAAAAAAAAAAABA5/L/QOfy/0Dn8v9A5/L/QOfy/0Dn8v9A5/H/QObx/0Dm8f9A5vH/QObx/0Dm8f9A5vH/QObx/0Dm8f9A5vH/QObx/0Dm8f9A5vH/QObx/0Dm8f9A5vH/QObx/0Dm8f9A5vH/QObx/0Dn8v9A5/L/QOfy/0Dn8v9A5/L/QOfy/0Dn8v9A5/L/QOfy/0Dn8v9A5/L/QObw/0Dn8P9B6/X/Qev1/0Hr9f9B6/X/Qev1/0Hr9f9B6/X/Qev1/0Hr9f9B6/X/Qev1/0Hr9f9B6/X/Qev1/0Hs9P9B7PT/Qez0/0Hs8/9A5/D/P+Xv/0Ho8v9A5/L/QOfy/0Dn8v9A5/L/QOfy/0Dn8v9A5/L/QOfy/0Dm8f9D7fb/Re/4/0Hb4/9B2eH/QNrh/0DZ4f9A2eH/QNnh/0DZ4f9A2eH/QNnh/0DZ4f9A2eH/QNnh/0DZ4f9A2uH/QNrh/0Da4f9C2uL/P9ng/0Pr8f9E7/z/QObx/0Do8f9A5/L/QOfy/0Dn8v9A5/L/QOfy/0Dn8v9A5vH/Qev2/0HY3f8jWFz/GScq/xgpK/8XKCr/GCgq/xkoKv8ZKCr/GSgq/xkoKv8ZKCr/GSgq/xkoKv8ZKCr/GSgq/xcoKv8XKCr/Fygq/xgpK/8aJyn/H0JE/z/Cyv9D7vr/QObv/0Dn8v9A5/L/QOfy/0Dn8v9A5/L/QOfy/z/k7/9F9P//LG52/xUEBP8cGx3/HBkZ/x0cG/8cHBz/HRwc/x0cHP8dHBz/HRwc/x0cHP8dHBz/HRwc/x0cHP8dHBz/HB0c/xsdHP8aHRz/HBoZ/x4bG/8YCQj/HUNG/0Tp9v9A5/L/QOfy/0Dn8v9A5/L/QOfy/0Dn8v9A5/L/P+Xw/0jx/f8pW17/FhIS/x8fIv8dGxv/Fw8P/xoODv8ZDg7/GQ4O/xoOD/8aDg7/GQ4O/xkOD/8ZDg7/GQ4O/xoOD/8ZDw//GA8P/xgNDf8bGBj/HB8f/x0aGv8ZLS//QN3n/0Hq9f9A5vH/QOfy/0Dn8v9A5/L/QOfy/0Dn8v8/5fD/R/H9/ypfX/8ZDxH/Gh4c/xojIf8ndHj/K36D/yx7gf8sfIH/K3uB/yt8gf8sfIH/K3yB/yx8gf8sfIH/K3yB/yt7gf8qe4D/LH2D/yE0OP8aGRj/HhgZ/xsxMv8/3+f/Qen1/0Dm8f9A5/L/QOfy/0Dn8v9A5/L/QOfy/z/l8P9H8f3/Kl1h/xgQEf8dGBn/HDg7/0fw+P9F9///RPT+/0Ty/v9E8/7/RPL+/0Tz/v9E8v7/RPL+/0Tz/v9F8v7/RfL+/0Pw+/9K////LGtv/xoOD/8eGxv/Gi8z/z/e6f9B6vX/QObx/0Dn8v9A5/L/QOfy/0Dn8v9A5/L/P+Xw/0fx/f8qXmD/GRAR/xsZGf8YODv/P+Hp/z/k7v9B6fT/RPP9/0Py/v9D8v7/Q/L+/0Py/v9D8v7/Q/L+/0Py/v9D8/7/Qu/7/0n///8xbW//GA4P/xwbGv8aLzL/P97p/0Hq9f9A5vH/QOfy/0Dn8v9A5/L/QOfy/0Dn8v8/5fD/R/H9/ypeYP8ZEBH/GxoX/xw4Ov9A4+r/Q/D5/ze9yf8rg4j/L4uQ/y6Jjv8uiY7/LomO/y6Jjv8uiY7/LomO/y6Jjv8tiY7/MI2Q/yE5PP8aGBf/HBoW/xowNP8/3un/Qer1/0Dm8f9A5/L/QOfy/0Dn8v9A5/L/QOfy/z/l8P9H8f3/Kl1g/xgQEf8aGhf/HDk6/0Dh6f9F9v//Lo2T/xICAP8aEBL/GQ4P/xkOD/8ZDg//GQ4P/xkOD/8ZDg//GQ4P/xkOD/8ZDhD/GhYZ/x4eH/8fGhr/Gy4x/z7c5v9B6vX/QObx/0Dn8v9A5/L/QOfy/0Dn8v9A5/L/P+Xw/0fx/f8qXmD/GBAS/xoaF/8dODr/QODq/0T0/v8uk5n/GBAR/yAeIf8eGx7/Hhse/x4bHv8eGx7/Hhse/x4bHv8eGx7/Hxse/x4bHv8bGhv/Hxse/xcMC/8dPUH/Qejx/0Dn8v9A5/L/QOfy/0Dn8v9A5/L/QOfy/0Dn8v8/5fD/R/H9/ypeYP8ZEBL/GhoX/xw4Ov9A4er/RfT+/y+YnP8VGxz/Higp/x0mJ/8dJif/HSYn/x0mJ/8dJif/HSYn/x0mJ/8dJif/HCYm/x4mJ/8bJSX/HDU5/z65wv9E8Pn/P+Xv/0Dn8v9A5/L/QOfy/0Dn8v9A5/L/QOfy/z/l8P9H8f3/Kl1h/xkQEf8bGhf/Gzg5/0Hk6v9C6fP/Q+Do/0PV3P9D1uD/Q9bf/0PW3/9D1t//Qtbf/0LW3/9D1t//Q9bf/0PW3v9C1t7/Q9ff/0LU3P9E5Oz/RPH3/0Do7f9A5/H/QOfy/0Dn8v9A5/L/QOfy/0Dn8v9A5/L/P+Xw/0fx/f8qXmD/GRAR/xwZF/8ZOjn/QOXv/0Dp8v9A6vL/Qe32/0Hu9f9A7vT/QO70/0Du9f9A7vT/QO70/0Du9f9A7vT/QO30/0Hu9f9A7vT/Qe71/0Dp8v8/5vD/Qejw/0Dn8v9A5/L/QOfy/0Dn8v9A5/L/QOfy/0Dn8v8/5fD/RvL9/ypdYf8ZEBH/GRsZ/xoyM/9D3OT/R+ny/0Xm7f9F5u3/ReXt/0Xm7f9F5u3/ReXt/0Xm7f9F5u3/ReXt/0Xm7f9F5u3/ReXt/0Xm7f9F5u3/Rubt/0bn7/9C5/L/QOfz/0Dn8v9A5/L/QOfy/0Dn8v9A5/L/QOfy/z/l8P9H8f3/KV9h/xgPEf8dHx7/GRwa/x40N/8hPkD/ITw//yE8QP8hPD//IT1A/yE9QP8hPD//IT1A/yE9QP8hPD//IT1A/yE9QP8hPD//ITw//yA9Qf8iOTr/I0xP/0Di6f9A6fX/QObx/0Dn8v9A5/L/QOfy/0Dn8v9A5/L/P+Tw/0fz/v8qW1v/HA8Q/xwiJP8dHx//GxkZ/xoXF/8bGBf/GxgX/xsYF/8bFxf/GxgX/xsYF/8bFxf/GhcX/xsYF/8bFxf/GhcX/xsYF/8aGBf/GxkY/xoTE/8ZKy//P93p/0Hq9f9A5vH/QOfy/0Dn8v9A5/L/QOfy/0Dn8v8/5O//Q/L+/zaUnP8UCgz/Gw8S/xoQEP8bERH/GxER/xsREf8bERH/GxER/xsREf8bERH/GxER/xsREf8bERH/GxER/xsREf8bERH/GxER/xsREf8bEhL/GgwM/xglKf9A3ej/Qer1/0Dm8f9A5/L/QOfy/0Dn8v9A5/L/QOfy/0Dn8v9A5vL/Ru72/ziqrf8pb3P/KG1w/yhtcf8obXH/KG1x/yhtcf8obXD/KG1w/yhtcf8obXH/KG1x/yhtcf8obXD/KG1w/yhtcf8obXH/KG1x/yhtcv8oa23/J3l//0Di7f9A5/T/QOfx/0Dn8v9A5/L/QOfy/0Dn8v9A5/L/QOfy/0Hn8v9A5+7/Q+/4/0Px+/9B7/n/Qe/5/0Hv+f9B7/n/Qe/5/0Hv+f9B7/n/Qe/5/0Hv+f9B7/n/Qe/5/0Hv+f9B7/n/Qe/5/0Hv+f9C7/n/Qe/5/0Lw+v9E8vv/Qunw/0Dn8v9A5/L/QOfy/0Dn8v9A5/L/QOfy/0Dn8v9A5/L/P+by/0Lx+v9F9P3/RPL9/0Xz/f9F8/3/RfP9/0Xz/f9F8/3/RfP9/0Xz/f9F8/3/RfP9/0Xz/f9F8/3/RfP9/0Xz/f9F8/3/RfP9/0Tz/f9F9P3/Q+76/0Dk7/9A6e7/QOjx/0Dn8v9A5/L/QOfy/0Dn8v9A5/L/QOfy/0Dm8f9C6vX/OrK5/y+Wm/8xmqP/MJqh/zCaof8wmqH/MJqh/zCaof8wmqH/MJqh/zCaof8wmqH/MJqh/zCaof8wmqH/MJqh/zCaof8wmqH/MZqi/y6Xnf87u8P/RvH6/0Dm8/9A5vP/QOfy/0Dn8v9A5/L/QOfy/0Dn8v9A5/L/P+Xw/0jy/P8pUln/EQAD/xgQEv8ZDQ7/GA4O/xgNDv8YDQ7/GA0O/xgNDv8YDQ7/GA0O/xgNDv8YDQ7/GA0O/xgODv8YDg7/GA4O/xcODv8XDg//Fw4O/xYSE/8zj5X/RPL7/z/k8P9A5/L/QOfy/0Dn8v9A5/L/QOfy/0Dn8v8/5fD/RvL8/yxgY/8aExT/IiIj/yEfIP8gHyD/IB8g/yAfIP8gHyD/IB8g/yAfIP8gHyD/IB8g/yAfIP8gHyD/IB8g/x8gIP8fIB//HiEf/xohH/8gICH/HRYU/xovNP9B4er/QOnz/0Dm8f9A5/L/QOfy/0Dn8v9A5/L/QOfy/z/l8P9G8fz/Jlxc/xULC/8WGxv/FBca/xQXGv8UFxr/FBca/xQXGv8UFxr/FBca/xQXGv8UFxr/FBca/xQXGv8UFxn/FBgZ/xUYGf8VFRj/Fxgb/x4eHf8aGBn/GS8y/z3c5/9B6fX/QObx/0Dn8v9A5/L/QOfy/0Dn8v9A5/L/QOfy/0Ho8/9AytT/PL/E/zvCyP88wMr/O8DJ/zvAyf87wMn/O8DJ/zvAyf87wMn/O8DJ/zvAyf87wMn/O8DJ/zvByf87wMj/O7/G/z/Hzv8qUVf/HRYW/x0bHv8aMTb/QN7q/0Hq9f9A5vH/QOfy/0Dn8v9A5/L/QOfy/0Dn8v9A5/L/QOfx/0Ht+P9E8Pv/Q+/6/0Lv+/9D7/v/Q+/7/0Pv+/9D7/v/Q+/7/0Pv+/9D7/v/Q+/7/0Pv+/9D7/v/Q+/8/0Pv+/9B7Pj/Sv///y5lbP8YAwH/Hg8P/xclJ/9A3+b/Qer0/0Dm8f9A5/L/QOfy/0Dn8v9A5/L/QOfy/0Dn8v9A5/L/QObx/z/l8P9A5fD/P+Xw/z/l8P8/5fD/P+Xw/z/l8P8/5fD/P+Xw/z/l8P8/5fD/P+Xw/z/l8P9A5fD/QOXw/z/j7v9E7vn/MomQ/yBKS/8nUVP/JWFl/0Dj6v9A6fP/QObx/0Dn8v9A5/L/QOfy/0Dn8v9A5/L/QOfy/0Dn8v9A5/L/QOfy/0Dn8v9A5/L/QOfy/0Dn8v9A5/L/QOfy/0Dn8v9A5/L/QOfy/0Dn8v9A5/L/QOfy/0Dn8v9A5/L/Qefx/0Hm8v9D7Pb/Re/6/0bv+/9F7/f/Qefz/0Dn8v9A5/L/QOfy/0Dn8v9A5/L/QOfy/0Dn8v9A5/L/QOfy/0Dn8v9A5/L/QOfy/0Dn8v9A5/L/QOfy/0Dn8v9A5/L/QOfy/0Dn8v9A5/L/QOfy/0Dn8v9A5/L/QOfy/0Dn8v9A5/L/QOfy/0Dm8f8/5fH/P+Xx/z/l8f9A5/L/QOfy/0Dn8v9A5/L/QOfy/0Dn8v9A5/L/QOfy/0Dn8v9A5/L/QOfy/0Dn8v9A5/L/QOfy/0Dn8v9A5/L/QOfy/0Dn8v9A5/L/QOfy/0Dn8v9A5/L/QOfy/0Dn8v9A5/L/QOfy/0Dn8v9A5/L/QOfy/0Dn8v9A5/L/QOfy/0Dn8v9A5/L/QOfy/0Dn8v9A5/L/AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="
const (
	commonExtensions = 0 |
		md.EXTENSION_NO_INTRA_EMPHASIS |
		md.EXTENSION_TABLES |
		md.EXTENSION_FENCED_CODE |
		md.EXTENSION_AUTOLINK |
		md.EXTENSION_STRIKETHROUGH |
		md.EXTENSION_SPACE_HEADERS |
		md.EXTENSION_HEADER_IDS |
		md.EXTENSION_BACKSLASH_LINE_BREAK |
		md.EXTENSION_DEFINITION_LISTS
)

type MDFiles struct {
	Name, Path string
}

var (
	globalAddr, globalPath string
)

var tilteFiles = make([]MDFiles, 0)

func WalkDir(dirPth string) (files []MDFiles, err error) {
	files = make([]MDFiles, 0)
	err = filepath.Walk(dirPth, func(filename string, fi os.FileInfo, err error) error {
		if !fi.IsDir() {
			if strings.HasSuffix(strings.ToUpper(fi.Name()), strings.ToUpper(".md")) || strings.HasSuffix(strings.ToUpper(fi.Name()), strings.ToUpper(".markdown")) {

				if dirPth == "." {
					files = append(files, MDFiles{strings.TrimSuffix(fi.Name(), path.Ext(fi.Name())), "/" + filename})
				} else {
					files = append(files, MDFiles{strings.TrimSuffix(fi.Name(), path.Ext(fi.Name())), "/" + strings.Replace(filename, dirPth, "", -1)})
				}
			}
		}
		return nil
	})
	return files, err
}

func markdown(in io.Reader, out io.Writer) error {
	buf, err := ioutil.ReadAll(in)
	if err != nil {
		return err
	}

	body := blackfriday.Run(buf, blackfriday.WithExtensions(commonExtensions))
	m := map[string]interface{}{
		"body":  string(body),
		"files": tilteFiles,
		"name":  "文档管理中心",
	}
	return template.Must(template.New("markdown").Parse(tpl)).Execute(out, m)
}

func handleServerMarkdown(w http.ResponseWriter, r *http.Request) {
	handleFuncHttp(w, r)
	var code = 200
	var err error
	defer func() {
		log.Printf("%s %d %s", r.Method, code, r.URL.Path)
		if err != nil {
			w.Header().Add("Access-Control-Allow-Origin", "*")
			w.WriteHeader(code)
			io.WriteString(w, err.Error())
		}
	}()

	var file string

	if r.URL.Path == "/favicon.ico" {
		w.Header().Set("Content-Type", "image/x-icon;charset=UTF-8")
		w.WriteHeader(code)
		i := strings.Index(ICO, ",")
		if i < 0 {
			log.Fatal("no comma")
		}
		dec := base64.NewDecoder(base64.StdEncoding, strings.NewReader(ICO[i+1:]))
		io.Copy(w, dec)
		return
	} else if (r.URL.Path == "/" || r.URL.Path == "/index.html" ) && len(tilteFiles) > 0 {
		file = filepath.Join(globalPath, tilteFiles[0].Path)
	} else {
		file = filepath.Join(globalPath, r.URL.Path)
	}

	if !(strings.HasSuffix(file, ".md") || strings.HasSuffix(file, ".markdown")) {
		http.FileServer(http.Dir(globalPath)).ServeHTTP(w, r)
		return
	}

	f, err := os.Open(file)
	if err != nil {
		if os.IsNotExist(err) {
			code = 404
		}
		return
	}
	defer f.Close()

	err = markdown(f, w)
	if err != nil {
		code = 500
		return
	}
}

func hasSuffix(url string, prefix []string) bool {

	for _, p := range prefix {
		if strings.HasSuffix(url, p) {
			return true
		}
	}
	return false
}

func handleFuncHttp(w http.ResponseWriter, r *http.Request) {

	w.Header().Add("Access-Control-Allow-Origin", "*")

	if hasSuffix(r.URL.Path, []string{".jpg", ".css", ".png", ".png", ".js", ".gif"}) {
		w.Header().Add("Cache-Control", "public, max-age=604800, must-revalidate")
	} else {
		w.Header().Add("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Add("Pragma", "no-cache")
		w.Header().Add("Expires", "0")
	}

}

func handleProxy(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handleFuncHttp(w, r)
		h.ServeHTTP(w, r)
	}
}
func RunMarkDownServer(args ...string) {

	globalAddr, globalPath = args[0], args[1]
	fmt.Println(LOGO)
	log.Printf("Listening on %s,  path  %s", globalAddr, globalPath)

	fs := assetfs.AssetFS{
		Asset:     asset.Asset,
		AssetDir:  asset.AssetDir,
		AssetInfo: asset.AssetInfo,
		Prefix:    "static",
	}
	path := filepath.Join(globalPath, ".")
	files, _ := WalkDir(path)
	tilteFiles = files
	http.HandleFunc("/static/", handleProxy(http.StripPrefix("/static/", http.FileServer(&fs))))
	http.HandleFunc("/", handleServerMarkdown)

	log.Fatal(http.ListenAndServe(globalAddr, nil))

}
