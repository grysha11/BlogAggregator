package service

import (
	"fmt"
	"time"
	"context"
	"database/sql"
	"strings"
	"grysha11/BlogAggregator/internal/database"
	"grysha11/BlogAggregator/internal/rss"
	"github.com/google/uuid"
)

func ScrapeFeeds(s *State) {
	feed, err := s.DB.GetNextFeedToFetch(context.Background())
	if err != nil {
		fmt.Printf("Couldn't get next feed to fetch: %v\n", err)
		return
	}

	_, err = s.DB.MarkFeedFetched(context.Background(), database.MarkFeedFetchedParams{
		LastFetchedAt: sql.NullTime{Time: time.Now(), Valid: true},
		UpdatedAt: time.Now().UTC(),
		ID: feed.ID,
	})
	if err != nil {
		fmt.Printf("Couldn't mark fetch time of feed %v: %v\n", feed.Name, err)
		return
	}

	rssFeed, err := rss.FetchFeed(context.Background(), feed.Url)
	if err != nil {
		fmt.Printf("Couldn't fetch feed %v: %v", feed.Name, err)
	}

	for _, item := range rssFeed.Channel.Item {
		pubDate, err := time.Parse(time.RFC1123Z, item.PubDate)
		if err != nil {
			pubDate, err = time.Parse(time.RFC1123, item.PubDate)
			if err != nil {
				fmt.Printf("Could not parse Date time <%v> of item %v err: %v\n", item.PubDate, item.Title, err)
				continue
			}
		} 

		description := sql.NullString{}
		if item.Description != "" {
			description.String = item.Description
			description.Valid = true
		}

		_, err = s.DB.CreatePost(context.Background(), database.CreatePostParams{
			ID: uuid.New(),
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
			Title: item.Title,
			Url: item.Link,
			Description: description,
			PublishedAt: pubDate,
			FeedID: feed.ID,
		})
		if err != nil {
			if strings.Contains(err.Error(), "duplicate key") {
				continue
			}
			fmt.Printf("Could not create post for db: %v\n", err)
		}
	}

	fmt.Printf("Feed %v is collected, %v posts scanned\n", feed.Name, len(rssFeed.Channel.Item))
}