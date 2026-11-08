package main

import "testing"

func TestParseLoadAvg(t *testing.T) {
	got, err := parseLoadAvg("0.52 0.58 0.59 1/389 12345\n")

	if err != nil {
		t.Fatal(err)
	}
	if got != 0.52 {
		t.Errorf("got %v, want 0.52", got)
	}
}

func TestParseLoadAvgEmpty(t *testing.T) {

	if _, err := parseLoadAvg(""); err == nil {

		t.Error("expected an error for empty input")
	}
}

func TestParseMemInfo(t *testing.T) {

	input := "MemTotal:       16318480 kB\nMemFree:         1234567 kB\nMemAvailable:    8000000 kB\n"
	total, avail, err := parseMemInfo(input)

	if err != nil {
		t.Fatal(err)
	}
	
	if total != 16318480 || avail != 8000000 {
		t.Errorf("got total=%d avail=%d", total, avail)
	}
}

func TestEvaluate(t *testing.T) {
	tests := []struct {
		name string
		r    Report
		want int // number of warnings
	}{
		{"healthy", Report{CPUs: 4, Load1: 1, MemTotalMB: 1000, MemAvailMB: 500, DiskTotalGB: 100, DiskFreeGB: 50}, 0},
		{"overloaded", Report{CPUs: 2, Load1: 5, MemTotalMB: 1000, MemAvailMB: 500, DiskTotalGB: 100, DiskFreeGB: 50}, 1},
		{"everything bad", Report{CPUs: 2, Load1: 5, MemTotalMB: 1000, MemAvailMB: 50, DiskTotalGB: 100, DiskFreeGB: 5}, 3},
	}

	for _, tc := range tests {

		t.Run(tc.name, func(t *testing.T) {
			if got := len(evaluate(tc.r)); got != tc.want {

				t.Errorf("got %d warnings, want %d", got, tc.want)
			}
		})
	}
}