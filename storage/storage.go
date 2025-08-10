package storage

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gotify/plugin-api"
	"github.com/mmcdole/gofeed"
)

type Storage struct {
	StorageHandler plugin.StorageHandler
	Logger         *log.Logger
	innerStore     innerStorageStruct
}

type innerStorageStruct struct {
	ClientToken string
	NextID      int
	Feeds       map[int]*Feed
}

type Feed struct {
	id       int
	Url      string
	LastDate *time.Time
	ItemUrls map[string]bool
}

func (feed *Feed) GetID() int {
	return feed.id
}

// Saves the current inner storage struct. Should be called after every save/set
func (storage *Storage) save() {
	storageBytes, _ := json.Marshal(storage.innerStore)
	storage.StorageHandler.Save(storageBytes)
}

// Loads the stored values from the DB into the current inner storage struct. Should be called before every get.
func (storage *Storage) load() {
	storageBytes, err := storage.StorageHandler.Load()
	if err != nil {
		storage.Logger.Println(err)
		return
	}

	if len(storageBytes) == 0 {
		storageBytes, _ = json.Marshal(storage.innerStore)
		storage.StorageHandler.Save(storageBytes)
	} else {
		json.Unmarshal(storageBytes, &storage.innerStore)
	}
}

func (storage *Storage) GetClientToken() string {
	storage.load()
	return storage.innerStore.ClientToken
}

func (storage *Storage) SaveClientToken(token string) {
	storage.load()
	storage.innerStore.ClientToken = token
	storage.save()
}

func (storage *Storage) GetNextFeedID() int {
	var id = 0
	var item *Feed = storage.innerStore.Feeds[id]
	for item != nil {
		id++
		item = storage.innerStore.Feeds[id]
	}
	return id
}

func (storage *Storage) SaveNewFeed(url string) *Feed {
	storage.load()
	var newID = storage.GetNextFeedID()
	if storage.innerStore.Feeds == nil {
		storage.innerStore.Feeds = make(map[int]*Feed)
		storage.save()
	}
	storage.innerStore.Feeds[newID] = &Feed{Url: url, id: newID, ItemUrls: make(map[string]bool)}
	storage.Logger.Printf("Saved New Feed: %s", url)
	storage.save()
	return storage.innerStore.Feeds[newID]
}

func (storage *Storage) GetFeedByID(id int) *Feed {
	storage.load()
	if storage.innerStore.Feeds == nil {
		return nil
	}

	return storage.innerStore.Feeds[id]
}

func (storage *Storage) RemoveFeedByID(id int) {
	storage.load()
	if storage.innerStore.Feeds == nil || storage.innerStore.Feeds[id] == nil {
		return
	}
	storage.Logger.Printf("Deleted Feed: %s", storage.innerStore.Feeds[id].Url)
	delete(storage.innerStore.Feeds, id)
	storage.save()
}

func (storage *Storage) GetFeeds() *map[int]*Feed {
	storage.load()
	if storage.innerStore.Feeds == nil {
		storage.innerStore.Feeds = make(map[int]*Feed)
		storage.save()
	}
	return &storage.innerStore.Feeds
}

func (feed *Feed) IsItemNew(item *gofeed.Item, storage *Storage) bool {
	var timeOfPost = item.UpdatedParsed
	if timeOfPost == nil {
		timeOfPost = item.PublishedParsed
	}

	if timeOfPost != nil && (feed.LastDate == nil || feed.LastDate.Compare(*timeOfPost) < 0) {
		return true
	}

	_, isPresent := feed.ItemUrls[item.Link]

	if timeOfPost == nil && !isPresent {
		println("New By Link", isPresent, item.Link)
	}

	return timeOfPost == nil && !isPresent
}

func (storage *Storage) SaveITemUrlsAndLatestDate(id int, urls []string, time *time.Time) {
	storage.innerStore.Feeds[id].LastDate = time
	storage.innerStore.Feeds[id].LastDate = time
	for _, url := range urls {
		storage.innerStore.Feeds[id].ItemUrls[url] = true
	}
	storage.save()
}
