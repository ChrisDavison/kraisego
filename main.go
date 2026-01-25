package main

import (
	"flag"
	"log"
	"os"
	"os/exec"
	"strings"
)

var logFile = os.ExpandEnv("$HOME/.kraise.log")

type parsedArgs struct {
	wmclass      string
	title        string
	excludeTitle string
	runCmd       string
}

func parseArgs() parsedArgs {
	var args parsedArgs

	flag.StringVar(&args.wmclass, "c", "", "window class")
	flag.StringVar(&args.wmclass, "wmclass", "", "window class")
	flag.StringVar(&args.title, "t", "", "window title")
	flag.StringVar(&args.title, "title", "", "window title")
	flag.StringVar(&args.excludeTitle, "e", "", "exclude title")
	flag.StringVar(&args.excludeTitle, "exclude-title", "", "exclude title")
	flag.StringVar(&args.runCmd, "l", "", "run command if no match")
	flag.StringVar(&args.runCmd, "run", "", "run command if no match")
	flag.Parse()

	return args
}

func kdo(args ...string) ([]string, error) {
	cmd := exec.Command("kdotool", args...)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return []string{}, nil
	}
	return lines, nil
}

func setIntersection(a, b []string) []string {
	set := make(map[string]struct{})
	for _, v := range a {
		set[v] = struct{}{}
	}
	var result []string
	for _, v := range b {
		if _, ok := set[v]; ok {
			result = append(result, v)
		}
	}
	return result
}

func setDifference(a, b []string) []string {
	set := make(map[string]struct{})
	for _, v := range b {
		set[v] = struct{}{}
	}
	var result []string
	for _, v := range a {
		if _, ok := set[v]; !ok {
			result = append(result, v)
		}
	}
	return result
}

func main() {
	args := parseArgs()

	activeWindow, err := kdo("getactivewindow")
	if err != nil || len(activeWindow) == 0 {
		log.Fatalf("Failed to get active window: %v", err)
	}
	active := activeWindow[0]

	allWindows, err := kdo("search")
	if err != nil {
		log.Fatalf("Failed to list windows: %v", err)
	}

	if args.wmclass != "" {
		classWindows, err := kdo("search", "--class", args.wmclass)
		if err != nil {
			log.Fatalf("Failed to search by class: %v", err)
		}
		allWindows = setIntersection(allWindows, classWindows)
	}

	if args.title != "" {
		titleWindows, err := kdo("search", "--name", args.title)
		if err != nil {
			log.Fatalf("Failed to search by title: %v", err)
		}
		allWindows = setIntersection(allWindows, titleWindows)
	}

	var excludeWindows []string
	if args.excludeTitle != "" {
		excludeWindows, err = kdo("search", "--name", args.excludeTitle)
		if err != nil {
			log.Fatalf("Failed to search exclude title: %v", err)
		}
	}

	remaining := setDifference(allWindows, excludeWindows)

	// Logging
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer f.Close()
	logger := log.New(f, "", 0)
	logger.Printf("\nactive_window=%s\nwmclass=%s\ntitle=%s\nexclude_title=%s\nrun=%s\n", active, args.wmclass, args.title, args.excludeTitle, args.runCmd)
	logger.Printf("remaining=%v\n", remaining)

	if len(remaining) > 0 {
		// Check if active window is in remaining
		found := -1
		for i, w := range remaining {
			if w == active {
				found = i
				break
			}
		}
		if found >= 0 {
			// Cycle to next window
			next := (found + 1) % len(remaining)
			cmd := exec.Command("kdotool", "windowactivate", remaining[next])
			if err := cmd.Run(); err != nil {
				log.Fatalf("Failed to activate window: %v", err)
			}
		} else {
			// Activate first match
			cmd := exec.Command("kdotool", "windowactivate", remaining[0])
			if err := cmd.Run(); err != nil {
				log.Fatalf("Failed to activate window: %v", err)
			}
		}
	} else {
		if args.runCmd != "" {
			parts := strings.Fields(args.runCmd)
			if len(parts) == 0 {
				return
			}
			cmd := exec.Command(parts[0], parts[1:]...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				log.Fatalf("Failed to run command: %v", err)
			}
		} else {
			logger.Println("No match")
		}
	}
}
