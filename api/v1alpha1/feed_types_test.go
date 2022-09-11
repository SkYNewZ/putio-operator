package v1alpha1

import (
	"reflect"
	"testing"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestFeed_AuthSecretRef(t *testing.T) {
	type fields struct {
		TypeMeta   v1.TypeMeta
		ObjectMeta v1.ObjectMeta
		Spec       FeedSpec
		Status     FeedStatus
	}
	tests := []struct {
		name   string
		fields fields
		want   AuthSecretReference
	}{
		{
			name: "expected",
			fields: fields{
				TypeMeta:   v1.TypeMeta{},
				ObjectMeta: v1.ObjectMeta{},
				Spec: FeedSpec{
					Title:                "",
					RssSourceURL:         "",
					ParentDirID:          nil,
					DeleteOldFiles:       false,
					DontProcessWholeFeed: false,
					Keyword:              "",
					UnwantedKeywords:     "",
					Paused:               false,
					AuthSecretRef: AuthSecretReference{
						Name: "foo",
						Key:  "bar",
					},
				},
				Status: FeedStatus{},
			},
			want: AuthSecretReference{
				Name: "foo",
				Key:  "bar",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := &Feed{
				TypeMeta:   tt.fields.TypeMeta,
				ObjectMeta: tt.fields.ObjectMeta,
				Spec:       tt.fields.Spec,
				Status:     tt.fields.Status,
			}
			if got := in.AuthSecretRef(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AuthSecretRef() = %v, want %v", got, tt.want)
			}
		})
	}
}
