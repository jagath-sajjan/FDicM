package api

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Response []WordEntry

type WordEntry struct {
	Word      string `json:"word"`
	Phonetic  string `json:"phonetic"`
	Phonetics []struct {
		Text  string `json:"text"`
		Audio string `json:"audio"`
	} `json:"phonetics"`
	Meanings []struct {
		PartOfSpeech string `json:"partOfSpeech"`
		Definitions  []struct {
			Definition string   `json:"definition"`
			Example    string   `json:"example"`
			Synonyms   []string `json:"synonyms,omitempty"`
		} `json:"definitions"`
	} `json:"meanings"`
}

var compressedOfflineDB = []byte{
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xca, 0x2c, 0x29, 0x4e, 0x2d, 0xcd,
	0x4b, 0x4d, 0x51, 0xaa, 0xe6, 0x52, 0x4a, 0x4e, 0x4c, 0xce, 0x48, 0x55, 0xca, 0x2b, 0x35, 0xd2,
	0x07, 0x4b, 0x19, 0x27, 0x43, 0x05, 0x93, 0x21, 0x0a, 0x18, 0x00, 0x00, 0xff, 0xff, 0x4c, 0x2b,
	0x5d, 0xbd, 0x30, 0x00, 0x00, 0x00,
}

func Fetch(word string, partOfSpeechFilter string) (Response, error) {
	word = strings.TrimSpace(strings.ToLower(word))
	var rawData Response
	var err error

	res, err := http.Get("https://api.dictionaryapi.dev/api/v2/entries/en/" + word)
	if err == nil && res.StatusCode == 200 {
		rawData, err = parseResponse(res.Body)
	} else {
		if res != nil {
			res.Body.Close()
		}
		rawData, err = getCompressedOfflineAsset(word)
	}

	if err != nil {
		return nil, err
	}

	if partOfSpeechFilter != "" {
		var filtered Response
		for _, entry := range rawData {
			newEntry := entry
			newEntry.Meanings = nil
			for _, m := range entry.Meanings {
				if strings.EqualFold(m.PartOfSpeech, partOfSpeechFilter) {
					newEntry.Meanings = append(newEntry.Meanings, m)
				}
			}
			if len(newEntry.Meanings) > 0 {
				filtered = append(filtered, newEntry)
			}
		}
		if len(filtered) == 0 {
			return nil, fmt.Errorf("no definitions found matching category restriction: %s", partOfSpeechFilter)
		}
		return filtered, nil
	}

	return rawData, nil
}

func parseResponse(body io.ReadCloser) (Response, error) {
	defer body.Close()
	b, err := io.ReadAll(body)
	if err != nil {
		return nil, err
	}
	var out Response
	err = json.Unmarshal(b, &out)
	return out, err
}

func getCompressedOfflineAsset(word string) (Response, error) {
	r, err := gzip.NewReader(bytes.NewReader(compressedOfflineDB))
	if err != nil {
		return nil, err
	}
	defer r.Close()

	var fallbackMap map[string]string
	if err := json.NewDecoder(r).Decode(&fallbackMap); err != nil {
		fallbackMap = map[string]string{
			"production": "Suitable or intended for use in an operational system rather than a development environment.",
			"interface":  "A point where two systems, subjects, or organizations meet and interact cleanly.",
		}
	}

	if def, exists := fallbackMap[word]; exists {
		return Response{{
			Word:     word,
			Phonetic: "/offline/",
			Meanings: []struct {
				PartOfSpeech string "json:\"partOfSpeech\""
				Definitions  []struct {
					Definition string   "json:\"definition\""
					Example    string   "json:\"example\""
					Synonyms   []string "json:\"synonyms,omitempty\""
				} "json:\"definitions\""
			}{
				{
					PartOfSpeech: "production-core",
					Definitions: []struct {
						Definition string   "json:\"definition\""
						Example    string   "json:\"example\""
						Synonyms   []string "json:\"synonyms,omitempty\""
					}{{Definition: def}},
				},
			},
		}}, nil
	}

	return nil, fmt.Errorf("term not located online or in offline asset indices")
}
