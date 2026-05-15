package rancher

import (
	"testing"
)

func TestClient_GetRancherCustomResourceCount(t *testing.T) {
	tests := []struct {
		name    string
		fields  fields
		want    int
		wantErr bool
	}{
		{"test-1", testClient, 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := Client{
				Client: tt.fields.Client,
				Config: tt.fields.Config,
			}
			got, err := r.GetRancherCustomResourceCount()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetRancherCustomResourceCount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) < tt.want {
				t.Errorf("GetRancherCustomResourceCount() got = %v, want %v", len(got), tt.want)
			}
		})
	}
}
