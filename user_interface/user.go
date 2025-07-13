package user_interface

import (
	"bytes"
	_ "embed"
	"html/template"
	"log"
	"net/http"

	"github.com/CEKlopfenstein/gotify-repeater/gotify_api"
	"github.com/CEKlopfenstein/gotify-repeater/rssreader"
	"github.com/CEKlopfenstein/gotify-repeater/storage"
	"github.com/CEKlopfenstein/gotify-repeater/structs"
	"github.com/gin-gonic/gin"
)

//go:embed main.html
var main string

//go:embed cards/logger-card.html
var loggerCardBody string

//go:embed cards/general-info-card.html
var generalInfoCardBody string

//go:embed wrapper.html
var wrapper string

//go:embed htmx.min.js
var htmxMinJS string

//go:embed main.js
var mainJS string

//go:embed bootstrap.min.css
var bootstrap string

type userPage struct {
	HtmxBasePath string
	Cards        []card
	MainJSPath   string
	Bootstrap    string
}

type card struct {
	Title string
	Body  template.HTML
}

func BuildInterface(basePath string, mux *gin.RouterGroup, rss *rssreader.RSS_Reader, hookConfig *structs.Config, c storage.Storage, hostname string, logger *log.Logger, logBuffer *bytes.Buffer) {
	var cards = []card{}

	cards = append(cards, card{Body: template.HTML(generalInfoCardBody)})
	cards = append(cards, card{Body: template.HTML(loggerCardBody)})

	var pageData = userPage{HtmxBasePath: "htmx.min.js", Cards: cards, MainJSPath: "main.js", Bootstrap: "bootstrap.min.css"}

	mux.GET("/"+pageData.HtmxBasePath, func(ctx *gin.Context) {
		ctx.Data(http.StatusOK, "text/javascript", []byte(htmxMinJS))
	})
	mux.GET("/"+pageData.MainJSPath, func(ctx *gin.Context) {
		ctx.Data(http.StatusOK, "text/javascript", []byte(mainJS))
	})
	mux.GET("/"+pageData.Bootstrap, func(ctx *gin.Context) {
		ctx.Data(http.StatusOK, "text/css", []byte(bootstrap))
	})

	mux.GET("/", func(ctx *gin.Context) {
		var clientKey = ctx.Request.Header.Get("X-Gotify-Key")
		if len(clientKey) == 0 {
			tmpl, err := template.New("").Parse(wrapper)
			if err != nil {
				logger.Println(err)
				ctx.Done()
				return
			}
			err = tmpl.Execute(ctx.Writer, pageData)
			if err != nil {
				logger.Println(err)
			}
			ctx.Done()
		} else {
			var server = rss.GetGotifyApi()
			var failed = server.CheckToken(clientKey)
			if failed != nil {
				logger.Println(failed)
				ctx.Data(http.StatusOK, "text/html", []byte("<h2>Unauthorized token. Redirecting to main page.</h2><script>window.location = '/';</script>"))
				ctx.Done()
				return
			}
			tmpl, err := template.New("").Parse(main)
			if err != nil {
				logger.Println(err)
				ctx.Done()
				return
			}
			err = tmpl.Execute(ctx.Writer, pageData)
			if err != nil {
				logger.Println(err)
			}
		}

	})

	internalGotifyApi := gotify_api.SetupGotifyApi(hostname, "")
	mux.Use(func(ctx *gin.Context) {
		var clientKey = ctx.Request.Header.Get("X-Gotify-Key")
		if len(clientKey) == 0 {
			ctx.Data(http.StatusUnauthorized, "text/html", []byte("X-Gotify-Key Missing"))
			ctx.Done()
			return
		}

		var failed = internalGotifyApi.UpdateToken(clientKey)
		if failed != nil {
			logger.Println(failed)
			ctx.Data(http.StatusUnauthorized, "application/json", []byte(failed.Error()))
			ctx.Done()
			return
		}
		ctx.Set("token", clientKey)
		ctx.Next()
	})

	mux.GET("/logs", func(ctx *gin.Context) {
		ctx.Data(http.StatusOK, "text/html", logBuffer.Bytes())
	})

	mux.GET("/getLoginToken", func(ctx *gin.Context) {
		ctx.Data(http.StatusOK, "text/html", []byte(ctx.GetString("token")))
	})

	mux.GET("/defaultToken", func(ctx *gin.Context) {
		var token = c.GetClientToken()
		if len(token) == 0 {
			ctx.Data(http.StatusOK, "text/html", []byte(`<div hx-target="this" hx-swap="outerHTML">
			<div>No Token Set. Select an option below to set one.</div>
			<button class="btn btn-secondary m-1" hx-put="defaultToken" hx-vals='js:{"token":localStorage.getItem("gotify-login-key")}'>Use Current Client Token</button><button class="btn btn-secondary m-1" hx-put="defaultToken" hx-vals='{"token":"new"}'>Create Custom Client Token</button>
			</div>`))
		} else {
			ctx.Data(http.StatusOK, "text/html", []byte(`<div hx-target="this" hx-swap="outerHTML">
			<div>Current Default Token: `+token+`</div>
			<div>Use Options Below to Change Token</div>
			<button class="btn btn-secondary m-1" hx-put="defaultToken" hx-vals='js:{"token":localStorage.getItem("gotify-login-key")}'>Use Current Client Token</button><button class="btn btn-secondary m-1" hx-put="defaultToken" hx-vals='{"token":"new"}'>Create Custom Client Token</button>
			</div>`))
		}

	})

	mux.PUT("/defaultToken", func(ctx *gin.Context) {
		var headerToken = ctx.GetString("token")
		var token = ctx.PostForm("token")

		if token == "new" {
			currentToken := c.GetClientToken()
			if internalGotifyApi.CheckToken(currentToken) == nil {
				client := internalGotifyApi.FindClientFromToken(currentToken)
				if len(client.Token) != 0 && client.Token != headerToken && client.Name == "RSS Client" {
					internalGotifyApi.DeleteClient(client.Id)
				}
			}
			newClient, err := internalGotifyApi.CreateClient("RSS Client")
			if err != nil {
				logger.Println(err)
				ctx.Redirect(303, "defaultToken")
				return
			}
			token = newClient.Token
		}

		rss.UpdateToken(token)

		ctx.Redirect(303, "defaultToken")
	})
}
