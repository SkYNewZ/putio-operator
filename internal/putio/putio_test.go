package putio

import (
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/putdotio/go-putio"
	"go.opentelemetry.io/otel"
)

type RoundTripFunc func(req *http.Request) *http.Response

func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

func NewTestClient(t *testing.T, fn RoundTripFunc) *http.Client {
	t.Helper()
	return &http.Client{
		Transport: fn,
	}
}

// readGoldenFile reads given file name
func readGoldenFile(t *testing.T, filename string) io.ReadCloser {
	t.Helper()
	goldenfile := filepath.Join("testdata", filename+".json")
	file, err := os.Open(goldenfile)
	if err != nil {
		t.Fatal("error reading golden file:", err)
	}

	return file
}

func Test_rssService_List(t *testing.T) {
	type fields struct {
		client *Client
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*Feed
		wantErr bool
	}{
		{
			name: "expected",
			fields: fields{
				client: &Client{
					Client: putio.NewClient(NewTestClient(t, func(req *http.Request) *http.Response {
						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       readGoldenFile(t, "list"),
							Header:     make(http.Header),
						}
					})),
					Rss:    nil, // currently tested
					tracer: otel.GetTracerProvider().Tracer("putio-testing"),
				},
			},
			args: args{context.Background()},
			want: func() []*Feed {
				feedID := uint(125559)
				return []*Feed{
					{
						ID:                   &feedID,
						Title:                "For all mankind",
						RssSourceURL:         "https://rss.site.fr",
						ParentDirID:          998868232,
						DeleteOldFiles:       false,
						DontProcessWholeFeed: false,
						Keyword:              "FOR.ALL.MANKIND&S03&2160P&FRATERNITY",
						UnwantedKeywords:     "",
						Paused:               false,
						Extract:              false,
						FailedItemCount:      0,
						LastError:            "",
						LastFetch:            Time{time.Date(2022, time.September, 11, 19, 46, 39, 0, time.UTC)},
						CreatedAt:            Time{time.Date(2022, time.June, 13, 00, 01, 52, 0, time.UTC)},
						PausedAt:             Time{},
						StartAt:              Time{},
						UpdatedAt:            Time{time.Date(2022, time.September, 11, 19, 46, 39, 0, time.UTC)},
					},
				}
			}(),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &rssService{
				client: tt.fields.client,
			}
			got, err := s.List(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("List() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func Test_rssService_Get(t *testing.T) {
	type fields struct {
		client *Client
	}
	type args struct {
		ctx context.Context
		id  uint
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *Feed
		wantErr bool
	}{
		{
			name: "expected",
			fields: fields{
				client: &Client{
					Client: putio.NewClient(NewTestClient(t, func(req *http.Request) *http.Response {
						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       readGoldenFile(t, "get"),
							Header:     make(http.Header),
						}
					})),
					Rss:    nil, // currently tested
					tracer: otel.GetTracerProvider().Tracer("putio-testing"),
				},
			},
			args: args{
				ctx: context.Background(),
			},
			want: func() *Feed {
				feedID := uint(125559)
				return &Feed{
					ID:                   &feedID,
					Title:                "For all mankind",
					RssSourceURL:         "https://rss.site.fr",
					ParentDirID:          998868232,
					DeleteOldFiles:       false,
					DontProcessWholeFeed: false,
					Keyword:              "FOR.ALL.MANKIND&S03&2160P&FRATERNITY",
					UnwantedKeywords:     "",
					Paused:               false,
					Extract:              false,
					FailedItemCount:      0,
					LastError:            "",
					LastFetch:            Time{time.Date(2022, time.September, 11, 19, 46, 39, 0, time.UTC)},
					CreatedAt:            Time{time.Date(2022, time.June, 13, 00, 01, 52, 0, time.UTC)},
					PausedAt:             Time{},
					StartAt:              Time{},
					UpdatedAt:            Time{time.Date(2022, time.September, 11, 19, 46, 39, 0, time.UTC)},
				}
			}(),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &rssService{
				client: tt.fields.client,
			}
			got, err := s.Get(tt.args.ctx, tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Get() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
