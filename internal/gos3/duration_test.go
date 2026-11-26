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

package gos3_test

import (
	"testing"
	"time"

	"github.com/d0whc3r/go-s3/internal/gos3"
)

func TestParseCutoffTime(t *testing.T) {
	now := time.Date(2026, time.November, 26, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		expr    string
		want    time.Time
		wantErr bool
	}{
		{"seconds", "30s", now.Add(-30 * time.Second), false},
		{"minutes", "15m", now.Add(-15 * time.Minute), false},
		{"hours", "2h", now.Add(-2 * time.Hour), false},
		{"fractional hours", "1.5h", now.Add(-90 * time.Minute), false},
		{"days short", "2d", now.AddDate(0, 0, -2), false},
		{"day singular", "1day", now.AddDate(0, 0, -1), false},
		{"days plural", "10days", now.AddDate(0, 0, -10), false},
		{"weeks short", "2w", now.AddDate(0, 0, -14), false},
		{"week singular", "1week", now.AddDate(0, 0, -7), false},
		{"months short uppercase M", "1M", now.AddDate(0, -1, 0), false},
		{"month singular", "1month", now.AddDate(0, -1, 0), false},
		{"months plural", "3months", now.AddDate(0, -3, 0), false},
		{"years short", "1y", now.AddDate(-1, 0, 0), false},
		{"years plural", "2years", now.AddDate(-2, 0, 0), false},
		{"compound month and days", "1M15d", now.AddDate(0, -1, -15), false},
		{"compound days and hours", "2d12h", now.AddDate(0, 0, -2).Add(-12 * time.Hour), false},
		{"compound week days minutes", "1w2d30m", now.AddDate(0, 0, -9).Add(-30 * time.Minute), false},
		{"empty", "", time.Time{}, true},
		{"negative", "-5d", time.Time{}, true},
		{"zero", "0d", time.Time{}, true},
		{"zero go duration", "0s", time.Time{}, true},
		{"unknown unit", "5q", time.Time{}, true},
		{"garbage", "abc", time.Time{}, true},
		{"trailing garbage", "2d!!!", time.Time{}, true},
		{"spaces between tokens", "2d 3h", time.Time{}, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := gos3.ParseCutoffTime(now, tc.expr)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("ParseCutoffTime(%q) expected error, got %v", tc.expr, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseCutoffTime(%q) unexpected error: %v", tc.expr, err)
			}
			if !got.Equal(tc.want) {
				t.Errorf("ParseCutoffTime(%q) = %v, want %v", tc.expr, got, tc.want)
			}
		})
	}
}
