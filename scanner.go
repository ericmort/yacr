// +build ignore

package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

// Scanner provides an interface for reading CSV data
// (compatible with rfc4180 and extended with the option of having a separator other than ",").
// Successive calls to the Scan method will step through the 'fields', skipping the separator/newline between the fields.
// The EndOfRecord method tells when a field is terminated by a line break.
type Scanner struct {
	*bufio.Scanner
	sep    byte // values separator
	quoted bool // specify if values may be quoted (when they contains separator or newline)
	eor    bool // true when the most recent field has been terminated by a newline (not a separator).
	line   int  // current line number (not record number)
}

// NewScanner returns a new CSV scanner to read from r.
func NewScanner(r io.Reader, sep byte, quoted bool) *Scanner {
	s := &Scanner{bufio.NewScanner(r), sep, quoted, false, 1}
	s.Split(s.scanField)
	return s
}

// EndOfRecord returns true when the most recent field has been terminated by a newline (not a separator).
func (s *Scanner) EndOfRecord() bool {
	return s.eor
}

// Lexing adapted from csv_read_one_field function in SQLite3 shell sources.
func (s *Scanner) scanField(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if s.quoted && len(data) > 0 && data[0] == '"' { // quoted field (may contains separator, newline and escaped quote)
		startLine := s.line
		escapedQuotes := 0
		var c, pc byte
		// Scan until the separator or newline following the closing quote (and ignore escaped quote)
		for i := 1; i < len(data); i++ {
			c = data[i]
			if c == '\n' {
				s.line++
			} else if c == '"' {
				if pc == c { // escaped quote
					pc = 0
					escapedQuotes++
					continue
				}
			}
			if pc == '"' && (c == s.sep || c == '\n') {
				s.eor = c == '\n'
				return i + 1, unescapeQuotes(data[1:i-1], escapedQuotes), nil
			} else if c == '\n' && pc == '\r' && i > 2 && data[i-2] == '"' {
				s.eor = true
				return i + 1, unescapeQuotes(data[1:i-2], escapedQuotes), nil
			}
			if pc == '"' && c != '\r' {
				return 0, nil, fmt.Errorf("unescaped %c character at line %d", pc, s.line)
			}
			pc = c
		}
		// If we're at EOF, we have a non-terminated field.
		if atEOF {
			return 0, nil, fmt.Errorf("non-terminated quoted field at line %d", startLine)
		}
	} else { // non-quoted field
		// Scan until separator or newline, marking end of field.
		for i, c := range data {
			if c == s.sep {
				s.eor = false
				return i + 1, data[0:i], nil
			} else if c == '\n' {
				s.eor = true
				s.line++
				if i > 0 && data[i-1] == '\r' {
					return i + 1, data[0 : i-1], nil
				}
				return i + 1, data[0:i], nil
			}
		}
		// If we're at EOF, we have a final, non-empty field. Return it.
		if atEOF {
			s.eor = true
			return len(data), data, nil
		}
	}
	// Request more data.
	return 0, nil, nil
}

func unescapeQuotes(b []byte, count int) []byte {
	if count == 0 {
		return b
	}
	for i, j := 0, 0; i < len(b); i, j = i+1, j+1 {
		b[j] = b[i]
		if b[i] == '"' {
			i++
		}
	}
	return b[:len(b)-count]
}

func main() {
	s := NewScanner(os.Stdin, '\t', false)
	/*null, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		panic(err)
	}
	defer null.Close()*/

	for s.Scan() {
		//null.Write(s.Bytes()) // TODO error handling
		if s.EndOfRecord() {
			//null.WriteString("\n") // TODO error handling
		} else {
			//null.WriteString("|") // TODO error handling
		}
	}
	if err := s.Err(); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}