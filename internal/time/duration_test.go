package time

import (
	"testing"
	"time"
)

var parseDurationTests = []struct {
	in   string
	ok   bool
	want time.Duration
}{
	// simple
	{"0", true, 0},
	{"5s", true, 5 * time.Second},
	{"30s", true, 30 * time.Second},
	{"1478s", true, 1478 * time.Second},
	// sign
	{"-5s", true, -5 * time.Second},
	{"+5s", true, 5 * time.Second},
	{"-0", true, 0},
	{"+0", true, 0},
	// decimal
	{"5.0s", true, 5 * time.Second},
	{"5.6s", true, 5*time.Second + 600*time.Millisecond},
	{"5.s", true, 5 * time.Second},
	{".5s", true, 500 * time.Millisecond},
	{"1.0s", true, 1 * time.Second},
	{"1.00s", true, 1 * time.Second},
	{"1.004s", true, 1*time.Second + 4*time.Millisecond},
	{"1.0040s", true, 1*time.Second + 4*time.Millisecond},
	{"100.00100s", true, 100*time.Second + 1*time.Millisecond},
	// different units
	{"14s", true, 14 * time.Second},
	{"15m", true, 15 * time.Minute},
	{"16h", true, 16 * time.Hour},
	{"17d", true, 24 * 17 * time.Hour},
	// composite durations
	{"3h30m", true, 3*time.Hour + 30*time.Minute},
	{"10.5s4m", true, 4*time.Minute + 10*time.Second + 500*time.Millisecond},
	{"-2m3.4s", true, -(2*time.Minute + 3*time.Second + 400*time.Millisecond)},
	{"39h9m14.425s", true, 39*time.Hour + 9*time.Minute + 14*time.Second + 425*time.Millisecond},
	// more than 9 digits after decimal point, see https://golang.org/issue/6617
	{"0.3333333333333333333h", true, 20 * time.Minute},
	// huge string; issue 15011.
	{"0.100000000000000000000h", true, 6 * time.Minute},
	// This value tests the first overflow check in leadingFraction.
	{"0.830103483285477580700h", true, 49*time.Minute + 48*time.Second + 372539827*time.Nanosecond},

	// errors
	{"", false, 0},
	{"3", false, 0},
	{"-", false, 0},
	{"s", false, 0},
	{".", false, 0},
	{"-.", false, 0},
	{".s", false, 0},
	{"+.s", false, 0},
}

func TestParseDuration(t *testing.T) {
	for _, tc := range parseDurationTests {
		d, err := ParseDuration(tc.in)
		if tc.ok && (err != nil || d != tc.want) {
			t.Errorf("ParseDuration(%q) = %v, %v, want %v, nil", tc.in, d, err, tc.want)
		} else if !tc.ok && err == nil {
			t.Errorf("ParseDuration(%q) = _, nil, want _, non-nil", tc.in)
		}
	}
}
