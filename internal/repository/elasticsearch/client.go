package elasticsearch

import (
	"fmt"
	"net/http"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
)

type Client struct {
	ES        *elasticsearch.Client
	IndexName string
}

func Connect(url, indexName string) (*Client, error) {
	if url == "" {
		return nil, fmt.Errorf("elasticsearch url is empty")
	}
	if indexName == "" {
		indexName = "books"
	}

	es, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{url},
		Transport: &http.Transport{
			ResponseHeaderTimeout: 30 * time.Second,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("create elasticsearch client: %w", err)
	}

	return &Client{ES: es, IndexName: indexName}, nil
}
