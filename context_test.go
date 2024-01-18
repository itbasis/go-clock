package clock_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/itbasis/go-clock/v2"
	"github.com/itbasis/go-clock/v2/pkg"
)

func TestFromContext(t *testing.T) {
	tests := []struct {
		name string
		ctx  context.Context
		want pkg.Clock
	}{
		{
			name: "empty",
			ctx:  context.Background(),
			want: clock.New(),
		},
		{
			name: "real",
			ctx:  clock.WithContext(context.Background(), clock.New()),
			want: clock.New(),
		},
		{
			name: "mock",
			ctx:  clock.WithContext(context.Background(), clock.NewMock()),
			want: clock.NewMock(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := clock.FromContext(tt.ctx); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FromContext() = %v, want %v", got, tt.want)
			}
		})
	}
}
