package rancher

import (
	"testing"
)

func TestClient_GetClusterLabels(t *testing.T) {
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
			got, err := r.GetClusterLabels()
			// 打印返回值
			t.Logf("Total labels found: %d", len(got))

			for i, label := range got {
				t.Logf("Label[%d]: %+v", i, label)
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("GetClusterLabels() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) < tt.want {
				t.Errorf("GetProjectAnnotations() got = %v, want %v", got, tt.want)
			}
		})
	}
}
