package search

const BookIndexMapping = `{
  "settings": {
    "number_of_shards": 1,
    "number_of_replicas": 0,
    "analysis": {
      "analyzer": {
        "book_analyzer": {
          "type": "custom",
          "tokenizer": "standard",
          "filter": ["lowercase", "asciifolding"]
        }
      }
    }
  },
  "mappings": {
    "properties": {
      "id": { "type": "long" },
      "title": {
        "type": "text",
        "analyzer": "book_analyzer",
        "fields": {
          "keyword": { "type": "keyword", "ignore_above": 256 }
        }
      },
      "isbn": { "type": "keyword" },
      "description": { "type": "text", "analyzer": "book_analyzer" },
      "publication_year": { "type": "integer" },
      "authors": {
        "type": "text",
        "analyzer": "book_analyzer",
        "fields": {
          "keyword": { "type": "keyword", "ignore_above": 256 }
        }
      },
      "genres": { "type": "keyword" },
      "available_copies": { "type": "integer" },
      "indexed_at": { "type": "date" }
    }
  }
}`
