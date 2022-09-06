/*
Copyright 2022 Quentin Lemaire <quentin@lemairepro.fr>.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AuthSecretReference references a Secret containing a Put.io authentication token
type AuthSecretReference struct {
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`
	// +kubebuilder:validation:MinLength=1
	Key string `json:"key"`
}

// FeedSpec defines the desired state of Feed.
type FeedSpec struct {
	// +kubebuilder:validation:MinLength:=1
	// Title of the RSS feed as will appear on the site
	Title string `json:"title"`

	// +kubebuilder:validation:MinLength:=1
	// The URL of the RSS feed to be watched
	RssSourceURL string `json:"rss_source_url"`

	// +kubebuilder:default:=0
	// The file ID of the folder to place the RSS feed files in
	ParentDirID *uint32 `json:"parent_dir_id,omitempty"`

	// +kubebuilder:default:=false
	// Should old files in the folder be deleted when space is low
	DeleteOldFiles bool `json:"delete_old_files,omitempty"`

	// +kubebuilder:default:=false
	// Should the current items in the feed, at creation time, be ignored
	DontProcessWholeFeed bool `json:"dont_process_whole_feed,omitempty"`

	// +kubebuilder:validation:MinLength:=1
	// Only items with titles that contain any of these words will be transferred (comma-separated list of words)
	Keyword string `json:"keyword"`

	// +optional
	// No items with titles that contain any of these words will be transferred (comma-separated list of words)
	UnwantedKeywords string `json:"unwanted_keywords,omitempty"`

	// +kubebuilder:default:=false
	// Should the RSS feed be created in the paused state
	Paused bool `json:"paused,omitempty"`

	// Authentication reference to Put.io token in a secret
	AuthSecretRef AuthSecretReference `json:"authSecretRef"`
}

// FeedStatus defines the observed state of Feed.
type FeedStatus struct {
	ID *uint `json:"id,omitempty"`

	LastError       string       `json:"last_error,omitempty"`
	LastFetch       *metav1.Time `json:"last_fetch,omitempty"`
	FailedItemCount uint         `json:"failed_item_count,omitempty"`

	PausedAt  *metav1.Time `json:"paused_at,omitempty"`
	CreatedAt *metav1.Time `json:"created_at,omitempty"`
	UpdatedAt *metav1.Time `json:"updated_at,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Keyword",type=string,JSONPath=".spec.keyword"
// +kubebuilder:printcolumn:name="Last fetch",type=date,JSONPath=".status.last_fetch"
// +kubebuilder:printcolumn:name="Paused",type=boolean,JSONPath=".spec.paused"
// +kubebuilder:printcolumn:name="Title",type=string,JSONPath=".spec.title"
// +kubebuilder:printcolumn:name="URL",type=string,priority=1,JSONPath=".spec.rss_source_url"
// +kubebuilder:printcolumn:name="ID",type=string,priority=1,JSONPath=".status.id"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// Feed is the Schema to manage your rss feeds.
type Feed struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FeedSpec   `json:"spec,omitempty"`
	Status FeedStatus `json:"status,omitempty"`
}

func (in *Feed) AuthSecretRef() AuthSecretReference {
	return in.Spec.AuthSecretRef
}

//+kubebuilder:object:root=true

// FeedList contains a list of Feed.
type FeedList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Feed `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Feed{}, &FeedList{})
}
