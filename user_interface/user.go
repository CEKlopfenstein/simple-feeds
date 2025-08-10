package user_interface

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/CEKlopfenstein/gotify-repeater/gotify_api"
	"github.com/CEKlopfenstein/gotify-repeater/rssreader"
	"github.com/CEKlopfenstein/gotify-repeater/structs"
	"github.com/gin-gonic/gin"
	"github.com/mmcdole/gofeed"
)

//go:embed main.html
var main string

//go:embed cards/logger-card.html
var loggerCardBody string

//go:embed cards/general-info-card.html
var generalInfoCardBody string

//go:embed cards/feed-card.html
var feedCardBody string

//go:embed cards/new-feed-card.html
var newFeedCardBody string

//go:embed cards/blank-card-wrapper.html
var blankCardWrapper string

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

type feedCardData struct {
	Id        int
	LastFound string
	TimeSince string
	Url       string
	Descript  string
	Title     string
}

func BuildInterface(basePath string, mux *gin.RouterGroup, rss *rssreader.RSS_Reader, hookConfig *structs.Config, hostname string, logger *log.Logger, logBuffer *bytes.Buffer) {
	var cards = []card{}

	var generalInfoCard = card{Body: template.HTML(generalInfoCardBody)}
	cards = append(cards, generalInfoCard)
	var loggerCard = card{Body: template.HTML(loggerCardBody)}
	cards = append(cards, loggerCard)
	var feedsCard = card{Body: template.HTML("<span hx-get='feeds' hx-trigger='load' hx-target='closest div' hx-swap='outerHTML'></span>")}
	var newFeedCard = card{Title: "Create Feed", Body: template.HTML(newFeedCardBody)}

	var pageData = userPage{HtmxBasePath: "htmx.min.js", Cards: cards, MainJSPath: "main.js", Bootstrap: "bootstrap.min.css"}

	wrapperTemplate, wrapperTemplateParseError := template.New("").Parse(wrapper)
	if wrapperTemplateParseError != nil {
		logger.Println("Failed to parse Wrapper Template")
		logger.Println(wrapperTemplateParseError.Error())
		return
	}

	feedCardTemplate, feedCardParseError := template.New("").Parse(feedCardBody)
	if feedCardParseError != nil {
		logger.Println("Failed to parse Feed Card Template")
		logger.Println(feedCardParseError.Error())
		return
	}

	cardWrapperTemplate, cardWrapperError := template.New("").Parse(blankCardWrapper)
	if cardWrapperError != nil {
		logger.Println("Failed to parse Blank Card Template")
		logger.Println(cardWrapperError.Error())
		return
	}

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

			var cards = []card{}
			cards = append(cards, generalInfoCard)
			cards = append(cards, feedsCard)
			cards = append(cards, newFeedCard)
			cards = append(cards, loggerCard)
			pageData.Cards = cards

			err := wrapperTemplate.Execute(ctx.Writer, pageData)
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
		var token = rss.Storage.GetClientToken()
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
			currentToken := rss.Storage.GetClientToken()
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

	mux.GET("/feeds", func(ctx *gin.Context) {
		var feeds = ""
		for id, _ := range *rss.Storage.GetFeeds() {
			feeds += fmt.Sprintf("<div hx-swap='outerHTML' hx-get='feed/%d' hx-trigger='load'></div>", id)
		}
		ctx.Data(http.StatusOK, "text/html", []byte(feeds))
	})

	mux.POST("/create-feed", func(ctx *gin.Context) {
		var feedUrl = ctx.PostForm("feed-url")

		var finalHTML = new(bytes.Buffer)
		feedData, feedError := gofeed.NewParser().ParseURL(feedUrl)
		if feedError == nil && feedData != nil {
			var feed = rss.Storage.SaveNewFeed(feedUrl)
			var id = feed.GetID()

			var cardData = feedCardData{Id: id, Url: feed.Url, Descript: feedData.Description, Title: feedData.Title}

			if feed.LastDate != nil {
				cardData.LastFound = feed.LastDate.Round(time.Second).String()
				cardData.TimeSince = time.Since(*feed.LastDate).Round(time.Second).String()
			}
			feedCardTemplate.Execute(finalHTML, cardData)
		} else {
			logger.Printf("Failed to add: %s", feedUrl)
		}

		cardWrapperTemplate.Execute(finalHTML, newFeedCard)

		ctx.Data(http.StatusOK, "text/html", []byte(finalHTML.String()))
	})

	feedsGroup := mux.Group("/feed/:feedID", func(ctx *gin.Context) {
		var feeds = *rss.Storage.GetFeeds()
		var id = ctx.Param("feedID")
		var intId, _ = strconv.Atoi(id)

		var feed = feeds[intId]
		if feed == nil {
			ctx.Data(http.StatusNotFound, "text/html", []byte("Invalid ID"))
			return
		}

		ctx.Set("ID", intId)
		ctx.Next()
	})

	feedsGroup.GET("/", func(ctx *gin.Context) {
		var id = ctx.GetInt("ID")
		var feed = (*rss.Storage.GetFeeds())[id]

		var finalHTML = new(bytes.Buffer)
		feedData, _ := gofeed.NewParser().ParseURL(feed.Url)
		var cardData feedCardData
		if feedData != nil {
			cardData = feedCardData{Id: id, Url: feed.Url, Descript: feedData.Description, Title: feedData.Title}
		} else {
			cardData = feedCardData{Id: id, Url: feed.Url, Title: "Invalid URL"}
		}

		if feed.LastDate != nil {
			cardData.LastFound = feed.LastDate.Round(time.Second).String()
			cardData.TimeSince = time.Since(*feed.LastDate).Round(time.Second).String()
		}
		feedCardTemplate.Execute(finalHTML, cardData)

		ctx.Data(http.StatusOK, "text/html", []byte(finalHTML.String()))
	})

	feedsGroup.DELETE("/", func(ctx *gin.Context) {
		var id = ctx.GetInt("ID")
		rss.Storage.RemoveFeedByID(id)
		ctx.Data(http.StatusOK, "text/html", []byte(""))
	})
}
