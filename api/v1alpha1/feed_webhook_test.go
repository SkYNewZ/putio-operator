package v1alpha1

import (
	"context"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

func TestFeed_validateRSSSourceURL(t *testing.T) {
	type fields struct {
		TypeMeta   v1.TypeMeta
		ObjectMeta v1.ObjectMeta
		Spec       FeedSpec
		Status     FeedStatus
	}
	type args struct {
		u       string
		fldPath *field.Path
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "valid URL",
			fields: fields{},
			args: args{
				u:       "https://google.fr",
				fldPath: field.NewPath("spec").Child("rss_source_url"),
			},
			wantErr: false,
		},
		{
			name:   "invalid url",
			fields: fields{},
			args: args{
				u:       "foo bar",
				fldPath: field.NewPath("spec").Child("rss_source_url"),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Feed{
				TypeMeta:   tt.fields.TypeMeta,
				ObjectMeta: tt.fields.ObjectMeta,
				Spec:       tt.fields.Spec,
				Status:     tt.fields.Status,
			}
			if err := r.validateRSSSourceURL(tt.args.u, tt.args.fldPath); (err != nil) != tt.wantErr {
				t.Errorf("validateRSSSourceURL() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

var _ = Describe("Feed webhook", func() {
	// Define utility constants for object names and testing timeouts/durations and intervals.
	const (
		FeedName      = "test-feed"
		FeedNamespace = "default"
		timeout       = time.Second * 10
		interval      = time.Millisecond * 250
	)

	Context("When creating Feed", func() {
		It("Should validate Feed URL", func() {
			By("By giving a wrong URL")
			ctx := context.Background()
			feed := &Feed{
				TypeMeta: v1.TypeMeta{
					Kind:       "Feed",
					APIVersion: "putio.skynewz.dev/v1alpha1",
				},
				ObjectMeta: v1.ObjectMeta{
					Name:      FeedName,
					Namespace: FeedNamespace,
				},
				Spec: FeedSpec{
					Title:                "foo",
					RssSourceURL:         "foo bar",
					ParentDirID:          nil,
					DeleteOldFiles:       new(bool),
					DontProcessWholeFeed: new(bool),
					Keyword:              "foo",
					UnwantedKeywords:     "",
					Paused:               new(bool),
					AuthSecretRef: AuthSecretReference{
						Name: "putio-token",
						Key:  "token",
					},
				},
				Status: FeedStatus{},
			}

			expectedError := "admission webhook \"vfeed.kb.io\" denied the request: spec.rss_source_url: Invalid value: \"foo bar\": invalid URL provided"
			Expect(k8sClient.Create(ctx, feed)).Should(MatchError(expectedError))

			By("By giving a valid URL")
			feed = &Feed{
				TypeMeta: v1.TypeMeta{
					Kind:       "Feed",
					APIVersion: "putio.skynewz.dev/v1alpha1",
				},
				ObjectMeta: v1.ObjectMeta{
					Name:      FeedName,
					Namespace: FeedNamespace,
				},
				Spec: FeedSpec{
					Title:                "foo",
					RssSourceURL:         "https://www.google.fr",
					ParentDirID:          nil,
					DeleteOldFiles:       new(bool),
					DontProcessWholeFeed: new(bool),
					Keyword:              "foo",
					UnwantedKeywords:     "",
					Paused:               new(bool),
					AuthSecretRef: AuthSecretReference{
						Name: "putio-token",
						Key:  "token",
					},
				},
				Status: FeedStatus{},
			}

			Expect(k8sClient.Create(ctx, feed)).Should(Succeed())
			Expect(k8sClient.Delete(ctx, feed)).Should(Succeed())
		})

		It("Should set Feed default values", func() {
			ctx := context.Background()
			feed := &Feed{
				TypeMeta: v1.TypeMeta{
					Kind:       "Feed",
					APIVersion: "putio.skynewz.dev/v1alpha1",
				},
				ObjectMeta: v1.ObjectMeta{
					Name:      FeedName,
					Namespace: FeedNamespace,
				},
				// fill required sepcs
				Spec: FeedSpec{
					Title:                "foo",
					RssSourceURL:         "https://www.google.fr",
					ParentDirID:          nil,
					DeleteOldFiles:       nil,
					DontProcessWholeFeed: nil,
					Keyword:              "foo",
					UnwantedKeywords:     "",
					Paused:               nil,
					AuthSecretRef: AuthSecretReference{
						Name: "foo",
						Key:  "bar",
					},
				},
				Status: FeedStatus{},
			}

			Expect(k8sClient.Create(ctx, feed)).Should(Succeed())

			// get the created feed
			feedLookupKey := types.NamespacedName{Name: FeedName, Namespace: FeedNamespace}
			createdFeed := &Feed{}
			Expect(k8sClient.Get(ctx, feedLookupKey, createdFeed)).Should(Succeed())

			// Ensure default values
			Expect(createdFeed.Spec.ParentDirID).Should(Equal(new(uint)))
			Expect(createdFeed.Spec.DeleteOldFiles).Should(Equal(new(bool)))
			Expect(createdFeed.Spec.DontProcessWholeFeed).Should(Equal(new(bool)))
			Expect(createdFeed.Spec.Paused).Should(Equal(new(bool)))
		})
	})
})
