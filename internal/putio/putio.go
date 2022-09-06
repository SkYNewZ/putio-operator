package putio

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/putdotio/go-putio"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Client is a Put.io client on top of github.com/putdotio/go-putio
// Add missing features like RSS management.
type Client struct {
	*putio.Client
	Rss    RssService
	tracer trace.Tracer
}

func New(ctx context.Context, httpClient *http.Client) *Client {
	tracer := otel.GetTracerProvider().Tracer("putio")

	_, span := tracer.Start(ctx, "putio.New")
	defer span.End()

	client := putio.NewClient(httpClient)
	c := &Client{Client: client, tracer: tracer}
	c.Rss = &rssService{c}
	return c
}

type rssService struct {
	client *Client
}

// List RSS feeds.
func (s *rssService) List(ctx context.Context) ([]*Feed, error) {
	ctx, span := s.client.tracer.Start(ctx, "putio.rssService.List")
	defer span.End()

	req, err := s.client.NewRequest(ctx, http.MethodGet, "/v2/rss/list", nil)
	if err != nil {
		return nil, err
	}

	var r struct {
		Feeds []*Feed `json:"feeds"`
	}
	_, err = s.client.Do(req, &r)
	if err != nil {
		return nil, err
	}

	return r.Feeds, nil
}

// Get a RSS feed.
func (s *rssService) Get(ctx context.Context, id uint) (*Feed, error) {
	ctx, span := s.client.tracer.Start(ctx, "putio.rssService.Get")
	defer span.End()

	span.SetAttributes(attribute.Int("id", int(id)))

	req, err := s.client.NewRequest(ctx, http.MethodGet, "/v2/rss/"+strconv.Itoa(int(id)), nil)
	if err != nil {
		return nil, err
	}

	var r struct {
		Feed *Feed `json:"feed"`
	}
	_, err = s.client.Do(req, &r)
	if err != nil {
		return nil, err
	}

	return r.Feed, nil
}

// Delete a RSS feed.
func (s *rssService) Delete(ctx context.Context, id uint) error {
	ctx, span := s.client.tracer.Start(ctx, "putio.rssService.Delete")
	defer span.End()

	span.SetAttributes(attribute.Int("id", int(id)))

	req, err := s.client.NewRequest(ctx, http.MethodPost, fmt.Sprintf("/v2/rss/%d/delete", id), nil)
	if err != nil {
		return err
	}

	_, err = s.client.Do(req, nil)
	if err != nil {
		return err
	}
	return nil
}

// Create an RSS feed.
func (s *rssService) Create(ctx context.Context, feed *Feed) (*Feed, error) {
	ctx, span := s.client.tracer.Start(ctx, "putio.rssService.Create")
	defer span.End()

	span.SetAttributes(attribute.String("title", feed.Title))

	params := url.Values{}
	params.Set("title", feed.Title)
	params.Set("rss_source_url", feed.RssSourceURL)
	params.Set("keyword", feed.Keyword)
	params.Set("unwanted_keywords", feed.UnwantedKeywords)

	params.Set("delete_old_files", boolToString(feed.DeleteOldFiles))
	params.Set("dont_process_whole_feed", boolToString(feed.DontProcessWholeFeed))
	params.Set("paused", boolToString(feed.Paused))

	if v := feed.ParentDirID; v != nil {
		params.Set("parent_dir_id", strconv.Itoa(int(*v)))
	}

	req, err := s.client.NewRequest(ctx, http.MethodPost, "/v2/rss/create", strings.NewReader(params.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	var r struct {
		Feed *Feed `json:"feed"`
	}
	_, err = s.client.Do(req, &r)
	if err != nil {
		return nil, err
	}

	return r.Feed, nil
}

// Update an RSS feed.
func (s *rssService) Update(ctx context.Context, feed *Feed, id uint) error {
	ctx, span := s.client.tracer.Start(ctx, "putio.rssService.Update")
	defer span.End()

	span.SetAttributes(attribute.Int("id", int(id)))

	params := url.Values{}
	params.Set("title", feed.Title)
	params.Set("rss_source_url", feed.RssSourceURL)
	params.Set("keyword", feed.Keyword)
	params.Set("unwanted_keywords", feed.UnwantedKeywords)

	params.Set("delete_old_files", boolToString(feed.DeleteOldFiles))
	params.Set("dont_process_whole_feed", boolToString(feed.DontProcessWholeFeed))

	if v := feed.ParentDirID; v != nil {
		params.Set("parent_dir_id", strconv.Itoa(int(*v)))
	}

	req, err := s.client.NewRequest(ctx, http.MethodPost, fmt.Sprintf("/v2/rss/%d", id), strings.NewReader(params.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	var r struct {
		Status string `json:"status"`
	}

	if _, err := s.client.Do(req, &r); err != nil {
		return err
	}

	if r.Status != "OK" {
		return fmt.Errorf("put.io: invalid status received %q", r.Status)
	}

	return nil
}

// Pause given feed.
func (s *rssService) Pause(ctx context.Context, id uint) error {
	ctx, span := s.client.tracer.Start(ctx, "putio.rssService.Pause")
	defer span.End()

	span.SetAttributes(attribute.Int("id", int(id)))

	req, err := s.client.NewRequest(ctx, http.MethodPost, fmt.Sprintf("/v2/rss/%d/pause", id), nil)
	if err != nil {
		return err
	}

	var r struct {
		Status string `json:"status"`
	}

	if _, err := s.client.Do(req, &r); err != nil {
		return err
	}

	if r.Status != "OK" {
		return fmt.Errorf("put.io: invalid status received %q", r.Status)
	}

	return nil
}

// Resume given feed.
func (s *rssService) Resume(ctx context.Context, id uint) error {
	ctx, span := s.client.tracer.Start(ctx, "putio.rssService.Resume")
	defer span.End()

	span.SetAttributes(attribute.Int("id", int(id)))

	req, err := s.client.NewRequest(ctx, http.MethodPost, fmt.Sprintf("/v2/rss/%d/resume", id), nil)
	if err != nil {
		return err
	}

	var r struct {
		Status string `json:"status"`
	}

	if _, err := s.client.Do(req, &r); err != nil {
		return err
	}

	if r.Status != "OK" {
		return fmt.Errorf("put.io: invalid status received %q", r.Status)
	}

	return nil
}

func boolToString(b bool) string {
	return fmt.Sprintf("%t", b)
}
