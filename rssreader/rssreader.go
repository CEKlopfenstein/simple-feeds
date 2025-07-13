package rssreader

import (
	"log"

	"github.com/CEKlopfenstein/gotify-repeater/gotify_api"
	"github.com/CEKlopfenstein/gotify-repeater/storage"
	"github.com/gorilla/websocket"
)

type RSS_Reader struct {
	listener  *websocket.Conn
	gotifyApi gotify_api.GotifyApi
	storage   storage.Storage
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
	rssreader.storage = storage
}

func (rssreader *RSS_Reader) SetLogger(logger *log.Logger) {
	rssreader.logger = logger
}
func (rssreader *RSS_Reader) GetGotifyApi() gotify_api.GotifyApi {
	return rssreader.gotifyApi
}

func (rssreader *RSS_Reader) UpdateToken(token string) error {
	rssreader.storage.SaveClientToken(token)
	err := rssreader.gotifyApi.UpdateToken(token)
	if err != nil {
		return err
	}
	return nil
}
