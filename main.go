package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

// goreleaser will handle replacing version at build time
var version = "dev"

func main() {
	jsonOut := flag.Bool("json", false, "prints report in json format")
	showVersion := flag.Bool("version", false, "prints version and exits")
	procDir := flag.String("proc", "/proc", "where to read /proc from (e.g. /host/proc inside a container)")
	flag.Parse()

	if *showVersion {

		fmt.Println("nodecheck", version)
		return
	}

	report := collect(*procDir)

	if *jsonOut {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(report); err != nil {

			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
	} else {
		printText(report)
	}

	if len(report.Warnings) > 0 {
		os.Exit(1)
	}
}

func printText(r Report) {

	fmt.Printf("Host: %s\n", r.Hostname)
	fmt.Printf("Load: %.2f (1 min) on %d CPUs\n", r.Load1, r.CPUs)
	fmt.Printf("Memory:  %d MB available of %d MB\n", r.MemAvailMB, r.MemTotalMB)

	if r.MemLimitMB > 0 {
		fmt.Printf("Limit:   %d MB (container memory limit)\n", r.MemLimitMB)
	}
	
	fmt.Printf("Disk /:  %.1f GB free of %.1f GB\n", r.DiskFreeGB, r.DiskTotalGB)

	if len(r.GPUs) == 0 {
		fmt.Println("GPUs:    none detected")
	} else {
		for i, g := range r.GPUs {
			fmt.Printf("GPU %d:   %s\n", i, g)
		}
	}
	if len(r.Warnings) == 0 {
		fmt.Println("Status:  OK")
	} else {
		for _, w := range r.Warnings {
			fmt.Println("WARNING:", w)
		}
	}
}
