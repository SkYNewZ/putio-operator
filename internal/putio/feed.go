package putio

import "time"

const customTimeLayout string = "2006-01-02T15:04:05.999999999"

type Time struct {
	time.Time
}

func (t *Time) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}

	layout := `"` + customTimeLayout + `"`
	parsedTime, err := time.Parse(layout, string(data))
	*t = Time{Time: parsedTime}
	return err //nolint:wrapcheck
}

func (t *Time) MarshalJSON() ([]byte, error) {
	return t.Time.MarshalJSON() //nolint:wrapcheck
}

func (t *Time) GetTime() time.Time {
	return t.Time
}

type Feed struct {
	ID                   *uint   `json:"id"`
	Title                string  `json:"title"`
	RssSourceURL         string  `json:"rss_source_url"`
	ParentDirID          *uint32 `json:"parent_dir_id"`
	DeleteOldFiles       bool    `json:"delete_old_files"`
	DontProcessWholeFeed bool    `json:"dont_process_whole_feed"`
	Keyword              string  `json:"keyword"`
	UnwantedKeywords     string  `json:"unwanted_keywords"`
	Paused               bool    `json:"paused"`

	Extract         bool   `json:"extract"`
	FailedItemCount uint   `json:"failed_item_count"`
	LastError       string `json:"last_error"`
	LastFetch       Time   `json:"last_fetch"`
	CreatedAt       Time   `json:"created_at"`
	PausedAt        Time   `json:"paused_at"`
	StartAt         Time   `json:"start_at"`
	UpdatedAt       Time   `json:"updated_at"`
}
