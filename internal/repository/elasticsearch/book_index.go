package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/BackpropagationOfRegret/db-proj-library/internal/domain"
	"github.com/BackpropagationOfRegret/db-proj-library/internal/search"
)

func (c *Client) Ping(ctx context.Context) error {
	res, err := c.ES.Ping(c.ES.Ping.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("elasticsearch ping: %w", err)
	}
	defer res.Body.Close()
	if res.IsError() {
		return fmt.Errorf("elasticsearch ping: %s", res.String())
	}
	return nil
}

func (c *Client) EnsureIndex(ctx context.Context) error {
	res, err := c.ES.Indices.Exists(
		[]string{c.IndexName},
		c.ES.Indices.Exists.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("check index exists: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode == 200 {
		return nil
	}

	createRes, err := c.ES.Indices.Create(
		c.IndexName,
		c.ES.Indices.Create.WithContext(ctx),
		c.ES.Indices.Create.WithBody(strings.NewReader(search.BookIndexMapping)),
	)
	if err != nil {
		return fmt.Errorf("create index: %w", err)
	}
	defer createRes.Body.Close()
	if createRes.IsError() {
		body, _ := io.ReadAll(createRes.Body)
		return fmt.Errorf("create index: %s", string(body))
	}
	return nil
}

func (c *Client) DeleteIndex(ctx context.Context) error {
	res, err := c.ES.Indices.Delete(
		[]string{c.IndexName},
		c.ES.Indices.Delete.WithContext(ctx),
		c.ES.Indices.Delete.WithIgnoreUnavailable(true),
	)
	if err != nil {
		return fmt.Errorf("delete index: %w", err)
	}
	defer res.Body.Close()
	if res.IsError() {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("delete index: %s", string(body))
	}
	return nil
}

func (c *Client) Index(ctx context.Context, doc domain.BookDocument) error {
	payload, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("marshal book document: %w", err)
	}

	res, err := c.ES.Index(
		c.IndexName,
		bytes.NewReader(payload),
		c.ES.Index.WithContext(ctx),
		c.ES.Index.WithDocumentID(strconv.FormatInt(doc.ID, 10)),
		c.ES.Index.WithRefresh("true"),
	)
	if err != nil {
		return fmt.Errorf("index book: %w", err)
	}
	defer res.Body.Close()
	if res.IsError() {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("index book: %s", string(body))
	}
	return nil
}

func (c *Client) BulkIndex(ctx context.Context, docs []domain.BookDocument) error {
	if len(docs) == 0 {
		return nil
	}

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	for _, doc := range docs {
		meta := map[string]any{
			"index": map[string]any{
				"_index": c.IndexName,
				"_id":    strconv.FormatInt(doc.ID, 10),
			},
		}
		if err := enc.Encode(meta); err != nil {
			return fmt.Errorf("encode bulk meta: %w", err)
		}
		if err := enc.Encode(doc); err != nil {
			return fmt.Errorf("encode bulk doc: %w", err)
		}
	}

	res, err := c.ES.Bulk(
		bytes.NewReader(buf.Bytes()),
		c.ES.Bulk.WithContext(ctx),
		c.ES.Bulk.WithRefresh("true"),
	)
	if err != nil {
		return fmt.Errorf("bulk index: %w", err)
	}
	defer res.Body.Close()
	if res.IsError() {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("bulk index: %s", string(body))
	}

	var bulkResp struct {
		Errors bool `json:"errors"`
		Items  []map[string]struct {
			Error *struct {
				Type   string `json:"type"`
				Reason string `json:"reason"`
			} `json:"error"`
		} `json:"items"`
	}
	if err := json.NewDecoder(res.Body).Decode(&bulkResp); err != nil {
		return fmt.Errorf("decode bulk response: %w", err)
	}
	if bulkResp.Errors {
		for _, item := range bulkResp.Items {
			if op, ok := item["index"]; ok && op.Error != nil {
				return fmt.Errorf("bulk index item: %s: %s", op.Error.Type, op.Error.Reason)
			}
		}
		return fmt.Errorf("bulk index completed with errors")
	}
	return nil
}

func (c *Client) Delete(ctx context.Context, bookID int64) error {
	res, err := c.ES.Delete(
		c.IndexName,
		strconv.FormatInt(bookID, 10),
		c.ES.Delete.WithContext(ctx),
		c.ES.Delete.WithRefresh("true"),
	)
	if err != nil {
		return fmt.Errorf("delete book doc: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode == 404 {
		return nil
	}
	if res.IsError() {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("delete book doc: %s", string(body))
	}
	return nil
}
