package cache

import (
	"encoding/json"
	"fdicm/internal/api"
	"os"
	"path/filepath"
	"time"
)

type HistoryMetadata struct {
	Word       string    `json:"word"`
	Timestamp  time.Time `json:"timestamp"`
	QueryCount int       `json:"query_count"`
}

func getCacheDir() string {
	dir := filepath.Join(os.Getenv("HOME"), ".config", "fdicm")
	_ = os.MkdirAll(dir, 0755)
	return dir
}

func SaveWord(w string, data api.Response) {
	b, _ := json.Marshal(data)
	_ = os.WriteFile(filepath.Join(getCacheDir(), "w_"+w+".json"), b, 0644)
}

func LoadWord(w string) (api.Response, bool) {
	b, err := os.ReadFile(filepath.Join(getCacheDir(), "w_"+w+".json"))
	if err != nil {
		return nil, false
	}
	var out api.Response
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, false
	}
	return out, true
}

func SaveHistory(w string) {
	if w == "" {
		return
	}
	records := LoadHistoryExtended()
	updated := false

	for i, r := range records {
		if r.Word == w {
			records[i].QueryCount++
			records[i].Timestamp = time.Now()
			updated = true
			break
		}
	}

	if !updated {
		records = append([]HistoryMetadata{{
			Word:       w,
			Timestamp:  time.Now(),
			QueryCount: 1,
		}}, records...)
	}

	b, _ := json.Marshal(records)
	_ = os.WriteFile(filepath.Join(getCacheDir(), "history_v2.json"), b, 0644)
}

func LoadHistory() []string {
	meta := LoadHistoryExtended()
	var out []string
	for _, m := range meta {
		out = append(out, m.Word)
	}
	return out
}

func LoadHistoryExtended() []HistoryMetadata {
	b, err := os.ReadFile(filepath.Join(getCacheDir(), "history_v2.json"))
	if err != nil {
		return []HistoryMetadata{}
	}
	var out []HistoryMetadata
	_ = json.Unmarshal(b, &out)
	return out
}
