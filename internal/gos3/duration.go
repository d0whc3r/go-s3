// Copyright © 2020-2026 d0whc3r
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gos3

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var durationTokenRe = regexp.MustCompile(`(\d+)([a-zA-Zµμ]+)`)

// ParseCutoffTime parses a human-friendly duration expression and returns the
// cutoff timestamp (now minus the parsed duration).
//
// Supported units:
//   - Nanoseconds: ns
//   - Microseconds: us, µs, μs
//   - Milliseconds: ms
//   - Seconds: s, sec, second, seconds
//   - Minutes: m, min, minute, minutes
//   - Hours: h, hr, hour, hours
//   - Days: d, day, days
//   - Weeks: w, week, weeks
//   - Months: M, mo, month, months (calendar months via time.AddDate)
//   - Years: y, yr, year, years (calendar years via time.AddDate)
//
// Compound expressions such as "1M15d", "2d12h", or "1w2d" are supported.
func ParseCutoffTime(now time.Time, expr string) (time.Time, error) {
	trimmed := strings.TrimSpace(expr)
	if trimmed == "" {
		return time.Time{}, errors.New("duration cannot be empty")
	}
	if strings.HasPrefix(trimmed, "-") {
		return time.Time{}, fmt.Errorf("duration must be positive: %q", expr)
	}

	// Fast path for standard Go durations (including fractional values like 1.5h).
	if d, err := time.ParseDuration(trimmed); err == nil {
		if d <= 0 {
			return time.Time{}, fmt.Errorf("duration must be positive: %q", expr)
		}
		return now.Add(-d), nil
	}

	matches := durationTokenRe.FindAllStringSubmatchIndex(trimmed, -1)
	if len(matches) == 0 {
		return time.Time{}, fmt.Errorf("invalid duration %q", expr)
	}

	// Verify that tokens cover the entire string without invalid gaps.
	pos := 0
	var (
		years, months, days int
		dur                 time.Duration
	)

	for _, idx := range matches {
		if idx[0] != pos {
			return time.Time{}, fmt.Errorf("invalid duration %q: unexpected token near %q", expr, trimmed[pos:idx[0]])
		}
		pos = idx[1]

		numStr := trimmed[idx[2]:idx[3]]
		unit := trimmed[idx[4]:idx[5]]

		n, err := strconv.Atoi(numStr)
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid number %q in duration %q: %w", numStr, expr, err)
		}

		switch unit {
		case "ns":
			dur += time.Duration(n) * time.Nanosecond
		case "us", "µs", "μs":
			dur += time.Duration(n) * time.Microsecond
		case "ms":
			dur += time.Duration(n) * time.Millisecond
		case "s", "sec", "second", "seconds":
			dur += time.Duration(n) * time.Second
		case "m", "min", "minute", "minutes":
			dur += time.Duration(n) * time.Minute
		case "h", "hr", "hour", "hours":
			dur += time.Duration(n) * time.Hour
		case "d", "day", "days":
			days += n
		case "w", "week", "weeks":
			days += n * 7
		case "M", "mo", "month", "months":
			months += n
		case "y", "yr", "year", "years":
			years += n
		default:
			return time.Time{}, fmt.Errorf("unknown duration unit %q in %q", unit, expr)
		}
	}

	if pos != len(trimmed) {
		return time.Time{}, fmt.Errorf("invalid duration %q: trailing characters %q", expr, trimmed[pos:])
	}

	if years == 0 && months == 0 && days == 0 && dur == 0 {
		return time.Time{}, fmt.Errorf("duration must be positive: %q", expr)
	}

	cutoff := now.AddDate(-years, -months, -days).Add(-dur)
	return cutoff, nil
}
