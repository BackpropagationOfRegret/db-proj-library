package elasticsearch

// Package elasticsearch will host the go-elasticsearch client and
// book index/search implementations. Currently unused; search uses a noop repo.
type Client struct {
	URL   string
	Index string
}

func NewClient(url, index string) *Client {
	return &Client{URL: url, Index: index}
}
