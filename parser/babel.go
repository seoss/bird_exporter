package parser

import (
	"regexp"

	"bufio"
	"bytes"
	"strconv"
	"strings"

	"github.com/czerwonk/bird_exporter/protocol"
)

type babelRegex struct {
	area  *regexp.Regexp
	entry *regexp.Regexp
}

type babelContext struct {
	line    string
	entries []*protocol.BabelEntry
	areas   []*protocol.BabelEntry
	current *protocol.BabelEntry
}

func init() {
	// Match babel entries like this one:
	//
	// Prefix                   Router ID               Metric Seqno  Routes Sources
	// fc00:fe3e::1/128         00:00:00:00:00:00:00:01      0     2       0       0
	babel = &babelRegex{
		entry: regexp.MustCompile(
			"([[:xdigit:]:/]+)" +
			"[[:blank:]]+"      +
			"[[:xdigit:]:]+"  +
			"[[:blank:]]"       +
			"+([[:digit:]]+)"   +
			"[[:blank:]]+"      +
			"([[:digit:]]+)"    +
			"[[:blank:]]+"      +
			"([[:digit:]]+)"    +
			"[[:blank:]]+"      +
			"([[:digit:]]+)"),
	}
}

var babel *babelRegex

func ParseBabel(data []byte) []*protocol.BabelEntry {
	reader := bytes.NewReader(data)
	scanner := bufio.NewScanner(reader)

	c := &babelContext{
		entries: make([]*protocol.BabelEntry, 0),
	}

	for scanner.Scan() {
		c.line = strings.Trim(scanner.Text(), " ")
		parseLineForBabelEntry(c)
	}

	return c.entries
}

func parseLineForBabelEntry(c *babelContext) {
	m := babel.entry.FindStringSubmatch(c.line)
	if m == nil {
		return
	}

	seq32, err := strconv.ParseUint(m[3], 32, 32)
	if err == nil {
		e := &protocol.BabelEntry{
			Prefix: m[1],
			Metric: parseInt(m[2]),
			SequenceNumber: uint16(seq32),
			Routes: parseInt(m[4]),
			Sources: parseInt(m[5]),
		}
		c.entries = append(c.entries, e)
	}
}
