// main.go -- rr-audit-chain: verifier for the guards' hash-chained audit logs.
//
// Every rr-*-guard audit log is JSONL where each line carries "prev_hash":
// the SHA-256 (64 lowercase hex chars) of the raw bytes of the previous
// line INCLUDING its trailing newline; the first line of a log carries 64
// zeros (RECORD-IMMUTABILITY-TECH-001 section 9.1, bound by the Security
// Plane Wave 0 spec section 11.3). The chain is over raw line bytes, not
// parsed records, so this one verifier serves all six guard schemas.
//
//	rr-audit-chain verify <log-path>
//	rr-audit-chain head <log-path>
//
// verify re-walks the log -- for verify-guard style rotation, the segments
// <log>.3, <log>.2, <log>.1 and then the active file, in that order --
// recomputes every link, reports the first broken one, and prints the chain
// head (SHA-256 of the final line, newline included). head prints only that
// hash.
//
// Exit codes: 0 = chain intact; 1 = chain broken (hash mismatch, missing or
// malformed prev_hash); 2 = usage or I/O error (no verdict possible).
//
// Honest limit (TECH-001 9.4): the chain proves internal consistency only.
// Deletion or truncation of the log's tail is invisible unless the head
// hash is anchored elsewhere.
package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

const (
	exitIntact = 0
	exitBroken = 1
	exitError  = 2

	// genesisPrevHash is the prev_hash carried by the first line of a log.
	genesisPrevHash = "0000000000000000000000000000000000000000000000000000000000000000"
)

func main() {
	os.Exit(run(os.Args, os.Stdout))
}

func run(argv []string, out io.Writer) int {
	if len(argv) != 3 {
		fmt.Fprintf(out, "usage: %s verify|head <log-path>\n", progName(argv))
		return exitError
	}
	switch argv[1] {
	case "verify":
		return runVerify(argv[2], out)
	case "head":
		return runHead(argv[2], out)
	default:
		fmt.Fprintf(out, "usage: %s verify|head <log-path>\n", progName(argv))
		return exitError
	}
}

func progName(argv []string) string {
	if len(argv) == 0 || argv[0] == "" {
		return "rr-audit-chain"
	}
	return argv[0]
}

// brokenError marks a chain verdict (exit 1) as opposed to an I/O failure
// (exit 2).
type brokenError struct{ msg string }

func (e *brokenError) Error() string { return e.msg }

// logSegments returns the log's segments in chain order: rotated backups
// .3, .2, .1 (oldest first) followed by the active file. droppedHistory
// reports whether backup .3 exists; if it does, rotation may already have
// deleted older segments, so the first retained line is exempt from the
// genesis-zeros requirement -- its predecessor is unverifiable.
func logSegments(path string) (segments []string, droppedHistory bool, err error) {
	for i := 3; i >= 1; i-- {
		seg := path + "." + strconv.Itoa(i)
		if _, statErr := os.Stat(seg); statErr == nil {
			segments = append(segments, seg)
			if i == 3 {
				droppedHistory = true
			}
		} else if !os.IsNotExist(statErr) {
			return nil, false, statErr
		}
	}
	if _, statErr := os.Stat(path); statErr == nil {
		segments = append(segments, path)
	} else if !os.IsNotExist(statErr) {
		return nil, false, statErr
	}
	if len(segments) == 0 {
		return nil, false, fmt.Errorf("no such audit log: %s", path)
	}
	return segments, droppedHistory, nil
}

// chainWalker recomputes prev_hash links across the walked segments.
type chainWalker struct {
	prev     []byte // raw bytes of the previous line, newline included
	prevFile string // segment the previous line was read from
	prevLine int    // line number of the previous line within its segment
	seen     int    // total lines walked
	head     string // SHA-256 hex of the most recent line
	genesis  bool   // still expecting the very first line
	exempt   bool   // first line exempt from genesis-zeros (dropped history)
}

// check verifies one raw line against the chain. raw must be the exact
// bytes read from the file, trailing newline included when present.
func (w *chainWalker) check(raw []byte, file string, lineNo int) error {
	var parsed map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSuffix(string(raw), "\n")), &parsed); err != nil {
		return &brokenError{fmt.Sprintf("%s line %d: unparseable JSON: %v", file, lineNo, err)}
	}
	value, ok := parsed["prev_hash"].(string)
	if !ok {
		return &brokenError{fmt.Sprintf("%s line %d: missing prev_hash", file, lineNo)}
	}
	if !isHex64(value) {
		return &brokenError{fmt.Sprintf("%s line %d: malformed prev_hash %q (want 64 lowercase hex)", file, lineNo, value)}
	}
	sum := sha256.Sum256(raw)
	w.head = hex.EncodeToString(sum[:])
	w.seen++
	if w.genesis {
		w.genesis = false
		if !w.exempt && value != genesisPrevHash {
			return &brokenError{fmt.Sprintf("%s line %d: first line prev_hash is %s, want genesis %s", file, lineNo, value, genesisPrevHash)}
		}
	} else {
		prevSum := sha256.Sum256(w.prev)
		if computed := hex.EncodeToString(prevSum[:]); computed != value {
			// Attribute the break to the earlier line: its raw bytes no
			// longer match what the following line recorded for them.
			// (Editing a line's content is the common tamper; editing the
			// next line's prev_hash produces the same signature -- the two
			// cannot be distinguished from the log alone.)
			return &brokenError{fmt.Sprintf("%s line %d: raw bytes hash to %s but %s line %d records prev_hash %s (one of the two lines was altered)",
				w.prevFile, w.prevLine, computed, file, lineNo, value)}
		}
	}
	w.prev = append(w.prev[:0], raw...)
	w.prevFile = file
	w.prevLine = lineNo
	return nil
}

// walkFile feeds every line of one segment through the walker, preserving
// the raw line bytes (ReadBytes keeps the trailing newline; a final
// unterminated line is still a line).
func (w *chainWalker) walkFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	reader := bufio.NewReader(f)
	lineNo := 0
	for {
		raw, err := reader.ReadBytes('\n')
		if len(raw) > 0 {
			lineNo++
			if berr := w.check(raw, path, lineNo); berr != nil {
				return berr
			}
		}
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
	}
}

func isHex64(s string) bool {
	if len(s) != 64 {
		return false
	}
	for _, c := range s {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			return false
		}
	}
	return true
}

func runVerify(path string, out io.Writer) int {
	segments, droppedHistory, err := logSegments(path)
	if err != nil {
		fmt.Fprintf(out, "ERROR %v\n", err)
		return exitError
	}
	w := &chainWalker{genesis: true, exempt: droppedHistory}
	for _, seg := range segments {
		if werr := w.walkFile(seg); werr != nil {
			if berr, ok := werr.(*brokenError); ok {
				fmt.Fprintf(out, "BROKEN %s\n", berr.msg)
				return exitBroken
			}
			fmt.Fprintf(out, "ERROR %s: %v\n", seg, werr)
			return exitError
		}
	}
	if w.seen == 0 {
		fmt.Fprintf(out, "ERROR %s: log is empty\n", path)
		return exitError
	}
	fmt.Fprintf(out, "OK %s: chain intact, %d lines, %d segment(s), head %s\n",
		path, w.seen, len(segments), w.head)
	return exitIntact
}

func runHead(path string, out io.Writer) int {
	segments, _, err := logSegments(path)
	if err != nil {
		fmt.Fprintf(out, "ERROR %v\n", err)
		return exitError
	}
	last, err := lastRawLine(segments[len(segments)-1])
	if err != nil {
		fmt.Fprintf(out, "ERROR %s: %v\n", segments[len(segments)-1], err)
		return exitError
	}
	if len(last) == 0 {
		fmt.Fprintf(out, "ERROR %s: log is empty\n", segments[len(segments)-1])
		return exitError
	}
	sum := sha256.Sum256(last)
	fmt.Fprintln(out, hex.EncodeToString(sum[:]))
	return exitIntact
}

// lastRawLine returns the raw bytes of the final line of one file,
// trailing newline included, in O(line) memory.
func lastRawLine(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var last []byte
	reader := bufio.NewReader(f)
	for {
		raw, err := reader.ReadBytes('\n')
		if len(raw) > 0 {
			last = append(last[:0], raw...)
		}
		if err == io.EOF {
			return last, nil
		}
		if err != nil {
			return nil, err
		}
	}
}
