package glance

import (
	"errors"
	"fmt"
	builder "github.com/doug-martin/goqu/v9"
	"math"
	"sort"
	"time"
)

var (
	ErrUnsupportedGranularity = errors.New("unsupported granularity")
)

const (
	GranMinute         = 60
	GranFiveMinutes    = 5 * GranMinute
	GranFifteenMinutes = 15 * GranMinute
	GranHour           = 60 * GranMinute
	GranDay            = 24 * GranHour
	GranWeek           = 7 * GranDay
	GranMonth          = 30 * GranDay
	GranQuarter        = 90 * GranDay
)

// ParseGranularity Converts the test granularity to its analogous value in the time object
func ParseGranularity(granularity string) (granularitySeconds time.Duration, err error) {
	switch granularity {
	case "minute", "1m":
		return time.Minute, nil
	case "5 minutes", "5m":
		return time.Minute * 5, nil
	case "15 minutes", "15m":
		return time.Minute * 15, nil
	case "hour", "1h":
		return time.Hour, nil
	case "day", "1d":
		return time.Hour * 24, nil
	case "week", "1w":
		return time.Hour * 24 * 7, nil
	case "month", "1mo":
		return time.Hour * 24 * 30, nil
	case "quarter", "1q":
		return time.Hour * 24 * 90, nil
	default:
		return 0, ErrUnsupportedGranularity
	}
}

// GranularityToString Converts the duration to the corresponding function in Clickhouse
func GranularityToString(granularity time.Duration) (string, int64, error) {
	var (
		timeFunc string
		seconds  int64
	)

	switch {
	case granularity >= time.Minute && granularity < 5*time.Minute:
		timeFunc = "toStartOfMinute"
		seconds = GranMinute
	case granularity >= 5*time.Minute && granularity < 15*time.Minute:
		timeFunc = "toStartOfFiveMinute"
		seconds = GranFiveMinutes
	case granularity >= 15*time.Minute && granularity < time.Hour:
		timeFunc = "toStartOfFifteenMinutes"
		seconds = GranFifteenMinutes
	case granularity >= time.Hour && granularity < 24*time.Hour:
		timeFunc = "toStartOfHour"
		seconds = GranHour
	case granularity >= 24*time.Hour && granularity < time.Hour*24*7:
		timeFunc = "toStartOfDay"
		seconds = GranDay
	case granularity >= time.Hour*24*7 && granularity < time.Hour*24*30:
		timeFunc = "toStartOfWeek"
		seconds = GranWeek
	case granularity >= time.Hour*24*30 && granularity < time.Hour*24*90:
		timeFunc = "toStartOfMonth"
		seconds = GranMonth
	case granularity >= time.Hour*24*90:
		timeFunc = "toStartOfQuarter"
		seconds = GranQuarter
	default:
		return "", 0, fmt.Errorf("mapping granularity %s to time func: %w", granularity, ErrUnsupportedGranularity)
	}

	return timeFunc, seconds, nil
}

// BuildTimeSeriesQuery Unified creates a final time series query
//
// sql, _, _ := mainSeries.ToSQL()
// err := connection.SelectContext(ctx, &yourStruct, sql)
func BuildTimeSeriesQuery(main, times, keys *builder.SelectDataset) *builder.SelectDataset {
	return builder.Select(builder.L("*")).From(times.CrossJoin(keys)).UnionAll(main)
}

// BuildTimeSeriesQueries Unified creates a time series of data from real data and generated data along a timeline
// so that there are no time "windows" in the series
// Returns three instances of the request for the possibility of customization beyond the limits of the function
// of each of them
//
// mainSeries, timeSeries, keySeries := glance.BuildTimeSeriesQueries(
//		from,
//		to,
//		timeFunc,
//		granularitySeconds,
//		"platform",
//		"uniq(device_id)",
//		"uniqs",
//		"stream.heatmap",
// )
//
// Customize
// mainSeries.Where(...)
// timeSeries.Where(...)
func BuildTimeSeriesQueries(
	from, to time.Time,
	timeFunc string,
	granularitySeconds int64,
	keyColumn string,
	valueColumnExp string,
	valueColumnName string,
	tableName string,
) (
	*builder.SelectDataset,
	*builder.SelectDataset,
	*builder.SelectDataset,
) {
	mainQuery := builder.
		Select(
			builder.L(fmt.Sprintf("%s(insert_ts)", timeFunc)).As("time"),
			builder.L(valueColumnExp).As(valueColumnName),
			builder.C(keyColumn),
		).
		From(tableName).
		Where(
			builder.Or(
				builder.C("insert_date").Gte(Date(from)),
				builder.C("insert_date").Lte(Date(to)),
			),
			builder.And(
				builder.C("insert_ts").Gte(Datetime(from)),
				builder.C("insert_ts").Lte(Datetime(to)),
			),
		).
		GroupBy(
			builder.C("time"),
			builder.C(keyColumn),
		).
		Order(
			builder.I("time").Asc(),
		)

	// generate zero points in timeline
	countNumbers := int64(math.Ceil(float64(to.Unix()-from.Unix()) / float64(granularitySeconds)))
	timeSeries := builder.
		Select(
			builder.L("ts").As("time"),
			builder.L("toUInt64(0)").As(valueColumnName),
			builder.C(keyColumn),
		).
		From(
			builder.
				Select(
					builder.L(fmt.Sprintf(matchFunction(timeFunc), timeFunc, Datetime(from), granularitySeconds)).As("ts"),
				).
				From(
					builder.L(fmt.Sprintf("system.numbers limit %d", countNumbers)),
				).
				As("X"),
		)

	// get keys for join zero points
	keySeries := builder.
		Select(
			builder.C(keyColumn),
		).
		From(tableName).
		Where(
			builder.And(
				builder.C("insert_date").Gte(Date(from)),
				builder.C("insert_date").Lte(Date(to)),
				builder.C("insert_ts").Gte(Datetime(from)),
				builder.C("insert_ts").Lt(Datetime(to)),
			),
		).
		GroupBy(
			builder.C(keyColumn),
		).
		As("Y")

	return mainQuery, timeSeries, keySeries
}

func matchFunction(timeFunc string) string {
	switch timeFunc {
	case "toStartOfWeek":
		return "toDateTime(toUnixTimestamp(toDateTime(%s(toDateTime('%s'), 1)))+number*%d)"
	case "toStartOfMonth", "toStartOfQuarter":
		return "toDateTime(toUnixTimestamp(toDateTime(%s(toDateTime('%s'))))+number*%d)"
	}

	return "toDateTime(toUnixTimestamp(%s(toDateTime('%s')))+number*%d)"
}

type Point struct {
	Key   string
	Value int64
	Time  time.Time
}
type point struct {
	Value int64  `json:"value"`
	Time  string `json:"time"`
}
type UnifiedLinearPoints map[string][]point
type timesMap map[string]map[int64]int64

func BuildPoints(timeLayout string, points []Point) UnifiedLinearPoints {
	times := timesMap{}
	for _, stat := range points {
		key := stat.Key
		value := stat.Value
		point := stat.Time

		if _, ok := times[key]; !ok {
			times[key] = map[int64]int64{}
		}

		t := point.Unix()
		if v, ok := times[key][t]; !ok || value >= v {
			times[key][t] = value
		}
	}

	serial := map[string][]point{}
	for code, timeSeries := range times {
		if _, ok := serial[code]; !ok {
			serial[code] = []point{}
		}

		for t, v := range timeSeries {
			serial[code] = append(serial[code], point{
				Value: v,
				Time:  dateFormat(timeLayout, t),
			})
		}

		sort.Slice(serial[code], func(d, e int) bool {
			return serial[code][d].Time < serial[code][e].Time
		})
	}
	return serial
}

func dateFormat(layout string, d int64) string {
	t := time.Unix(d, 0)
	if layout == "" {
		layout = "2006-01-02 15:04:05"
	}
	return t.Format(layout)
}
