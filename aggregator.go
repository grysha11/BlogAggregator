package main

import (
	"fmt"
	"time"
	"context"
	"database/sql"
	"grysha11/BlogAggregator/internal/database"
)

func scrapeFeeds(s *state) {
	feed, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		fmt.Printf("Couldn't get next feed to fetch: %v\n", err)
		return
	}

	_, err = s.db.MarkFeedFetched(context.Background(), database.MarkFeedFetchedParams{
		LastFetchedAt: sql.NullTime{Time: time.Now(), Valid: true},
		UpdatedAt: time.Now().UTC(),
		ID: feed.ID,
	})
	if err != nil {
		fmt.Printf("Couldn't mark fetch time of feed %v: %v\n", feed.Name, err)
		return
	}

	_, err = fetchFeed(context.Background(), feed.Url)
	if err != nil {
		fmt.Printf("Couldn't fetch feed %v: %v", feed.Name, err)
	}

	fmt.Printf("Fetched succesfully feed %v\n", feed.Name)
}