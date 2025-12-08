package strava

import (
	"context"
	"fmt"
	"github.com/dan-nicholls/danlovesto.run/internal/api/db"
	"github.com/dan-nicholls/danlovesto.run/pkg/contracts"
)

type ActivityService struct {
	strava Client
	acts   db.ActivityStore
	pbs    db.PBStore
}

func NewActivityService(strava Client, acts db.ActivityStore, pbs db.PBStore) *ActivityService {
	return &ActivityService{
		strava: strava,
		acts:   acts,
		pbs:    pbs,
	}
}

func (as *ActivityService) Sync(ctx context.Context) error {
	fmt.Println("Starting Sync...")

	if err := ctx.Err(); err != nil {
		return err
	}

	last, err := as.acts.LatestActivityStart()
	if err != nil {
		return fmt.Errorf("unable to fetch last activity start: %e", err)
	}
	fmt.Printf("Last Activity: %s (UNIX: %d)\n", last.Local().Format("2006-01-02 15:04:05"), last.Unix())

	// Change after to 0
	acts, err := as.strava.FetchAllActivities(ctx, last.Unix(), 0, 0, false)
	if err != nil {
		return fmt.Errorf("Error fetching acts: %w", err)
	}

	for i := range acts {
		if err := ctx.Err(); err != nil {
			return ctx.Err()
		}

		act := acts[i]
		detailedAct, err := as.strava.GetActivityDetails(ctx, act.ID, false)
		if err != nil {
			fmt.Printf("failed to fetch activity details: %v\n", err)
			continue
		}
		// fmt.Printf("Act[%d]: %+v\n", i, act)
		err = as.AddActivity(detailedAct)
		if err != nil {
			fmt.Printf("unable to store fetched activity %d: %v\n", act.ID, err)
			continue
		}
		fmt.Printf("added activity %d to store\n", act.ID)
	}

	fmt.Println("Sync Complete.")
	return nil
}

func (as *ActivityService) AddActivity(stravaAct contracts.StravaDetailedActivity) error {
	// 1. Check PBs
	pbList := as.DetectPBsFromActivity(stravaAct)

	// 2. Convert to domain Activity
	act := MapStravaToActivity(stravaAct)

	// 3. Store Activity
	id, err := as.acts.UpsertActivity(&act)
	if err != nil {
		return fmt.Errorf("failed to store act: %w", err)
	}
	fmt.Printf("added the activity to db: %v\n", id)

	// 4. Store PBs
	for i := range pbList {
		pb := pbList[i]
		err := as.pbs.SetPB(pb.Distance, pb.Duration, pb.ActivityID)
		if err != nil {
			fmt.Printf("failed to store PB %v for act %v: %v\n", pb, act.ID, err)
		}
	}
	return nil
}

func (as *ActivityService) DetectPBsFromActivity(a contracts.StravaDetailedActivity) []contracts.PersonalBest {
	pbs := make([]contracts.PersonalBest, 0)
	for i := range a.BestEfforts {
		be := a.BestEfforts[i]
		if be.PrRank != 1 {
			continue
		}
		fmt.Printf("PR for %v found: %d\n", be.Name, be.MovingTime)
		newPB := contracts.PersonalBest{
			Distance:   be.Name,
			Duration:   be.ElapsedTime,
			ActivityID: a.ID,
		}
		pbs = append(pbs, newPB)
	}
	return pbs
}

func MapStravaToActivity(stravaAct contracts.StravaDetailedActivity) contracts.Activity {
	return contracts.Activity{
		ID:                 stravaAct.ID,
		Name:               stravaAct.Name,
		AthleteID:          stravaAct.Athlete.ID,
		Distance:           stravaAct.Distance,
		MovingTime:         stravaAct.MovingTime,
		ElapsedTime:        stravaAct.ElapsedTime,
		TotalElevationGain: stravaAct.TotalElevationGain,
		Type:               stravaAct.Type,
		StartDate:          stravaAct.StartDate,
		StartDateLocal:     stravaAct.StartDateLocal,
		Timezone:           stravaAct.Timezone,
		UtcOffset:          int(stravaAct.UtcOffset),

		LocationCity:    stravaAct.LocationCity,
		LocationState:   stravaAct.LocationState,
		LocationCountry: stravaAct.LocationCountry,

		Map: contracts.ActivityMap{
			ID:              stravaAct.Map.ID,
			Polyline:        stravaAct.Map.Polyline,
			SummaryPolyline: stravaAct.Map.SummaryPolyline,
		},

		StartLatLng: stravaAct.StartLatlng,
		EndLatLng:   stravaAct.EndLatlng,

		AverageSpeed: stravaAct.AverageSpeed,
		MaxSpeed:     stravaAct.MaxSpeed,

		ElevHigh: stravaAct.ElevHigh,
		ElevLow:  stravaAct.ElevLow,
	}
}
