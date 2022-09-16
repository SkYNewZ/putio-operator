package controllers

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	skynewzdevv1alpha1 "github.com/SkYNewZ/putio-operator/api/v1alpha1"
	"github.com/SkYNewZ/putio-operator/internal/putio"
	"github.com/google/go-cmp/cmp"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"golang.org/x/oauth2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
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

var _ = Describe("Feed controller", func() {
	// Define utility constants for object names and testing timeouts/durations and intervals.
	const (
		FeedName        = "test-feed"
		FeedNamespace   = "default"
		SecretName      = "putio-token-test"
		SecretKeyName   = "token"
		SecretNamespace = FeedNamespace

		timeout  = time.Second * 10
		interval = time.Millisecond * 250
	)

	Context("When creating a feed", func() {
		It("Should create a Feed", func() {
			By("By Creating creating the putio secret")
			putioSecret := &corev1.Secret{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Secret",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      SecretName,
					Namespace: SecretNamespace,
				},
				Immutable: nil,
				Data:      nil,
				StringData: map[string]string{
					SecretKeyName: os.Getenv("PUTIO_TESTING_OAUTH_TOKEN"),
				},
				Type: corev1.SecretTypeOpaque,
			}
			Expect(k8sClient.Create(ctx, putioSecret)).Should(Succeed())

			By("By Creating the feed")
			feed := &skynewzdevv1alpha1.Feed{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Feed",
					APIVersion: "putio.skynewz.dev/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      FeedName,
					Namespace: FeedNamespace,
				},
				Spec: skynewzdevv1alpha1.FeedSpec{
					Title:        FeedName,
					RssSourceURL: "https://www.google.com",
					ParentDirID: func() *uint {
						v := uint(0)
						return &v
					}(),
					DeleteOldFiles:       boolToPtr(false),
					DontProcessWholeFeed: boolToPtr(false),
					Keyword:              "foo",
					UnwantedKeywords:     "",
					Paused:               boolToPtr(true),
					AuthSecretRef: skynewzdevv1alpha1.AuthSecretReference{
						Name: SecretName,
						Key:  SecretKeyName,
					},
				},
				Status: skynewzdevv1alpha1.FeedStatus{},
			}
			Expect(k8sClient.Create(ctx, feed)).Should(Succeed())

			feedLookupKey := types.NamespacedName{Name: FeedName, Namespace: FeedNamespace}
			createdFeed := &skynewzdevv1alpha1.Feed{}

			// We'll need to retry getting this newly created Feed, given that creation may not immediately happen.
			Eventually(func() bool {
				return k8sClient.Get(ctx, feedLookupKey, createdFeed) == nil
			}, timeout, interval).Should(BeTrue())

			// retry getting for the status become updated
			By("By checking feed ID in status")
			Eventually(func() (*uint, error) {
				if err := k8sClient.Get(ctx, feedLookupKey, createdFeed); err != nil {
					return nil, err
				}

				return createdFeed.Status.ID, nil
			}, timeout, interval).ShouldNot(BeNil())

			By("By checking finalizer")
			Expect(createdFeed.Annotations[finalizerAnnotation]).ShouldNot(BeNil())

			By("By checking statuses conditions")
			Expect(createdFeed.Status.Conditions).ShouldNot(BeEmpty())

			By("By checking feed on Put.io")
			f, err := getPutioFeed(createdFeed.Status.ID)
			Expect(err).Should(BeNil())
			Expect(f).ShouldNot(BeNil())
			Expect(f.Keyword).Should(Equal("foo"))

			By("By checking title")
			Expect(f.Title).Should(Equal(fmt.Sprintf(titleFormat, createdFeed.Spec.Title, createdFeed.GetGeneration())))
		})
	})

	Context("When updating a feed", func() {
		It("Should update Put.io feed from spec", func() {
			feedLookupKey := types.NamespacedName{Name: FeedName, Namespace: FeedNamespace}
			createdFeed := &skynewzdevv1alpha1.Feed{}
			Eventually(func() bool {
				return k8sClient.Get(ctx, feedLookupKey, createdFeed) == nil
			}, timeout, interval).Should(BeTrue())

			createdFeed.Spec.Keyword = "bar"
			createdFeed.Spec.UnwantedKeywords = "foo"
			Expect(k8sClient.Update(ctx, createdFeed)).Should(Succeed())

			// wait for update and check
			Eventually(func() bool {
				f, err := getPutioFeed(createdFeed.Status.ID)
				if err != nil {
					return false
				}

				if f.Keyword != "bar" {
					return false
				}

				if f.UnwantedKeywords != "foo" {
					return false
				}

				return true
			}, timeout, time.Second).Should(BeTrue())
		})
	})

	Context("When deleting a feed", func() {
		It("Should delete the feed", func() {
			feedLookupKey := types.NamespacedName{Name: FeedName, Namespace: FeedNamespace}
			createdFeed := &skynewzdevv1alpha1.Feed{}
			Expect(k8sClient.Get(ctx, feedLookupKey, createdFeed)).Should(Succeed())

			By("By deleting the feed")
			Expect(k8sClient.Delete(ctx, createdFeed)).Should(Succeed())
			Eventually(func() error {
				return k8sClient.Get(ctx, feedLookupKey, createdFeed)
			}, timeout, interval).ShouldNot(BeNil())

			By("By checking feed is deleted on Put.io")
			f, err := getPutioFeed(createdFeed.Status.ID)
			Expect(err).ShouldNot(BeNil())
			Expect(f).Should(BeNil())
		})
	})
})

func getPutioFeed(feedID *uint) (*putio.Feed, error) {
	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: os.Getenv("PUTIO_TESTING_OAUTH_TOKEN")})
	oauthClient := oauth2.NewClient(ctx, tokenSource)
	putioClient := putio.New(context.Background(), oauthClient)

	return putioClient.Rss.Get(ctx, *feedID)
}

func boolToPtr(v bool) *bool {
	return &v
}
