package schedule

import (
	"testing"
	"time"
)

func Test_resolveCron(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name    string
		args    args
		want    time.Duration
		wantErr bool
	}{
		{
			name:    "2023 year, everyday  8:15 am.",
			args:    args{str: "0 15 8 ? * * 2023"},
			want:    time.Date(2023, 1, 1, 8, 15, 0, 0, time.Local).Sub(time.Now()),
			wantErr: false,
		},
		{
			name:    "2021 year,everyday 8:15 am.",
			args:    args{"0 15 8 ? * * 2021"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveCron(tt.args.str)
			if (err != nil) != tt.wantErr {
				t.Errorf("resolveCron() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got.Seconds() != tt.want.Seconds() {
				t.Errorf("resolveCron() got = %v, want %v", got.Seconds(), tt.want.Seconds()) // only check second.
			}
		})
	}
}
