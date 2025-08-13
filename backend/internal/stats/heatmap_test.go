package stats

import (
	"time"
	"testing"

	//"github.com/dan-nicholls/danlovesto.run/backend/internal/model"
)

//func TestCreateHeatMap(t *testing.T) {
//	act1Time, _ := time.Parse(time.RFC3339, "2024-01-03T08:09:27Z")
//	act2Time, _ := time.Parse(time.RFC3339, "2023-12-29T08:09:27Z")
//	act3Time, _ := time.Parse(time.RFC3339, "2024-01-03T08:09:27Z")
//	
//	p := HeatMapParameters{
//		FromYear: 2023,
//		ToYear: 2024,
//	}
//
//	acts := []model.Activity{
//		model.Activity{
//			StartDateLocal: act1Time,
//			Distance: 10000,
//		},
//		model.Activity{
//			StartDateLocal: act2Time,
//			Distance: 30000,
//		},
//		model.Activity{
//			StartDateLocal: act3Time,
//			Distance: 50000,
//		},
//	}
//
//	res, err := CreateHeatMap(acts, p)
//	if err != nil {
//		t.Fatalf("Failed to create  %v", err)	
//	}
//
//
//	dc := findDay(res, act1Time.Format("2006-01-02"))
//	if dc == nil {
//		t.Fatalf("Failed to fetch day: %v", act1Time.Format("2006-01-02"))
//	}
//
//	expect := 6000.00
//	if dc.Distance != expect {
//		t.Fatalf("got %v km, want %v km", dc.Distance, expect)	
//	}
//	
//}
//

// -- helpers --

func dUTC(y int, m time.Month, day int) time.Time {
	return time.Date(y, m, day, 0, 0, 0, 0, time.UTC)
}

// --- BuildEmptyYears tests ---

func TestBuildEmptyYears_EndBeforeStartReturnsEmpty(t *testing.T) {
	start, end, now := dUTC(2025, 3, 5), dUTC(2025, 3, 1), dUTC(2025, 3, 10)

	got := BuildEmptyYears(start, end, now)

	if len(got) != 0 {
		t.Errorf("len(years) = %d, want 0", len(got))
	}
}

func TestBuildEmptyYears_OneYearRangeReturnsOne(t *testing.T) {
	start, end, now := dUTC(2025, 3, 1), dUTC(2025, 3, 5), dUTC(2025, 12, 31)

	got := BuildEmptyYears(start, end, now)

	if len(got) != 1 {
		t.Errorf("len(years) = %d, want 1", len(got))
	}
}

func TestBuildEmptyYears_MultiYearRangeReturnsTwo(t *testing.T) {
	start, end, now := dUTC(2024, 12, 20), dUTC(2025, 1, 10), dUTC(2025, 12, 31)

	got := BuildEmptyYears(start, end, now)

	if len(got) != 2 {
		t.Errorf("len(years) = %d, want 2", len(got))
	}
}

func TestBuildEmptyYears_FirstYearFromClampedToStart(t *testing.T) {
	start, end, now := dUTC(2025, 3, 3), dUTC(2025, 3, 5), dUTC(2025, 12, 31)

	got := BuildEmptyYears(start, end, now)

	if !got[0].From.Equal(start) {
		t.Errorf("Year[0].From = %v, want %v", got[0].From, start)
	}
}

func TestBuildEmptyYears_LastYearToClampedToEnd(t *testing.T) {
	start, end, now := dUTC(2025, 3, 1), dUTC(2025, 3, 5), dUTC(2025, 12, 31)

	got := BuildEmptyYears(start, end, now)

	if !got[len(got)-1].To.Equal(end) {
		t.Errorf("Year[last].To = %v, want %v", got[len(got)-1].To, end)
	}
}

func TestBuildEmptyYears_MiddleYearBoundsWholeCalendarYear(t *testing.T) {
	start, end, now := dUTC(2023, 11, 15), dUTC(2025, 2, 2), dUTC(2025, 12, 31)

	got := BuildEmptyYears(start, end, now)

	mid := got[1] // 2024
	if !mid.From.Equal(dUTC(2024, 1, 1)) {
		t.Errorf("middle.From = %v, want 2024-01-01", mid.From)
	}
	if !mid.To.Equal(dUTC(2024, 12, 31)) {
		t.Errorf("middle.To = %v, want 2024-12-31", mid.To)
	}
}

func TestBuildEmptyYears_ClampsEndToNow(t *testing.T) {
	start := dUTC(2020, 1, 1)
	end := dUTC(2099, 12, 31)
	now := dUTC(2025, 6, 15) // pretend today is mid-2025

	got := BuildEmptyYears(start, end, now)

	if got[len(got)-1].Year != 2025 {
		t.Errorf("last Year = %d, want 2025", got[len(got)-1].Year)
	}
}

func TestBuildEmptyYears_FirstYearHasDays(t *testing.T) {
	start, end, now := dUTC(2025, 3, 1), dUTC(2025, 3, 3), dUTC(2025, 12, 31)

	got := BuildEmptyYears(start, end, now)

	if len(got[0].Days) == 0 {
		t.Errorf("Year[0].Days is empty, want non-empty")
	}
}
