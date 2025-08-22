package stats

import (
	"time"
	"strconv"
	"fmt"

	"github.com/dan-nicholls/danlovesto.run/pkg/contracts"
)

// TODO - Create a builder to set the defaults

func GetYearBounds(year int) (startDate, endDate time.Time) {
	startDate = time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC)
	endDate = time.Date(year, time.December, 31, 0,0,0,0, time.UTC)
	return
}

func BuildEmptyYear(startDate, endDate time.Time) contracts.Year {
	daysArr := make([]contracts.Day, 0)	
	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		dc := contracts.Day{ Date: d.Format("2006-01-02")}
		daysArr = append(daysArr, dc)
	}
	return contracts.Year{
		Year: startDate.Year(),
		From: startDate,
		To: endDate,
		Days: daysArr,
	} 
}

func BuildEmptyYears(startDate, endDate, now time.Time) []contracts.Year {
	// check if endDate is before StartDate
	if startDate.After(endDate) {
		return []contracts.Year{}
	}

	// Calculate Year range
	fromYear := startDate.Year()
	if endDate.After(now) {
		endDate = now
	}
	toYear := endDate.Year()

	years := make([]contracts.Year, 0, toYear-fromYear+1)
	for i:= fromYear; i <= toYear; i++ {
		// Check if start and end dates fall in current year
		start, end :=  GetYearBounds(i)
		if startDate.After(start) {
			start = startDate
		}
		if endDate.Before(end) {
			end = endDate
		}
		years = append(years, BuildEmptyYear(start, end))
	}
	return years
}

func CalculateStops(levels int) []float64 {
	if levels < 2 {
		return nil
	}
	n := levels - 1
	stops := make([]float64, n)
	step := 1.0 / float64(levels)
	for i := 1; i < levels; i++ {
		stops[i-1] = float64(i) * step
	}
	return stops
}

func CalculateEdges(maxVal float64, stops []float64) []float64 {
	edges := make([]float64, len(stops))
	for i, s := range stops {
		edges[i] = float64(maxVal)*s
	}
	return edges
}

func levelFromEdges(dist float64, edges []float64) int {
	for i, e := range edges {
		if dist <= e {
			return i
		}
	}
	return len(edges)
}

func GetLabelsFromEdges(edges []float64) []string {
	res := make([]string, len(edges))
	for i, e := range edges {
		s := ""
		if i > 0 && i < len(edges)-1 {
			s += "<"
		}
		if i == len(edges)-1 {
			s += ">"
		}
		s += strconv.FormatFloat(e, 'f', -1, 64)
		res[i] = s
	}
	return res
}

func ToKm(d float64) float64 {
	return d / 1000
}

func CreateHeatMap(acts []*contracts.Activity, p contracts.HeatMapParams) (contracts.HeatmapData, error) {
	// Calculate firstActive
	var minYear int
	daily := make(map[string]float64, len(acts))
	for _, a := range acts {
		d := a.StartDateLocal.Format("2006-01-02")
		y := a.StartDateLocal.Year()
		daily[d] += ToKm(a.Distance)
		if minYear == 0 || minYear > y {
			minYear = y
		}
	}

	// Calc edges and Stops
	// TODO - reduce this into previous loop
	globalMax := 0.0
	for _, d := range daily {
		if d > globalMax {
			globalMax = d
		}
	}

	// Buckets
	stops := CalculateStops(p.Levels)
	edges := CalculateEdges(globalMax, stops)
	labels := GetLabelsFromEdges(edges)

	bucketDetails := contracts.BucketDetails{
		Scale: "linear",
		Domain: [2]float64{0, globalMax},
		Levels: p.Levels,
		Stops: stops,
		Edges: edges,
		Labels: labels,
	}

	// Build Empty Years
	endYear := time.Now().Year()
	if p.ToYear != 0 {
		endYear = min(p.ToYear,time.Now().Year())
	}
	start, _ := GetYearBounds(max(p.FromYear, minYear))
	_, end := GetYearBounds(endYear)
	fmt.Printf("%d - %d", start.Year(), end.Year())
	now := time.Now().Local()
	years := BuildEmptyYears(start, end, now)

	for dayStr, dist := range daily {
		day, err := time.ParseInLocation("2006-01-02", dayStr, time.UTC)
		if err != nil {
			return contracts.HeatmapData{}, fmt.Errorf("parse day %q: %w", dayStr, err)
		}
		dc, err := GetDay(years, day)
		if err != nil {
			// Day might be outside requested range; skip safely.
			continue
		}
		dc.Distance = dist
	}
		
	// Assign levels and compute stats
	for yi := range years {
		total := 0.0
		dayCount := 0
		for di := range years[yi].Days {
			d := &years[yi].Days[di]
			d.Level = levelFromEdges(d.Distance, edges)
			total += d.Distance
			dayCount++
		}
		avg := 0.0
		if dayCount > 0 {
			avg = total / float64(dayCount)
		}
		years[yi].Stats.TotalDistance = total
		years[yi].Stats.AvgDistance = avg
	}
	
	// Build Response
	h := contracts.HeatmapData{
		Buckets: bucketDetails,
		Years: years,
		Today: now,
	}

	return h, nil
}

func GetDay(years []contracts.Year, day time.Time) (*contracts.Day, error) {
	str := day.Format("2006-01-02")
	for i := range years {
		if years[i].Year == day.Year() {
			for j := range years[i].Days {
				if years[i].Days[j].Date == str {
					return &years[i].Days[j], nil
				}	
			}
		}
	}
	return nil, fmt.Errorf("day %v not found", day.Format("2006-01-02"))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
