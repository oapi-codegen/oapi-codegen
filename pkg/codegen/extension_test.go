package codegen

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_extTypeName(t *testing.T) {
	type args struct {
		extPropValue json.RawMessage
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "success",
			args:    args{json.RawMessage(`"uint64"`)},
			want:    "uint64",
			wantErr: false,
		},
		{
			name:    "nil conversion error",
			args:    args{nil},
			want:    "",
			wantErr: true,
		},
		{
			name:    "type conversion error",
			args:    args{json.RawMessage(`12`)},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// kin-openapi no longer returns these as RawMessage
			var extPropValue interface{}
			if tt.args.extPropValue != nil {
				err := json.Unmarshal(tt.args.extPropValue, &extPropValue)
				assert.NoError(t, err)
			}
			got, err := extTypeName(extPropValue)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_extParsePropGoTypeSkipOptionalPointer(t *testing.T) {
	type args struct {
		extPropValue json.RawMessage
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name:    "success when set to true",
			args:    args{json.RawMessage(`true`)},
			want:    true,
			wantErr: false,
		},
		{
			name:    "success when set to false",
			args:    args{json.RawMessage(`false`)},
			want:    false,
			wantErr: false,
		},
		{
			name:    "nil conversion error",
			args:    args{nil},
			want:    false,
			wantErr: true,
		},
		{
			name:    "type conversion error",
			args:    args{json.RawMessage(`"true"`)},
			want:    false,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// kin-openapi no longer returns these as RawMessage
			var extPropValue interface{}
			if tt.args.extPropValue != nil {
				err := json.Unmarshal(tt.args.extPropValue, &extPropValue)
				assert.NoError(t, err)
			}
			got, err := extParsePropGoTypeSkipOptionalPointer(extPropValue)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
