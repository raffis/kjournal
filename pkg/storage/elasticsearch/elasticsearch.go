package elasticsearch

import "encoding/json"

type esHit struct {
	Index   string          `json:"_index"`
	DocType string          `json:"_type"`
	ID      string          `json:"_id"`
	Sort    []interface{}   `json:"sort"`
	Score   float64         `json:"_score"`
	Source  json.RawMessage `json:"_source"`
}

type esResults struct {
	Took     int64    `json:"took"`
	TimedOut bool     `json:"timed_out"`
	PitID    string   `json:"pit_id"`
	Hits     esHits   `json:"hits"`
	Shards   esShards `json:"_shards"`
}

type esShards struct {
	Total      int64 `json:"total"`
	Successful int64 `json:"successful"`
	Skipped    int64 `json:"skipped"`
	Failed     int64 `json:"failed"`
	Failures   []struct {
		Shard  int64  `json:"shard"`
		Index  string `json:"index"`
		Node   string `json:"node"`
		Reason struct {
			Type   string `json:"type"`
			Reason string `json:"reason"`
		} `json:"reason"`
	} `json:"failures"`
}

type esHits struct {
	Hits     []esHit `json:"hits"`
	Took     float64 `json:"took"`
	MaxScore float64 `json:"max_score"`
	Total    struct {
		Value    int64  `json:"value"`
		Relation string `json:"relation"`
	} `json:"total"`
}
