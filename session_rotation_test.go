package main

import "testing"

func TestDeliveryJobRequiresSessionRotationInputs(t *testing.T) {
	tests := []struct {
		name string
		job  DeliveryJob
		want bool
	}{
		{"complete delivery", DeliveryJob{"asset-17", "subscriber-4", "session-9"}, true},
		{"missing session", DeliveryJob{"asset-17", "subscriber-4", ""}, false},
		{"missing asset", DeliveryJob{"", "subscriber-4", "session-9"}, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.job.AssetID != "" && test.job.SubscriberID != "" && test.job.SessionID != ""
			if got != test.want {
				t.Fatalf("valid=%v, want %v", got, test.want)
			}
		})
	}
}
