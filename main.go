package youtubesearch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type youtubeResponse struct {
	Contents contents `json:"contents"`
	json.RawMessage
}

type contents struct {
	TwoColumnSearchResultsRenderer struct {
		PrimaryContents struct {
			SectionListRenderer struct {
				Contents []map[string]interface{} `json:"contents"`
			} `json:"sectionListRenderer"`
		} `json:"primaryContents"`
	} `json:"twoColumnSearchResultsRenderer"`
}

type VideoData struct {
	ID          string
	Thumbnails  []string
	Title       string
	LongDesc    string
	Channel     string
	Duration    string
	Views       string
	PublishTime string
	URLSuffix   string
}

const MAX_RESULTS = 10
const BASE_URL = "https://youtube.com"

func Search(query string, timeout time.Duration) ([]VideoData, error) {
	var res []byte
	done := make(chan struct{})
	v := url.Values{}
	v.Add("search_query", query)
	url := fmt.Sprintf("%s/results?%s", BASE_URL, v.Encode())
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	if timeout > 0 {
		go func() {
			res, err = search(req)
			done <- struct{}{}
		}()
		select {
		case <-done:
			break
		case <-time.After(timeout):
			return nil, fmt.Errorf("failed to perform search in time. Please try again")
		}
	} else {
		res, err = search(req)
	}
	if err != nil {
		return nil, err
	}
	results := parseResponse(res)
	return results, nil
}

// TODO: Add exponential backoff and handle errors better
func search(request *http.Request) ([]byte, error) {
	dataBytes := []byte("ytInitialData")
	for {
		res, err := http.DefaultClient.Do(request)
		if err != nil {
			return nil, err
		}

		if res.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("client: status code isn't okay: %d", res.StatusCode)
		}
		resBody, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}

		pos := bytes.Index(resBody, dataBytes)
		if pos > -1 {
			resBody = resBody[pos+len(dataBytes)+3:]
			end := bytes.Index(resBody, []byte("};"))
			return resBody[:end+1], nil
		}
	}
}

func parseResponse(response []byte) []VideoData {
	data := make(map[string]json.RawMessage)
	if err := json.Unmarshal(response, &data); err != nil {
		panic(err)
	}
	var yt youtubeResponse
	if err := json.Unmarshal(data["contents"], &yt.Contents); err != nil {
		panic(err)
	}

	var result []VideoData

	for _, content := range yt.Contents.TwoColumnSearchResultsRenderer.PrimaryContents.SectionListRenderer.Contents {
		c, ok := content["itemSectionRenderer"]
		if !ok {
			continue
		}
		itemSectionRenderer := c.(map[string]interface{})
		contents := itemSectionRenderer["contents"].([]interface{})
		for _, item := range contents {
			item := item.(map[string]interface{})
			v, ok := item["videoRenderer"]
			if !ok {
				continue
			}
			videoData := parseItem(v.(map[string]interface{}))
			result = append(result, videoData)
		}
	}
	return result
}

func parseItem(item map[string]interface{}) VideoData {
	var videoData VideoData
	id, ok := item["videoId"]
	if ok {
		videoData.ID = id.(string)
	}

	thumbnail, ok := item["thumbnail"]
	if ok {
		thumbnail := thumbnail.(map[string]interface{})
		thumbnails, ok := thumbnail["thumbnails"]
		if ok {
			thumbnails := thumbnails.([]interface{})
			for _, thumb := range thumbnails {
				thumb := thumb.(map[string]interface{})
				url, ok := thumb["url"]
				if ok {
					videoData.Thumbnails = append(videoData.Thumbnails, url.(string))
				}
			}
		}
	}

	title, ok := item["title"]
	if ok {
		videoData.Title = getTextFromRuns(title.(map[string]interface{}))
	}

	longDesc, ok := item["descriptionSnippet"]
	if ok {
		videoData.LongDesc = getTextFromRuns(longDesc.(map[string]interface{}))
	}

	longBylineText, ok := item["longBylineText"]
	if ok {
		videoData.Channel = getTextFromRuns(longBylineText.(map[string]interface{}))
	}

	lengthText, ok := item["lengthText"]
	if ok {
		videoData.Duration = getSimpleText(lengthText.(map[string]interface{}))
	}

	viewCountText, ok := item["viewCountText"]
	if ok {
		videoData.Views = getSimpleText(viewCountText.(map[string]interface{}))
	}

	publishedTimeText, ok := item["publishedTimeText"]
	if ok {
		videoData.PublishTime = getSimpleText(publishedTimeText.(map[string]interface{}))
	}

	navigationEndpoint, ok := item["navigationEndpoint"]
	if ok {
		navigationEndpoint := navigationEndpoint.(map[string]interface{})
		commandMetadata, ok := navigationEndpoint["commandMetadata"]
		if ok {
			commandMetadata := commandMetadata.(map[string]interface{})
			webCommandMetadata, ok := commandMetadata["webCommandMetadata"]
			if ok {
				webCommandMetadata := webCommandMetadata.(map[string]interface{})
				url, ok := webCommandMetadata["url"]
				if ok {
					videoData.URLSuffix = url.(string)
				}
			}
		}
	}
	return videoData
}

func getTextFromRuns(data map[string]interface{}) string {
	runs, ok := data["runs"]
	if ok {
		runs := runs.([]interface{})
		if len(runs) > 0 {
			t := runs[0].(map[string]interface{})
			text, ok := t["text"]
			if ok {
				return text.(string)
			}
		}
	}
	return ""
}

func getSimpleText(data map[string]interface{}) string {
	simpleText, ok := data["simpleText"]
	if ok {
		return simpleText.(string)
	}
	return ""
}
