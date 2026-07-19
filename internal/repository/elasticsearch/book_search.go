package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/BackpropagationOfRegret/db-proj-library/internal/domain"
)

func (c *Client) Search(ctx context.Context, query domain.SearchQuery) (*domain.SearchResult, error) {
	page := query.Page
	if page <= 0 {
		page = 1
	}
	size := query.Size
	if size <= 0 {
		size = 20
	}
	from := (page - 1) * size

	must := make([]any, 0, 2)
	if q := strings.TrimSpace(query.Query); q != "" {
		must = append(must, map[string]any{
			"multi_match": map[string]any{
				"query":  q,
				"fields": []string{"title^3", "authors^2", "description", "isbn"},
				"type":   "best_fields",
			},
		})
	} else {
		must = append(must, map[string]any{"match_all": map[string]any{}})
	}

	filter := make([]any, 0, 1)
	if len(query.Genres) > 0 {
		filter = append(filter, map[string]any{
			"terms": map[string]any{
				"genres": query.Genres,
			},
		})
	}

	body := map[string]any{
		"from": from,
		"size": size,
		"query": map[string]any{
			"bool": map[string]any{
				"must":   must,
				"filter": filter,
			},
		},
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal search body: %w", err)
	}

	res, err := c.ES.Search(
		c.ES.Search.WithContext(ctx),
		c.ES.Search.WithIndex(c.IndexName),
		c.ES.Search.WithBody(bytes.NewReader(payload)),
		c.ES.Search.WithTrackTotalHits(true),
	)
	if err != nil {
		return nil, fmt.Errorf("search books: %w", err)
	}
	defer res.Body.Close()
	if res.IsError() {
		raw, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("search books: %s", string(raw))
	}

	var parsed struct {
		Hits struct {
			Total struct {
				Value int64 `json:"value"`
			} `json:"total"`
			Hits []struct {
				Score  float64             `json:"_score"`
				Source domain.BookDocument `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}
	if err := json.NewDecoder(res.Body).Decode(&parsed); err != nil {
		return nil, fmt.Errorf("decode search response: %w", err)
	}

	hits := make([]domain.SearchHit, 0, len(parsed.Hits.Hits))
	for _, hit := range parsed.Hits.Hits {
		hits = append(hits, domain.SearchHit{
			Book:  hit.Source,
			Score: hit.Score,
		})
	}

	return &domain.SearchResult{
		Hits:  hits,
		Total: parsed.Hits.Total.Value,
		Page:  page,
		Size:  size,
	}, nil
}
