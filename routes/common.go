package routes

import (
	//"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Azareal/Gosora/common"
)

var successJSONBytes = []byte(`{"success":"1"}`)

func ParseSEOURL(urlBit string) (slug string, id int, err error) {
	halves := strings.Split(urlBit, ".")
	if len(halves) < 2 {
		halves = append(halves, halves[0])
	}
	tid, err := strconv.Atoi(halves[1])
	return halves[0], tid, err
}

func doPush(w http.ResponseWriter, header *common.Header) {
	//fmt.Println("in doPush")
	if common.Config.EnableCDNPush {
		// TODO: Faster string building...
		var sbuf string
		var push = func(in []string) {
			for _, path := range in {
				sbuf += "</static/" + path + ">; rel=preload; as=script,"
			}
		}
		push(header.Scripts)
		//push(header.PreScriptsAsync)
		push(header.ScriptsAsync)

		if len(header.Stylesheets) > 0 {
			for _, path := range header.Stylesheets {
				sbuf += "</static/" + path + ">; rel=preload; as=style,"
			}
		}
		// TODO: Push avatars?

		if len(sbuf) > 0 {
			sbuf = sbuf[:len(sbuf)-1]
			w.Header().Set("Link", sbuf)
		}
	} else if !common.Config.DisableServerPush {
		//fmt.Println("push enabled")
		gzw, ok := w.(common.GzipResponseWriter)
		if ok {
			w = gzw.ResponseWriter
		}
		pusher, ok := w.(http.Pusher)
		if !ok {
			return
		}
		//fmt.Println("has pusher")

		var push = func(in []string) {
			for _, path := range in {
				//fmt.Println("pushing /static/" + path)
				err := pusher.Push("/static/"+path, nil)
				if err != nil {
					break
				}
			}
		}
		push(header.Scripts)
		//push(header.PreScriptsAsync)
		push(header.ScriptsAsync)
		push(header.Stylesheets)
		// TODO: Push avatars?
	}
}

func renderTemplate(tmplName string, w http.ResponseWriter, r *http.Request, header *common.Header, pi interface{}) common.RouteError {
	if header.CurrentUser.Loggedin {
		header.MetaDesc = ""
		header.OGDesc = ""
	} else if header.MetaDesc != "" && header.OGDesc == "" {
		header.OGDesc = header.MetaDesc
	}
	// TODO: Expand this to non-HTTPS requests too
	if !header.LooseCSP && common.Site.EnableSsl {
		w.Header().Set("Content-Security-Policy", "default-src 'self' 'unsafe-eval'; style-src 'self' 'unsafe-eval' 'unsafe-inline'; img-src * data: 'unsafe-eval' 'unsafe-inline'; connect-src * 'unsafe-eval' 'unsafe-inline'; frame-src 'self' www.youtube-nocookie.com;upgrade-insecure-requests")
	}
	header.AddScript("global.js")

	// Server pushes can backfire on certain browsers, so we want to make sure it's only triggered for ones where it'll help
	lastAgent := header.CurrentUser.LastAgent
	//fmt.Println("lastAgent:", lastAgent)
	if lastAgent == "chrome" || lastAgent == "firefox" {
		doPush(w, header)
	}

	if header.CurrentUser.IsAdmin {
		header.Elapsed1 = time.Since(header.StartedAt).String()
	}
	if common.RunPreRenderHook("pre_render_"+tmplName, w, r, &header.CurrentUser, pi) {
		return nil
	}
	err := header.Theme.RunTmpl(tmplName, pi, w)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	return nil
}
