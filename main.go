package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/common-nighthawk/go-figure"
	"github.com/fatih/color"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
)

var (
	logFileOnly *log.Logger
)

const version = "1.0.0"

func main() {
	showCPU := flag.Bool("cpu", false, "Show CPU usage")
	showMem := flag.Bool("mem", false, "Show memory usage")
	showNet := flag.Bool("net", false, "Show network I/O")
	showAll := flag.Bool("all", false, "Show all metrics (default if no flags)")
	refresh := flag.Int("refresh", 0, "Refresh interval in seconds (e.g. --refresh=5)")
	showVersion := flag.Bool("version", false, "Print the version number and exit")
	logPath := flag.String("log", "", "Optional: Path to save output log file")
	showDisk := flag.Bool("disk", false, "Show disk usage")

	flag.Parse()

	// Step 1: Set up *temporary* logging to stdout for help check
	log.SetOutput(os.Stdout)
	log.SetFlags(0)

	// Step 2: Manually handle --help first
	if len(os.Args) == 2 && (os.Args[1] == "--help" || os.Args[1] == "-h") {
		logFilePath := *logPath // copy pointer value before anything
		var logFile *os.File
		var err error

		if logFilePath != "" {
			logFile, err = os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
			if err != nil {
				fmt.Println("⚠️ Could not open log file:", err)
				os.Exit(1)
			}
			multiWriter := io.MultiWriter(os.Stdout, logFile)
			log.SetOutput(multiWriter)
			log.Println("✅ Logging works — writing help output to log file now:")
			flag.Usage()
			logFile.Close() // manually close
		} else {
			flag.Usage()
		}

		os.Exit(0)
	}

	// Normal logging setup
	// var logFile *os.File
	// var err error
	// if *logPath != "" {
	//	logFile, err = os.OpenFile(*logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	//	if err != nil {
	//		log.Println("⚠️ Could not open log file:", err)
	//		os.Exit(1)
	//	}
	//	defer logFile.Close()
	//	multiWriter := io.MultiWriter(os.Stdout, logFile)
	//	log.SetOutput(multiWriter)
	// } else {
	//	log.SetOutput(os.Stdout)
	// }
	// log.SetFlags(0)

	var logFile *os.File
	var err error

	if *logPath != "" {
		logFile, err = os.OpenFile(*logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			fmt.Println("⚠️ Could not open log file:", err)
			os.Exit(1)
		}
		defer logFile.Close()

		// General logger outputs to both terminal AND file
		multiWriter := io.MultiWriter(os.Stdout, logFile)
		log.SetOutput(multiWriter)
		log.SetFlags(0)

		// File-only logger for exclusive log entries (section headers etc)
		logFileOnly = log.New(logFile, "", 0)

	} else {
		// No log file — general logger just outputs to terminal
		log.SetOutput(os.Stdout)
		log.SetFlags(0)

		// File-only logger discards output (no log file)
		logFileOnly = log.New(io.Discard, "", 0)
	}

	flag.Usage = func() {
		log.Println("SysPeek — A Simple System Monitoring CLI Tool 🖥️")
		log.Println()
		log.Println("Usage:")
		log.Println("  syspeek [flags]")
		log.Println()
		log.Println("Available Flags:")
		log.Println("  --cpu         Show CPU usage")
		log.Println("  --mem         Show memory usage")
		log.Println("  --net         Show network usage")
		log.Println("  --disk        Show disk usage")
		log.Println("  --all         Show all information")
		log.Println("  --refresh=N   Refresh every N seconds (e.g. --refresh=2)")
		log.Println("  --version     Print the version number and exit")
		log.Println("  --help        Show this help message and exit")
		log.Println()
		log.Println("Examples:")
		log.Println("  syspeek --all")
		log.Println("  syspeek --cpu --refresh=2")
		log.Println()
	}

	// Early exit for --version
	if len(os.Args) == 2 && os.Args[1] == "--version" {
		fmt.Println("🧪 SysPeek Version:", version) // Or your version variable
		os.Exit(0)
	}

	if *refresh > 0 && !*showCPU && !*showMem && !*showNet && !*showDisk && !*showAll {
		log.Println("⚠️ Error: Refresh is set but no data flags selected. Use --all or individual flags.")
		os.Exit(1)
	}

	if !*showCPU && !*showMem && !*showNet && !*showDisk && !*showAll {
		log.Println("⚠️ Error: Please specify at least one of --cpu, --mem, --net, --disk, or --all")
		os.Exit(1)
	}

	start := time.Now()
	refreshCount := 0

	if *showVersion {
		log.Printf("SysPeek version %s\n", version)
		return
	}

	// Handle Ctrl+C
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-signalChan
		elapsed := time.Since(start)
		log.Println("\n👋 Exiting SysPeek. Thanks for monitoring with us!")
		log.Println("=== Summary ===")
		log.Printf("⏱️  Total runtime: %s\n", elapsed.Round(time.Second))
		log.Printf("🔁 Total refreshes: %d\n", refreshCount)
		os.Exit(0)
	}()

	for {
		// Clear the screen (cross-platform-ish)
		log.Print("\033[H\033[2J")

		banner := figure.NewFigure("SysPeek", "", true)
		banner.Print()
		log.Println("🔧 Your Lightweight System Monitor")
		log.Println()

		// Show default if nothing is specified
		if !*showCPU && !*showMem && !*showNet && !*showDisk && !*showAll {
			*showAll = true
		}

		if *showAll || *showCPU || *showMem || *showNet || *showDisk {
			color.New(color.FgCyan).Println("=== System Info ===")
			logFileOnly.Println("=== System Info ===")
			printSystemInfo()
			log.Println()
		}

		if *showAll {
			color.New(color.FgCyan).Println("=== Uptime ===")
			logFileOnly.Println("=== Uptime ===")
			uptimeSeconds, err := host.Uptime()
			if err != nil {
				log.Println("⏱️  Uptime Error:", err)
			} else {
				uptimeDuration := time.Duration(uptimeSeconds) * time.Second
				log.Println("⏱️  Uptime:", formatDuration(uptimeDuration))
			}
			log.Println()
		}

		if *showAll || *showMem {
			color.New(color.FgCyan).Println("=== Memory Usage ===")
			logFileOnly.Println("=== Memory Usage ===")
			vmStats, err := mem.VirtualMemory()
			if err != nil {
				log.Println("💾 Memory Error:", err)
			} else {
				printMemoryUsage(vmStats.UsedPercent, vmStats.Total)
			}
			log.Println()
		}

		if *showAll || *showCPU {
			color.New(color.FgCyan).Println("=== CPU Usage ===")
			logFileOnly.Println("=== CPU Usage ===")
			cpuPercent, err := cpu.Percent(0, false)
			if err != nil {
				log.Println("⚙️  CPU Error:", err)
			} else if len(cpuPercent) > 0 {
				printCPUUsage(cpuPercent[0])
			}
			log.Println()
		}

		if *showAll || *showDisk {
			color.New(color.FgCyan).Println("=== Disk Usage ===")
			logFileOnly.Println("=== Disk Usage ===")
			printDiskUsage()
			log.Println()
		}

		if *showAll || *showNet {
			color.New(color.FgCyan).Println("=== Network I/O ===")
			logFileOnly.Println("=== Network Usage ===")
			netIO, err := net.IOCounters(false)
			if err != nil {
				log.Println("🌐 Network Error:", err)
			} else if len(netIO) > 0 {
				log.Printf("🌐 Network: Sent %.2f MB | Received %.2f MB\n",
					bytesToMB(netIO[0].BytesSent), bytesToMB(netIO[0].BytesRecv))
			}
			log.Println()
		}

		refreshCount++
		// If no refresh specified, break after first run
		if *refresh <= 0 {
			break
		}

		// Sleep before next refresh
		time.Sleep(time.Duration(*refresh) * time.Second)
	}

	log.Println("👋 Exiting SysPeek. Thanks for monitoring with us!")
	elapsed := time.Since(start)
	log.Printf("⏱️  Total runtime: %s\n", elapsed.Round(time.Second))
	log.Printf("🔁 Total refreshes: %d\n", refreshCount)

}

// Print system information: hostname, platform, architecture, kernel
func printSystemInfo() {
	info, err := host.Info()
	if err != nil {
		log.Println("🧠 System Info Error:", err)
		return
	}

	log.Printf("🧠 Hostname: %s\n", info.Hostname)
	log.Printf("🧬 OS: %s %s\n", info.Platform, info.PlatformVersion)
	log.Printf("🧱 Architecture: %s\n", info.KernelArch)
	log.Printf("🧩 Kernel Version: %s\n", info.KernelVersion)
	log.Println()
}

// Format seconds as "X days, Y hours, Z minutes"
func formatDuration(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	return fmt.Sprintf("%d days, %d hours, %d minutes", days, hours, minutes)
}

// Convert bytes to gigabytes
func bytesToGB(b uint64) float64 {
	return float64(b) / (1024 * 1024 * 1024)
}

// Convert bytes to megabytes
func bytesToMB(b uint64) float64 {
	return float64(b) / (1024 * 1024)
}

func printMemoryUsage(usedPercent float64, totalBytes uint64) {
	var colour *color.Color

	switch {
	case usedPercent >= 85:
		colour = color.New(color.FgRed)
	case usedPercent >= 60:
		colour = color.New(color.FgYellow)
	default:
		colour = color.New(color.FgGreen)
	}
	output := fmt.Sprintf("💾 Memory: %.2f%% used of %.2f GB\n", usedPercent, bytesToGB(totalBytes))
	//colour.Printf("💾 Memory: %.2f%% used of %.2f GB\n", usedPercent, bytesToGB(totalBytes))

	// Print with colour to terminal
	colour.Print(output)

	// Write to log (no colour codes)
	logFileOnly.Print(output)
}

func printCPUUsage(usedPercent float64) {
	var colour *color.Color

	switch {
	case usedPercent >= 85:
		colour = color.New(color.FgRed)
	case usedPercent >= 60:
		colour = color.New(color.FgYellow)
	default:
		colour = color.New(color.FgGreen)
	}
	output := fmt.Sprintf("⚙️  CPU Usage: %.2f%%\n", usedPercent)
	//colour.Printf("⚙️  CPU Usage: %.2f%%\n", usedPercent)

	// Print with colour
	colour.Print(output)

	// Log plain text
	logFileOnly.Print(output)
}

func printDiskUsage() {
	partitions, err := disk.Partitions(false)
	if err != nil {
		log.Println("💽 Disk Error:", err)
		return
	}

	for _, p := range partitions {
		usage, err := disk.Usage(p.Mountpoint)
		if err != nil {
			continue // Ignore failed mount points
		}

		usedPercent := usage.UsedPercent
		var colour *color.Color

		switch {
		case usedPercent >= 85:
			colour = color.New(color.FgRed)
		case usedPercent >= 60:
			colour = color.New(color.FgYellow)
		default:
			colour = color.New(color.FgGreen)
		}

		output := fmt.Sprintf("💽 %s — Used: %.2f%% of %.2f GB\n", usage.Path, usedPercent, bytesToGB(usage.Total))

		// Print to terminal
		colour.Print(output)

		// Log (no colour)
		logFileOnly.Print(output)
	}
}
