package rancher

import (
	"reflect"
	"testing"
)

func TestClient_GetNumberofProjects(t *testing.T) {
	tests := []struct {
		name    string
		fields  fields
		want    int
		wantErr bool
	}{
		{"test-1", testClient, 7, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := Client{
				Client: tt.fields.Client,
				Config: tt.fields.Config,
			}
			got, err := r.GetNumberofProjects()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetNumberofProjects() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetNumberofProjects() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_GetProjectAnnotations(t *testing.T) {
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
			got, err := r.GetProjectAnnotations()

			if (err != nil) != tt.wantErr {
				t.Errorf("GetProjectAnnotations() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) < tt.want {
				t.Errorf("GetProjectAnnotations() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_GetProjectLabels(t *testing.T) {
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
			got, err := r.GetProjectLabels()
			// 打印返回值
			t.Logf("Total labels found: %d", len(got))

			for i, label := range got {
				t.Logf("Label[%d]: %+v", i, label)
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("GetProjectLabels() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) < tt.want {
				t.Errorf("GetProjectAnnotations() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_GetProjectResourceQuota(t *testing.T) {
	tests := []struct {
		name    string
		fields  fields
		want    []projectResource
		wantErr bool
	}{
		{"test-1", testClient, []projectResource{{"ci-project", "ci-project",
			"local", "configMaps", 1000, "hard"}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := Client{
				Client: tt.fields.Client,
				Config: tt.fields.Config,
			}
			got, err := r.GetProjectResourceQuota()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetProjectResourceQuota() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetProjectResourceQuota() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_clusterIdToName(t *testing.T) {
	tests := []struct {
		name    string
		fields  fields
		args    string
		want    string
		wantErr bool
	}{
		{"test-1", testClient, "c-m-pskdut5m", "fake-cluster", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := Client{
				Client: tt.fields.Client,
				Config: tt.fields.Config,
			}
			got, err := r.clusterIdToName(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("clusterIdToName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("clusterIdToName() got = %v, want %v", got, tt.want)
			}
		})
	}
}
