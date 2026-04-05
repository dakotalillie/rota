package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/dakotalillie/rota/internal/domain"
	"github.com/dakotalillie/rota/internal/infrastructure/sqlite"
)

type seedCadenceWeekly struct {
	Day      string `json:"day"`
	Time     string `json:"time"`
	TimeZone string `json:"timeZone"`
}

type seedCadence struct {
	Weekly *seedCadenceWeekly `json:"weekly,omitempty"`
}

type seedRotation struct {
	ID      string      `json:"id"`
	Name    string      `json:"name"`
	Cadence seedCadence `json:"cadence"`
}

func main() {
	dbPath := flag.String("db", "rota.db", "path to the SQLite database file")
	seedFile := flag.String("seed-file", "seed.json", "path to the JSON seed file")
	flag.Parse()

	raw, err := os.ReadFile(*seedFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading seed file: %v\n", err)
		os.Exit(1)
	}

	var rotations []seedRotation
	if err := json.Unmarshal(raw, &rotations); err != nil {
		fmt.Fprintf(os.Stderr, "error parsing seed file: %v\n", err)
		os.Exit(1)
	}

	db, err := sqlite.Open(*dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error opening database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close() //nolint:errcheck

	repo := sqlite.NewRotationRepository(db)
	ctx := context.Background()

	for _, rot := range rotations {
		if rot.ID == "" {
			fmt.Fprintf(os.Stderr, "skipping rotation with empty id (name=%q)\n", rot.Name)
			continue
		}

		r := &domain.Rotation{
			ID:   rot.ID,
			Name: rot.Name,
		}
		if rot.Cadence.Weekly != nil {
			r.Cadence.Weekly = &domain.RotationCadenceWeekly{
				Day:      rot.Cadence.Weekly.Day,
				Time:     rot.Cadence.Weekly.Time,
				TimeZone: rot.Cadence.Weekly.TimeZone,
			}
		}

		if err := repo.UpsertRotation(ctx, r); err != nil {
			fmt.Fprintf(os.Stderr, "error upserting rotation %q: %v\n", rot.ID, err)
			os.Exit(1)
		}

		fmt.Printf("upserted rotation %q (%s)\n", rot.Name, rot.ID)
	}

	fmt.Printf("\ndone: %d rotation(s) seeded\n", len(rotations))
}
