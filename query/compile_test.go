package query_test

import (
	"testing"
	"time"

	"github.com/influxdata/influxdb/influxql"
	"github.com/influxdata/influxdb/query"
)

func TestCompile_Success(t *testing.T) {
	for _, tt := range []string{
		`SELECT time, value FROM cpu`,
		`SELECT value FROM cpu`,
		`SELECT value, host FROM cpu`,
		`SELECT * FROM cpu`,
		`SELECT time, * FROM cpu`,
		`SELECT value, * FROM cpu`,
		`SELECT max(value) FROM cpu`,
		`SELECT max(value), host FROM cpu`,
		`SELECT max(value), * FROM cpu`,
		`SELECT max(*) FROM cpu`,
		`SELECT max(/val/) FROM cpu`,
		`SELECT min(value) FROM cpu`,
		`SELECT min(value), host FROM cpu`,
		`SELECT min(value), * FROM cpu`,
		`SELECT min(*) FROM cpu`,
		`SELECT min(/val/) FROM cpu`,
		`SELECT first(value) FROM cpu`,
		`SELECT first(value), host FROM cpu`,
		`SELECT first(value), * FROM cpu`,
		`SELECT first(*) FROM cpu`,
		`SELECT first(/val/) FROM cpu`,
		`SELECT last(value) FROM cpu`,
		`SELECT last(value), host FROM cpu`,
		`SELECT last(value), * FROM cpu`,
		`SELECT last(*) FROM cpu`,
		`SELECT last(/val/) FROM cpu`,
		`SELECT count(value) FROM cpu`,
		`SELECT count(distinct(value)) FROM cpu`,
		`SELECT count(*) FROM cpu`,
		`SELECT count(/val/) FROM cpu`,
		`SELECT mean(value) FROM cpu`,
		`SELECT mean(*) FROM cpu`,
		`SELECT mean(/val/) FROM cpu`,
		`SELECT min(value), max(value) FROM cpu`,
		`SELECT min(*), max(*) FROM cpu`,
		`SELECT min(/val/), max(/val/) FROM cpu`,
		`SELECT first(value), last(value) FROM cpu`,
		`SELECT first(*), last(*) FROM cpu`,
		`SELECT first(/val/), last(/val/) FROM cpu`,
		`SELECT count(value) FROM cpu WHERE time >= now() - 1h GROUP BY time(10m)`,
		`SELECT distinct value FROM cpu`,
		`SELECT distinct(value) FROM cpu`,
		`SELECT value / total FROM cpu`,
		`SELECT min(value) / total FROM cpu`,
		`SELECT max(value) / total FROM cpu`,
		`SELECT top(value, 1) FROM cpu`,
		`SELECT top(value, host, 1) FROM cpu`,
		`SELECT top(value, 1), host FROM cpu`,
		`SELECT bottom(value, 1) FROM cpu`,
		`SELECT bottom(value, host, 1) FROM cpu`,
		`SELECT bottom(value, 1), host FROM cpu`,
	} {
		t.Run(tt, func(t *testing.T) {
			stmt, err := influxql.ParseStatement(tt)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			s := stmt.(*influxql.SelectStatement)

			opt := query.CompileOptions{}
			if _, err := query.Compile(s, opt); err != nil {
				t.Errorf("unexpected error: %s", err)
			}
		})
	}
}

func TestCompile_Failures(t *testing.T) {
	for _, tt := range []struct {
		s   string
		err string
	}{
		{s: `SELECT time FROM cpu`, err: `at least 1 non-time field must be queried`},
		{s: `SELECT value, mean(value) FROM cpu`, err: `mixing aggregate and non-aggregate queries is not supported`},
		{s: `SELECT value, max(value), min(value) FROM cpu`, err: `mixing multiple selector functions with tags or fields is not supported`},
		{s: `SELECT top(value, 10), max(value) FROM cpu`, err: `selector function top() cannot be combined with other functions`},
		{s: `SELECT bottom(value, 10), max(value) FROM cpu`, err: `selector function bottom() cannot be combined with other functions`},
		{s: `SELECT count() FROM cpu`, err: `invalid number of arguments for count, expected 1, got 0`},
		{s: `SELECT count(value, host) FROM cpu`, err: `invalid number of arguments for count, expected 1, got 2`},
		{s: `SELECT min() FROM cpu`, err: `invalid number of arguments for min, expected 1, got 0`},
		{s: `SELECT min(value, host) FROM cpu`, err: `invalid number of arguments for min, expected 1, got 2`},
		{s: `SELECT max() FROM cpu`, err: `invalid number of arguments for max, expected 1, got 0`},
		{s: `SELECT max(value, host) FROM cpu`, err: `invalid number of arguments for max, expected 1, got 2`},
		{s: `SELECT sum() FROM cpu`, err: `invalid number of arguments for sum, expected 1, got 0`},
		{s: `SELECT sum(value, host) FROM cpu`, err: `invalid number of arguments for sum, expected 1, got 2`},
		{s: `SELECT first() FROM cpu`, err: `invalid number of arguments for first, expected 1, got 0`},
		{s: `SELECT first(value, host) FROM cpu`, err: `invalid number of arguments for first, expected 1, got 2`},
		{s: `SELECT last() FROM cpu`, err: `invalid number of arguments for last, expected 1, got 0`},
		{s: `SELECT last(value, host) FROM cpu`, err: `invalid number of arguments for last, expected 1, got 2`},
		{s: `SELECT mean() FROM cpu`, err: `invalid number of arguments for mean, expected 1, got 0`},
		{s: `SELECT mean(value, host) FROM cpu`, err: `invalid number of arguments for mean, expected 1, got 2`},
		{s: `SELECT distinct(value), max(value) FROM cpu`, err: `aggregate function distinct() cannot be combined with other functions or fields`},
		{s: `SELECT count(distinct(value)), max(value) FROM cpu`, err: `aggregate function distinct() cannot be combined with other functions or fields`},
		{s: `SELECT count(distinct()) FROM cpu`, err: `distinct function requires at least one argument`},
		{s: `SELECT count(distinct(value, host)) FROM cpu`, err: `distinct function can only have one argument`},
		{s: `SELECT count(distinct(2)) FROM cpu`, err: `expected field argument in distinct()`},
		{s: `SELECT value FROM cpu GROUP BY now()`, err: `only time() calls allowed in dimensions`},
		{s: `SELECT value FROM cpu GROUP BY time()`, err: `time dimension expected 1 or 2 arguments`},
		{s: `SELECT value FROM cpu GROUP BY time(5m, 30s, 1ms)`, err: `time dimension expected 1 or 2 arguments`},
		{s: `SELECT value FROM cpu GROUP BY time('unexpected')`, err: `time dimension must have duration argument`},
		{s: `SELECT value FROM cpu GROUP BY time(5m), time(1m)`, err: `multiple time dimensions not allowed`},
		{s: `SELECT value FROM cpu GROUP BY time(5m, unexpected())`, err: `time dimension offset function must be now()`},
		{s: `SELECT value FROM cpu GROUP BY time(5m, now(1m))`, err: `time dimension offset now() function requires no arguments`},
		{s: `SELECT value FROM cpu GROUP BY time(5m, 'unexpected')`, err: `time dimension offset must be duration or now()`},
		{s: `SELECT value FROM cpu GROUP BY 'unexpected'`, err: `only time and tag dimensions allowed`},
		{s: `SELECT top(value) FROM cpu`, err: `invalid number of arguments for top, expected at least 2, got 1`},
		{s: `SELECT top('unexpected', 5) FROM cpu`, err: `expected first argument to be a field in top(), found 'unexpected'`},
		{s: `SELECT top(value, 'unexpected', 5) FROM cpu`, err: `only fields or tags are allowed in top(), found 'unexpected'`},
		{s: `SELECT top(value, 2.5) FROM cpu`, err: `expected integer as last argument in top(), found 2.500`},
		{s: `SELECT top(value, -1) FROM cpu`, err: `limit (-1) in top function must be at least 1`},
		{s: `SELECT bottom(value) FROM cpu`, err: `invalid number of arguments for bottom, expected at least 2, got 1`},
		{s: `SELECT bottom('unexpected', 5) FROM cpu`, err: `expected first argument to be a field in bottom(), found 'unexpected'`},
		{s: `SELECT bottom(value, 'unexpected', 5) FROM cpu`, err: `only fields or tags are allowed in bottom(), found 'unexpected'`},
		{s: `SELECT bottom(value, 2.5) FROM cpu`, err: `expected integer as last argument in bottom(), found 2.500`},
		{s: `SELECT bottom(value, -1) FROM cpu`, err: `limit (-1) in bottom function must be at least 1`},
		{s: `SELECT value FROM cpu WHERE time >= now() - 10m OR time < now() - 5m`, err: `cannot use OR with time conditions`},
		{s: `SELECT value FROM cpu WHERE value`, err: `invalid condition expression: value`},
		{s: `SELECT count(value), * FROM cpu`, err: `mixing aggregate and non-aggregate queries is not supported`},
		{s: `SELECT max(*), host FROM cpu`, err: `mixing aggregate and non-aggregate queries is not supported`},
		{s: `SELECT count(value), /ho/ FROM cpu`, err: `mixing aggregate and non-aggregate queries is not supported`},
		{s: `SELECT max(/val/), * FROM cpu`, err: `mixing aggregate and non-aggregate queries is not supported`},
		{s: `SELECT a(value) FROM cpu`, err: `undefined function a()`},
		{s: `SELECT count(max(value)) FROM myseries`, err: `expected field argument in count()`},
		{s: `SELECT count(distinct('value')) FROM myseries`, err: `expected field argument in distinct()`},
		{s: `SELECT distinct('value') FROM myseries`, err: `expected field argument in distinct()`},
		{s: `SELECT min(max(value)) FROM myseries`, err: `expected field argument in min()`},
		{s: `SELECT min(distinct(value)) FROM myseries`, err: `expected field argument in min()`},
		{s: `SELECT max(max(value)) FROM myseries`, err: `expected field argument in max()`},
		{s: `SELECT sum(max(value)) FROM myseries`, err: `expected field argument in sum()`},
		{s: `SELECT first(max(value)) FROM myseries`, err: `expected field argument in first()`},
		{s: `SELECT last(max(value)) FROM myseries`, err: `expected field argument in last()`},
		{s: `SELECT mean(max(value)) FROM myseries`, err: `expected field argument in mean()`},
		{s: `SELECT median(max(value)) FROM myseries`, err: `expected field argument in median()`},
		{s: `SELECT mode(max(value)) FROM myseries`, err: `expected field argument in mode()`},
		{s: `SELECT stddev(max(value)) FROM myseries`, err: `expected field argument in stddev()`},
		{s: `SELECT spread(max(value)) FROM myseries`, err: `expected field argument in spread()`},
		{s: `SELECT top() FROM myseries`, err: `invalid number of arguments for top, expected at least 2, got 0`},
		{s: `SELECT top(field1) FROM myseries`, err: `invalid number of arguments for top, expected at least 2, got 1`},
		{s: `SELECT top(field1,foo) FROM myseries`, err: `expected integer as last argument in top(), found foo`},
		{s: `SELECT top(field1,host,'server',foo) FROM myseries`, err: `expected integer as last argument in top(), found foo`},
		{s: `SELECT top(field1,5,'server',2) FROM myseries`, err: `only fields or tags are allowed in top(), found 5`},
		{s: `SELECT top(field1,max(foo),'server',2) FROM myseries`, err: `only fields or tags are allowed in top(), found max(foo)`},
		{s: `SELECT top(value, 10) + count(value) FROM myseries`, err: `selector function top() cannot be combined with other functions`},
		{s: `SELECT top(max(value), 10) FROM myseries`, err: `expected first argument to be a field in top(), found max(value)`},
		{s: `SELECT bottom() FROM myseries`, err: `invalid number of arguments for bottom, expected at least 2, got 0`},
		{s: `SELECT bottom(field1) FROM myseries`, err: `invalid number of arguments for bottom, expected at least 2, got 1`},
		{s: `SELECT bottom(field1,foo) FROM myseries`, err: `expected integer as last argument in bottom(), found foo`},
		{s: `SELECT bottom(field1,host,'server',foo) FROM myseries`, err: `expected integer as last argument in bottom(), found foo`},
		{s: `SELECT bottom(field1,5,'server',2) FROM myseries`, err: `only fields or tags are allowed in bottom(), found 5`},
		{s: `SELECT bottom(field1,max(foo),'server',2) FROM myseries`, err: `only fields or tags are allowed in bottom(), found max(foo)`},
		{s: `SELECT bottom(value, 10) + count(value) FROM myseries`, err: `selector function bottom() cannot be combined with other functions`},
		{s: `SELECT bottom(max(value), 10) FROM myseries`, err: `expected first argument to be a field in bottom(), found max(value)`},
		{s: `SELECT percentile() FROM myseries`, err: `invalid number of arguments for percentile, expected 2, got 0`},
		{s: `SELECT percentile(field1) FROM myseries`, err: `invalid number of arguments for percentile, expected 2, got 1`},
		{s: `SELECT percentile(field1, foo) FROM myseries`, err: `expected float argument in percentile()`},
		{s: `SELECT percentile(max(field1), 75) FROM myseries`, err: `expected field argument in percentile()`},
		{s: `SELECT field1 FROM foo group by time(1s)`, err: `GROUP BY requires at least one aggregate function`},
		{s: `SELECT field1 FROM foo fill(none)`, err: `fill(none) must be used with a function`},
		{s: `SELECT field1 FROM foo fill(linear)`, err: `fill(linear) must be used with a function`},
		{s: `SELECT count(value), value FROM foo`, err: `mixing aggregate and non-aggregate queries is not supported`},
		{s: `SELECT count(value) FROM foo group by time(1s)`, err: `aggregate functions with GROUP BY time require a WHERE time clause with a lower limit`},
		{s: `SELECT count(value) FROM foo group by time(500ms)`, err: `aggregate functions with GROUP BY time require a WHERE time clause with a lower limit`},
		{s: `SELECT count(value) FROM foo group by time(1s) where host = 'hosta.influxdb.org'`, err: `aggregate functions with GROUP BY time require a WHERE time clause with a lower limit`},
		{s: `SELECT count(value) FROM foo group by time(1s) where time < now()`, err: `aggregate functions with GROUP BY time require a WHERE time clause with a lower limit`},
		{s: `SELECT count(value) FROM foo group by time`, err: `time() is a function and expects at least one argument`},
		{s: `SELECT count(value) FROM foo group by 'time'`, err: `only time and tag dimensions allowed`},
		{s: `SELECT count(value) FROM foo where time > now() and time < now() group by time()`, err: `time dimension expected 1 or 2 arguments`},
		{s: `SELECT count(value) FROM foo where time > now() and time < now() group by time(b)`, err: `time dimension must have duration argument`},
		{s: `SELECT count(value) FROM foo where time > now() and time < now() group by time(1s), time(2s)`, err: `multiple time dimensions not allowed`},
		{s: `SELECT count(value) FROM foo where time > now() and time < now() group by time(1s, b)`, err: `time dimension offset must be duration or now()`},
		{s: `SELECT distinct(field1), sum(field1) FROM myseries`, err: `aggregate function distinct() cannot be combined with other functions or fields`},
		{s: `SELECT distinct(field1), field2 FROM myseries`, err: `aggregate function distinct() cannot be combined with other functions or fields`},
		{s: `SELECT distinct(field1, field2) FROM myseries`, err: `distinct function can only have one argument`},
		{s: `SELECT distinct() FROM myseries`, err: `distinct function requires at least one argument`},
		{s: `SELECT distinct field1, field2 FROM myseries`, err: `aggregate function distinct() cannot be combined with other functions or fields`},
		{s: `SELECT count(distinct field1, field2) FROM myseries`, err: `invalid number of arguments for count, expected 1, got 2`},
		{s: `select count(distinct(too, many, arguments)) from myseries`, err: `distinct function can only have one argument`},
		{s: `select count() from myseries`, err: `invalid number of arguments for count, expected 1, got 0`},
	} {
		t.Run(tt.s, func(t *testing.T) {
			stmt, err := influxql.ParseStatement(tt.s)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			s := stmt.(*influxql.SelectStatement)

			opt := query.CompileOptions{}
			if _, err := query.Compile(s, opt); err == nil {
				t.Error("expected error")
			} else if have, want := err.Error(), tt.err; have != want {
				t.Errorf("unexpected error: %s != %s", have, want)
			}
		})
	}
}

func TestCompile_ParseCondition(t *testing.T) {
	mustParseTime := func(value string) time.Time {
		ts, err := time.Parse(time.RFC3339, value)
		if err != nil {
			t.Fatalf("unable to parse time: %s", err)
		}
		return ts
	}

	// Set the now() time to the beginning of 2017.
	now := mustParseTime("2017-01-01T00:00:00Z")
	for _, tt := range []struct {
		cond     string
		min, max time.Time
		exp      string
		err      string
	}{
		{
			cond: `time >= now()`,
			min:  now,
		},
		{
			cond: `time < now()`,
			max:  now.Add(-time.Nanosecond),
		},
		{
			cond: `time >= now() OR host = 'server01'`,
			err:  "cannot use OR with time conditions",
		},
		{
			cond: `time >= now() AND host = 'server01'`,
			exp:  `host = 'server01'`,
			min:  now,
		},
		{
			cond: `value`,
			err:  "invalid condition expression: value",
		},
		{
			cond: `4`,
			err:  "invalid condition expression: 4",
		},
		{
			cond: `time >= 'today'`,
			err:  "invalid operation: time and *influxql.StringLiteral are not compatible",
		},
		{
			cond: `time != now()`,
			err:  "invalid time comparison operator: !=",
		},
		{
			cond: `now() <= time`,
			min:  now,
		},
		{
			cond: `time >= now() AND (host = 'server01' OR host = 'server02')`,
			min:  now,
			exp:  `host = 'server01' OR host = 'server02'`,
		},
		{
			cond: `host = 'server01' OR (time >= now() AND host = 'server02')`,
			err:  "cannot use OR with time conditions",
		},
		{
			cond: `host = 'server01' AND (time >= now() AND host = 'server02')`,
			min:  now,
			exp:  `host = 'server01' AND host = 'server02'`,
		},
	} {
		t.Run(tt.cond, func(t *testing.T) {
			valuer := influxql.NowValuer{Now: now}
			expr, err := influxql.ParseExpr(tt.cond)
			if err != nil {
				t.Fatalf("unable to parse expression: %s", err)
			}

			cond, tr, err := query.ParseCondition(expr, &valuer)
			if err != nil {
				if tt.err == "" {
					t.Fatalf("unexpected error: %s", err)
				} else if have, want := err.Error(), tt.err; have != want {
					t.Fatalf("unexpected error: %s != %s", have, want)
				}
				return
			}

			if cond != nil {
				if have, want := cond.String(), tt.exp; have != want {
					t.Errorf("unexpected condition: %s != %s", have, want)
				}
			} else if have, want := "", tt.exp; have != want {
				t.Errorf("unexpected condition: %s != %s", have, want)
			}

			if tr != nil {
				if !tr.Min.Equal(tt.min) {
					t.Errorf("unexpected min time: %s != %s", tr.Min, tt.min)
				}
				if !tr.Max.Equal(tt.max) {
					t.Errorf("unexpected max time: %s != %s", tr.Max, tt.max)
				}
			} else {
				if !tt.min.IsZero() {
					t.Errorf("unexpected min time: %s != %s", time.Time{}, tt.min)
				}
				if !tt.max.IsZero() {
					t.Errorf("unexpected max time: %s != %s", time.Time{}, tt.max)
				}
			}
		})
	}
}
