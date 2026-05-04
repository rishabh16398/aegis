package scanner

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"
)

// clamavClient talks to the clamd daemon over TCP.
// Install ClamAV: brew install clamav
// Start daemon:   brew services start clamav   (or: clamd)
type clamavClient struct {
	addr string // default: "localhost:3310"
}

func newClamAVClient(addr string) *clamavClient {
	if addr == "" {
		addr = "localhost:3310"
	}
	return &clamavClient{addr: addr}
}

type clamResult struct {
	Clean     bool
	ThreatName string
}

// Ping checks whether clamd is reachable.
func (c *clamavClient) Ping() error {
	conn, err := net.DialTimeout("tcp", c.addr, 3*time.Second)
	if err != nil {
		return fmt.Errorf("clamd unreachable at %s: %w", c.addr, err)
	}
	defer conn.Close()

	fmt.Fprint(conn, "nPING\n")

	sc := bufio.NewScanner(conn)
	if sc.Scan() && sc.Text() == "PONG" {
		return nil
	}
	return fmt.Errorf("clamd unexpected ping response")
}

// ScanFile asks clamd to scan one file and returns the result.
// clamd protocol:
//   send →  "nSCAN /path/to/file\n"
//   recv ←  "/path/to/file: OK"         (clean)
//          "/path/to/file: Eicar FOUND"  (threat)
//          "/path/to/file: ERROR ..."    (error)
func (c *clamavClient) ScanFile(path string) (clamResult, error) {
	conn, err := net.DialTimeout("tcp", c.addr, 5*time.Second)
	if err != nil {
		return clamResult{}, fmt.Errorf("clamd dial: %w", err)
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(30 * time.Second))

	fmt.Fprintf(conn, "nSCAN %s\n", path)

	sc := bufio.NewScanner(conn)
	if !sc.Scan() {
		return clamResult{}, fmt.Errorf("clamd no response")
	}
	line := sc.Text()

	switch {
	case strings.HasSuffix(line, " OK"):
		return clamResult{Clean: true}, nil

	case strings.HasSuffix(line, " FOUND"):
		// Format: "/path: VirusName FOUND"
		parts := strings.SplitN(line, ": ", 2)
		name := strings.TrimSuffix(parts[len(parts)-1], " FOUND")
		return clamResult{Clean: false, ThreatName: name}, nil

	default:
		return clamResult{}, fmt.Errorf("clamd: %s", line)
	}
}
