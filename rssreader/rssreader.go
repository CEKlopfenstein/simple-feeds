package rssreader

import (
	"log"
	"time"

	"github.com/CEKlopfenstein/gotify-repeater/gotify_api"
	"github.com/CEKlopfenstein/gotify-repeater/storage"
	"github.com/gorilla/websocket"
	"github.com/gotify/plugin-api"
	"github.com/mmcdole/gofeed"
)

type RSS_Reader struct {
	listener  *websocket.Conn
	gotifyApi gotify_api.GotifyApi
	Storage   storage.Storage
	userName  string
	logger    *log.Logger
}

func (rssreader *RSS_Reader) SetGotifyApi(gotifyApi gotify_api.GotifyApi) {
	rssreader.gotifyApi = gotifyApi
}

func (rssreader *RSS_Reader) SetUserName(userName string) {
	rssreader.userName = userName
}

func (rssreader *RSS_Reader) SetStorage(storage storage.Storage) {
	rssreader.Storage = storage
}

func (rssreader *RSS_Reader) SetLogger(logger *log.Logger) {
	rssreader.logger = logger
}
func (rssreader *RSS_Reader) GetGotifyApi() gotify_api.GotifyApi {
	return rssreader.gotifyApi
}

func (rssreader *RSS_Reader) UpdateToken(token string) error {
	rssreader.Storage.SaveClientToken(token)
	err := rssreader.gotifyApi.UpdateToken(token)
	if err != nil {
		return err
	}
	return nil
}

func (rssreader *RSS_Reader) CheckFeeds(msgHandler plugin.MessageHandler) {
	var feeds = rssreader.Storage.GetFeeds()
	for id, feedRecord := range *feeds {
		fp := gofeed.NewParser()
		feed, _ := fp.ParseURL(feedRecord.Url)
		if feed == nil {
			rssreader.Storage.Logger.Printf("Failed to parse: %s", feedRecord.Url)
			continue
		}
		var latest *time.Time = nil
		var urls = []string{}
		for itemIndex := len(feed.Items) - 1; itemIndex >= 0; itemIndex-- {
			var item = feed.Items[itemIndex]
			urls = append(urls, item.Link)

			var timeOfPost = item.UpdatedParsed
			if timeOfPost == nil {
				timeOfPost = item.PublishedParsed
			}

			if timeOfPost != nil && (latest == nil || latest.Compare(*timeOfPost) < 0) {
				latest = timeOfPost
			}

			if feedRecord.IsItemNew(item, &rssreader.Storage) {
				rssreader.sendRSSMessage(msgHandler, *item)
			}
		}

		rssreader.Storage.SaveITemUrlsAndLatestDate(id, urls, latest)
	}

}

func (rssreader *RSS_Reader) sendRSSMessage(msgHandler plugin.MessageHandler, item gofeed.Item) error {
	return msgHandler.SendMessage(plugin.Message{Title: item.Title + " " + item.Title, Message: item.Link})
}
