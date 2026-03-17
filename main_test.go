package main

import "testing"

func TestIsDaemonReloadHint(t *testing.T) {
	cases := []struct {
		name string
		text string
		want bool
	}{
		{
			name: "contains daemon-reload warning",
			text: "Warning: The unit file changed on disk. Run 'systemctl daemon-reload'.",
			want: true,
		},
		{
			name: "contains daemon-reload in different case",
			text: "RUN SYSTEMCTL DAEMON-RELOAD before restart",
			want: true,
		},
		{
			name: "no daemon-reload hint",
			text: "Failed to execute command: connection timeout",
			want: false,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if got := isDaemonReloadHint(tc.text); got != tc.want {
				t.Fatalf("isDaemonReloadHint(%q) = %v, want %v", tc.text, got, tc.want)
			}
		})
	}
}
