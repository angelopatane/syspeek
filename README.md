# 🛠️ SysPeek — Lightweight System Monitoring CLI Tool

SysPeek is a simple and elegant terminal-based system monitor built in Go. It gives you essential real-time information about your system in a refreshingly clean and emoji-enhanced interface.

## ✨ Features

- 🧠 System Information (Hostname, OS, Architecture, Kernel)
- ⚙️ CPU Usage
- 💾 Memory Usage
- 🌐 Network I/O
- 💽 Disk Usage
- ⏱️ Uptime
- 🔁 Refresh Mode (e.g. every 2 seconds)
- 📄 Optional Log File Output
- ❌ Graceful Exit with Summary Report
- 🧾 Help and Version Flags
- 📦 Single Binary — No Dependencies

## 1. Getting Started

### Prerequisites

- Go 1.22+ installed
- macOS or Linux (Windows supported but best tested on Unix-like systems)

### Clone the Repo

```bash
git clone https://github.com/YOUR_USERNAME/syspeek.git
cd syspeek
```

## 2. Build the Binary

```bash
go build -o syspeek main.go
```

Now run the tool with:

```bash
./syspeek --all --refresh=2
```

## 3. Flags and Usage

| Flag         | Description                        |
|--------------|------------------------------------|
| `--cpu`      | Show CPU usage                     |
| `--mem`      | Show memory usage                  |
| `--net`      | Show network I/O                   |
| `--disk`     | Show disk usage                    |
| `--all`      | Show all metrics                   |
| `--refresh`  | Refresh interval in seconds        |
| `--log`      | Log output to specified file       |
| `--version`  | Display version and exit           |
| `--help`     | Show help message                  |

### Example

```bash
./syspeek --cpu --mem --refresh=3 --log=output.log
```

## 4. Logging

SysPeek supports logging system output to a specified file via the `--log` flag. This is great for audits, diagnostics, or simply saving a snapshot.

Example:

```bash
./syspeek --all --refresh=5 --log=system_report.log
```

## 5. Exit Message

When stopped (e.g. with Ctrl+C), the program prints a summary including total runtime and number of refreshes.

## 6. Manual Page (Optional)

A basic man page is available via:

```bash
man ./syspeek.1
```

To install it globally (advanced users):

```bash
sudo cp syspeek.1 /usr/local/share/man/man1/
man syspeek
```

## 7. Future Plans

- Temperature readings 🌡️
- Battery status 🔋
- Export in JSON/CSV formats
- Remote system monitoring

## 📄 License

This project is licensed under the MIT License — see the [LICENSE](https://chatgpt.com/c/LICENSE) for details.

---

> Built with ❤️ by Angelo Patane (2025) — A mechanical engineer with a passion for open source, Linux, and Golang development.
