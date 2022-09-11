package http

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

type RoundTripFunc func(req *http.Request) *http.Response

func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

func makeExpectedResponse(t *testing.T, token string) *http.Request {
	t.Helper()

	expectedRequest := httptest.NewRequest(http.MethodGet, "https://www.google.com", nil)
	expectedRequest.Header.Set("Authorization", "Bearer "+token)
	return expectedRequest
}

func Test_transport_RoundTrip(t1 *testing.T) {
	type fields struct {
		RoundTripper http.RoundTripper
		token        string
	}
	type args struct {
		req *http.Request
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *http.Response
		wantErr bool
	}{
		{
			name: "expected token in request header",
			fields: fields{
				RoundTripper: RoundTripFunc(func(req *http.Request) *http.Response {
					return &http.Response{Request: req}
				}),
				token: "foo",
			},
			args: args{
				req: httptest.NewRequest(http.MethodGet, "https://www.google.com", nil),
			},
			want:    &http.Response{Request: makeExpectedResponse(t1, "foo")},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := transport{
				RoundTripper: tt.fields.RoundTripper,
				token:        tt.fields.token,
			}
			got, err := t.RoundTrip(tt.args.req)
			if (err != nil) != tt.wantErr {
				t1.Errorf("RoundTrip() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t1.Errorf("RoundTrip() got = %v, want %v", got, tt.want)
			}
		})
	}
}
