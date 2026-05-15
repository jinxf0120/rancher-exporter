package rancher

import (
	"reflect"
	"testing"
)

func TestClient_GetBackups(t *testing.T) {
	tests := []struct {
		name    string
		fields  fields
		want    []backup
		wantErr bool
	}{
		{"test-1", testClient, []backup{}, false}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := Client{
				Client: tt.fields.Client,
				Config: tt.fields.Config,
			}
			got, err := r.GetBackups()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetBackups() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if reflect.TypeOf(tt.want) != reflect.TypeOf(got) {
				t.Errorf("GetBackups() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_GetNumberOfBackups(t *testing.T) {
	tests := []struct {
		name    string
		fields  fields
		want    int
		wantErr bool
	}{
		{"test-1", testClient, 1, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := Client{
				Client: tt.fields.Client,
				Config: tt.fields.Config,
			}
			got, err := r.GetNumberOfBackups()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetNumberOfBackups() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetNumberOfBackups() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_GetNumberOfRestores(t *testing.T) {
	tests := []struct {
		name    string
		fields  fields
		want    int
		wantErr bool
	}{
		{"test-1", testClient, 1, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := Client{
				Client: tt.fields.Client,
				Config: tt.fields.Config,
			}
			got, err := r.GetNumberOfRestores()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetNumberOfRestores() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetNumberOfRestores() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_GetRestores(t *testing.T) {
	tests := []struct {
		name    string
		fields  fields
		want    []restore
		wantErr bool
	}{
		{"test-1", testClient, []restore{}, false}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := Client{
				Client: tt.fields.Client,
				Config: tt.fields.Config,
			}
			got, err := r.GetRestores()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetRestores() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if reflect.TypeOf(tt.want) != reflect.TypeOf(got) {
				t.Errorf("GetRestores() got = %v, want %v", got, tt.want)
			}
		})
	}
}
