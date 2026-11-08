package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"syscall"
)

// struct for machines attributes 

type Report struct {
	Hostname    string   `json:"hostname"`
	CPUs        int      `json:"cpus"`
	Load1       float64  `json:"load_1m"`
	MemTotalMB  uint64   `json:"mem_total_mb"`
	MemAvailMB  uint64   `json:"mem_available_mb"`
	DiskTotalGB float64  `json:"disk_total_gb"`
	DiskFreeGB  float64  `json:"disk_free_gb"`
	GPUs        []string `json:"gpus"`
	Warnings    []string `json:"warnings"`
}

func collect() Report {
	r := Report{CPUs: runtime.NumCPU()}
	r.Hostname, _ = os.Hostname()

	if data, err := os.ReadFile("/proc/loadavg"); err == nil {
		r.Load1, _ = parseLoadAvg(string(data))
	}
	if data, err := os.ReadFile("/proc/meminfo"); err == nil {
		if total, avail, err := parseMemInfo(string(data)); err == nil {
			r.MemTotalMB = total / 1024
			r.MemAvailMB = avail / 1024
		}
	}
	var fs syscall.Statfs_t
	if err := syscall.Statfs("/", &fs); err == nil {
		r.DiskTotalGB = float64(fs.Blocks*uint64(fs.Bsize)) / 1e9
		r.DiskFreeGB = float64(fs.Bavail*uint64(fs.Bsize)) / 1e9
	}
	r.GPUs = listGPUs()
	r.Warnings = evaluate(r)
	return r
}

// reads the 1-minute load from /proc/loadavg
func parseLoadAvg(s string) (float64, error) {
	fields := strings.Fields(s)
	if len(fields) == 0 {
		return 0, fmt.Errorf("empty loadavg")
	}

	return strconv.ParseFloat(fields[0], 64)
}



// returns MemTotal and MemAvailable in kB from /proc/meminfo.
func parseMemInfo(s string) (totalKB, availKB uint64, err error) {
	sc := bufio.NewScanner(strings.NewReader(s))
	for sc.Scan() {

		fields := strings.Fields(sc.Text()) // ["MemTotal:", "16318480", "kB"]
		if len(fields) < 2 {
			continue
		}
		v, perr := strconv.ParseUint(fields[1], 10, 64)
		if perr != nil {
			continue
		}
		switch fields[0] {
		case "MemTotal:":
			totalKB = v
		case "MemAvailable:":
			availKB = v
		}
	}
	


	if totalKB == 0 {
		return 0, 0, fmt.Errorf("MemTotal not found")
	}

	return totalKB, availKB, nil
}

// uses nvidia-smi for GPU names (if installed).
func listGPUs() []string {
	if _, err := exec.LookPath("nvidia-smi"); err != nil {
		return nil
	}
	out, err := exec.Command("nvidia-smi", "--query-gpu=name,memory.total", "--format=csv,noheader").Output()
	if err != nil {
		return nil
	}
	var gpus []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line = strings.TrimSpace(line); line != "" {
			gpus = append(gpus, line)
		}
	}
	return gpus
}

// filter making the raw numbers readable.
func evaluate(r Report) []string {
	var w []string
	if r.CPUs > 0 && r.Load1 > float64(r.CPUs) {
		w = append(w, fmt.Sprintf("high load: %.2f on %d CPUs", r.Load1, r.CPUs))
	}
	if r.MemTotalMB > 0 && float64(r.MemAvailMB)/float64(r.MemTotalMB) < 0.10 {
		w = append(w, "low memory: under 10% available")
	}
	if r.DiskTotalGB > 0 && r.DiskFreeGB/r.DiskTotalGB < 0.10 {
		w = append(w, "low disk: under 10% free on /")
	}
	
	return w
}