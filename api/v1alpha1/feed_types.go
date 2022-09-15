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

// AuthSecretReference references a Secret containing a Put.io authentication token.
type AuthSecretReference struct {
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`
	// +kubebuilder:validation:MinLength=1
	Key string `json:"key"`
}

// FeedSpec defines the desired state of Feed.
type FeedSpec struct {
	// +kubebuilder:validation:MinLength:=1
	// Title of the RSS feed as will appear on the site.
	Title string `json:"title"`

	// +kubebuilder:validation:MinLength:=1
	// The URL of the RSS feed to be watched.
	RssSourceURL string `json:"rss_source_url"`

	// The file ID of the folder to place the RSS feed files in. Default to the root directory (0).
	// +optional
	ParentDirID *uint `json:"parent_dir_id,omitempty"`

	// Should old files in the folder be deleted when space is low. Default to false.
	// +optional
	DeleteOldFiles *bool `json:"delete_old_files,omitempty"`

	// Should the current items in the feed, at creation time, be ignored.
	// +optional
	DontProcessWholeFeed *bool `json:"dont_process_whole_feed,omitempty"`

	// +kubebuilder:validation:MinLength:=1
	// Only items with titles that contain any of these words will be transferred (comma-separated list of words).
	Keyword string `json:"keyword"`

	// No items with titles that contain any of these words will be transferred (comma-separated list of words).
	// +optional
	UnwantedKeywords string `json:"unwanted_keywords,omitempty"`

	// Should the RSS feed be created in the paused state. Default to false.
	// +optional
	Paused *bool `json:"paused,omitempty"`

	// Authentication reference to Put.io token in a secret.
	AuthSecretRef AuthSecretReference `json:"authSecretRef"`
}

// FeedStatus defines the observed state of Feed.
type FeedStatus struct {
	ID *uint `json:"id,omitempty"`

	// Conditions represent the latest available observations of a Feed state
	Conditions []metav1.Condition `json:"conditions"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Keyword",type=string,JSONPath=".spec.keyword"
// +kubebuilder:printcolumn:name="Paused",type=boolean,JSONPath=".spec.paused"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="Available",type="string",JSONPath=`.status.conditions[?(@.type == "Available")].status`
// +kubebuilder:printcolumn:name="ID",type=string,priority=1,JSONPath=".status.id"
// +kubebuilder:printcolumn:name="URL",type=string,priority=1,JSONPath=".spec.rss_source_url"
// +kubebuilder:printcolumn:name="Title",type=string,priority=1,JSONPath=".spec.title"
// +kubebuilder:printcolumn:name="Last fetch",type=date,priority=1,JSONPath=".status.last_fetch"

// Feed is the Schema to manage your rss feeds.
type Feed struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FeedSpec   `json:"spec,omitempty"`
	Status FeedStatus `json:"status,omitempty"`
}

func (r *Feed) AuthSecretRef() AuthSecretReference {
	return r.Spec.AuthSecretRef
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
