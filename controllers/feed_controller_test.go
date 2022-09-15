package controllers

import (
	"context"
	"testing"

	skynewzdevv1alpha1 "github.com/SkYNewZ/putio-operator/api/v1alpha1"
	"github.com/SkYNewZ/putio-operator/internal/putio"
	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_makePutioFeedFromSpec(t *testing.T) {
	parentDirID := uint(1234)

	type args struct {
		ctx  context.Context
		feed *skynewzdevv1alpha1.Feed
	}
	tests := []struct {
		name string
		args args
		want *putio.Feed
	}{
		{
			name: "expected",
			args: args{
				ctx: context.Background(),
				feed: &skynewzdevv1alpha1.Feed{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec: skynewzdevv1alpha1.FeedSpec{
						Title:                "foo",
						RssSourceURL:         "https://www.google.com",
						ParentDirID:          &parentDirID,
						DeleteOldFiles:       boolToPtr(true),
						DontProcessWholeFeed: boolToPtr(true),
						Keyword:              "foo",
						UnwantedKeywords:     "bar",
						// Paused:               true, // Pause is not handle during creation/update
						AuthSecretRef: skynewzdevv1alpha1.AuthSecretReference{},
					},
					Status: skynewzdevv1alpha1.FeedStatus{},
				},
			},
			want: &putio.Feed{
				ID:                   nil,
				Title:                "foo|0|managed by Kubernetes/putio-operator",
				RssSourceURL:         "https://www.google.com",
				ParentDirID:          parentDirID,
				DeleteOldFiles:       true,
				DontProcessWholeFeed: true,
				Keyword:              "foo",
				UnwantedKeywords:     "bar",
				// Paused:               true, // Pause is not handle during creation/update
				Extract:         false,
				FailedItemCount: 0,
				LastError:       "",
				LastFetch:       putio.Time{},
				CreatedAt:       putio.Time{},
				PausedAt:        putio.Time{},
				StartAt:         putio.Time{},
				UpdatedAt:       putio.Time{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//if got := makePutioFeedFromSpec(tt.args.ctx, tt.args.feed); !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("makePutioFeedFromSpec() = %v, want %v", got, tt.want)
			//}

			got := makePutioFeedFromSpec(tt.args.ctx, tt.args.feed)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("makePutioFeedFromSpec() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func Test_makeFeedTitleWithGenerationNumber(t *testing.T) {
	type args struct {
		ctx  context.Context
		feed *skynewzdevv1alpha1.Feed
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "expected title",
			args: args{
				ctx: context.Background(),
				feed: &skynewzdevv1alpha1.Feed{
					TypeMeta: metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{
						Generation: 1234,
					},
					Spec: skynewzdevv1alpha1.FeedSpec{
						Title:                "foo",
						RssSourceURL:         "",
						ParentDirID:          nil,
						DeleteOldFiles:       new(bool),
						DontProcessWholeFeed: new(bool),
						Keyword:              "",
						UnwantedKeywords:     "",
						Paused:               new(bool),
						AuthSecretRef:        skynewzdevv1alpha1.AuthSecretReference{},
					},
					Status: skynewzdevv1alpha1.FeedStatus{},
				},
			},
			want: "foo|1234|managed by Kubernetes/putio-operator",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := makeFeedTitleWithGenerationNumber(tt.args.ctx, tt.args.feed); got != tt.want {
				t.Errorf("makeFeedTitleWithGenerationNumber() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isAlreadyProcessed(t *testing.T) {
	type args struct {
		ctx       context.Context
		putioFeed *putio.Feed
		feed      *skynewzdevv1alpha1.Feed
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "feed already processed",
			args: args{
				ctx: context.Background(),
				putioFeed: &putio.Feed{
					ID:                   nil,
					Title:                "foo|1234|managed by Kubernetes/putio-operator",
					RssSourceURL:         "",
					ParentDirID:          0,
					DeleteOldFiles:       false,
					DontProcessWholeFeed: false,
					Keyword:              "",
					UnwantedKeywords:     "",
					Paused:               false,
					Extract:              false,
					FailedItemCount:      0,
					LastError:            "",
					LastFetch:            putio.Time{},
					CreatedAt:            putio.Time{},
					PausedAt:             putio.Time{},
					StartAt:              putio.Time{},
					UpdatedAt:            putio.Time{},
				},
				feed: &skynewzdevv1alpha1.Feed{
					TypeMeta: metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{
						Generation: 1234,
					},
					Spec:   skynewzdevv1alpha1.FeedSpec{},
					Status: skynewzdevv1alpha1.FeedStatus{},
				},
			},
			want: true,
		},
		{
			name: "feed already processed",
			args: args{
				ctx: context.Background(),
				putioFeed: &putio.Feed{
					ID:                   nil,
					Title:                "foo|4321|managed by Kubernetes/putio-operator",
					RssSourceURL:         "",
					ParentDirID:          0,
					DeleteOldFiles:       false,
					DontProcessWholeFeed: false,
					Keyword:              "",
					UnwantedKeywords:     "",
					Paused:               false,
					Extract:              false,
					FailedItemCount:      0,
					LastError:            "",
					LastFetch:            putio.Time{},
					CreatedAt:            putio.Time{},
					PausedAt:             putio.Time{},
					StartAt:              putio.Time{},
					UpdatedAt:            putio.Time{},
				},
				feed: &skynewzdevv1alpha1.Feed{
					TypeMeta: metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{
						Generation: 1234,
					},
					Spec:   skynewzdevv1alpha1.FeedSpec{},
					Status: skynewzdevv1alpha1.FeedStatus{},
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isAlreadyProcessed(tt.args.ctx, tt.args.putioFeed, tt.args.feed); got != tt.want {
				t.Errorf("isAlreadyProcessed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func boolToPtr(v bool) *bool {
	return &v
}
