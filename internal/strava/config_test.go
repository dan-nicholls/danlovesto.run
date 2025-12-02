package strava

import (
	"testing"
	"time"
)

// StravaConfig

func TestLoadStravaConfig_MissingRequiredValues(t *testing.T) {
	tests := []struct {
		name  string
		setup func(t *testing.T)
		exp   string
	}{
		{
			name: "missing client id",
			setup: func(t *testing.T) {
				t.Setenv("STRAVA_CLIENT_SECRET", "secret")
				t.Setenv("RedirectURL", "http://localhost/callback")
			},
			exp: "STRAVA_CLIENT_ID must not be empty",
		},
		{
			name: "missing client secret",
			setup: func(t *testing.T) {
				t.Setenv("STRAVA_CLIENT_ID", "abc")
				t.Setenv("RedirectURL", "http://localhost/callback")
			},
			exp: "STRAVA_CLIENT_SECRET must not be empty",
		},
		{
			name: "missing redirect url",
			setup: func(t *testing.T) {
				t.Setenv("STRAVA_CLIENT_ID", "abc")
				t.Setenv("STRAVA_CLIENT_SECRET", "secret")
			},
			exp: "STRAVA_REDIRECT_URL must not be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(t)

			_, err := LoadStravaConfig()

			if err == nil {
				t.Errorf("expected error, got nil")
			}
			if err.Error() != tt.exp {
				t.Errorf("error = %v, want %v", err.Error(), tt.exp)
			}
		})
	}
}

func TestLoadStravaConfig_UseDefaults(t *testing.T) {
	// Set required values
	t.Setenv("STRAVA_CLIENT_ID", "abc")
	t.Setenv("STRAVA_CLIENT_SECRET", "secret")
	t.Setenv("STRAVA_REDIRECT_URL", "http://localhost/callback")

	res, err := LoadStravaConfig()
	if err != nil {
		t.Errorf("LoadStravaConfig() unexpected error: %v", err)
	}

	expInterval := time.Minute * 15
	if res.SyncInterval != expInterval {
		t.Errorf("SyncInterval = %v, want %v", res.SyncInterval, expInterval)
	}
}

func TestLoadStravaConfig_SetValues(t *testing.T) {
	t.Setenv("STRAVA_CLIENT_ID", "abc")
	t.Setenv("STRAVA_CLIENT_SECRET", "secret")
	t.Setenv("STRAVA_REDIRECT_URL", "http://localhost/callback")
	t.Setenv("STRAVA_SYNC_INTERVAL", "10")

	res, err := LoadStravaConfig()
	if err != nil {
		t.Errorf("LoadStravaConfig() unexpected error: %v", err)
	}

	if res.ClientID != "abc" {
		t.Errorf("ClientID = %v, want %v", res.ClientID, "abc")
	}

	if res.ClientSecret != "secret" {
		t.Errorf("ClientSecret = %v, want %v", res.ClientSecret, "secret")
	}

	if res.RedirectURL != "http://localhost/callback" {
		t.Errorf("RedirectURL = %v, want %v", res.RedirectURL, "http://localhost/callback")
	}

	expInterval := 10 * time.Minute
	if res.SyncInterval != expInterval {
		t.Errorf("SyncInterval = %v, want %v", res.SyncInterval, expInterval)
	}
}

func TestLoadStravaConfig_SyncIntervalLessThanZero(t *testing.T) {
	t.Setenv("STRAVA_CLIENT_ID", "abc")
	t.Setenv("STRAVA_CLIENT_SECRET", "secret")
	t.Setenv("STRAVA_REDIRECT_URL", "http://localhost/callback")

	t.Setenv("STRAVA_SYNC_INTERVAL", "-10")

	_, err := LoadStravaConfig()
	if err == nil {
		t.Errorf("error expected, got nil")
	}

	expError := "STRAVA_SYNC_INTERVAL must not be < 0"
	if err.Error() != expError {
		t.Errorf("error = %v, want %v", err.Error(), expError)
	}
}

func TestLoadStravaConfig_SyncIntervalIsZero(t *testing.T) {
	t.Setenv("STRAVA_CLIENT_ID", "abc")
	t.Setenv("STRAVA_CLIENT_SECRET", "secret")
	t.Setenv("STRAVA_REDIRECT_URL", "http://localhost/callback")

	t.Setenv("STRAVA_SYNC_INTERVAL", "0")

	_, err := LoadStravaConfig()
	if err == nil {
		t.Errorf("error expected, got nil")
	}

	expError := "STRAVA_SYNC_INTERVAL must not be < 0"
	if err.Error() != expError {
		t.Errorf("error = %v, want %v", err.Error(), expError)
	}
}

func TestLoadStravaConfig_SyncIntervalNotANumber(t *testing.T) {
	t.Setenv("STRAVA_CLIENT_ID", "abc")
	t.Setenv("STRAVA_CLIENT_SECRET", "secret")
	t.Setenv("STRAVA_REDIRECT_URL", "http://localhost/callback")

	t.Setenv("STRAVA_SYNC_INTERVAL", "test")

	_, err := LoadStravaConfig()
	if err == nil {
		t.Errorf("error expected, got nil")
	}

	expError := "failed to parse sync interval"
	if err.Error() != expError {
		t.Errorf(" error = %v, want %v", err.Error(), expError)
	}
}

// DBConfig

func TestLoadDBConfig_UseDefaults(t *testing.T) {
	exp := "./data/data.db"

	cfg, err := LoadDBConfig()
	if err != nil {
		t.Errorf("LoadDBConfig() unexpected error: %v", err)
	}
	if cfg.Path != exp {
		t.Errorf("Path = %v, want %v", cfg.Path, exp)
	}
}

func TestLoadDBConfig_SetPathValue(t *testing.T) {
	exp := "./test.db"
	t.Setenv("DB_PATH", "./test.db")

	cfg, err := LoadDBConfig()
	if err != nil {
		t.Errorf("LoadDBConfig() unexpected error: %v", err)
	}

	if cfg.Path != exp {
		t.Errorf("Path = %v, want %v", cfg.Path, exp)
	}
}
