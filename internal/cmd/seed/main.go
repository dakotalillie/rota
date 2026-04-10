package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/dakotalillie/rota/internal/infrastructure/sqlite"
)

type seedFile struct {
	Users     []seedUser     `json:"users"`
	Rotations []seedRotation `json:"rotations"`
	Members   []seedMember   `json:"members"`
	Overrides []seedOverride `json:"overrides"`
}

type seedUser struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type seedRotation struct {
	ID      string      `json:"id"`
	Name    string      `json:"name"`
	Cadence seedCadence `json:"cadence"`
}

type seedCadence struct {
	Weekly *seedCadenceWeekly `json:"weekly,omitempty"`
}

type seedCadenceWeekly struct {
	Day      string `json:"day"`
	Time     string `json:"time"`
	TimeZone string `json:"timeZone"`
}

type seedMember struct {
	ID              string `json:"id"`
	RotationID      string `json:"rotationID"`
	UserID          string `json:"userID"`
	Order           int    `json:"order"`
	Color           string `json:"color"`
	IsCurrent       bool   `json:"isCurrent"`
	BecameCurrentAt string `json:"becameCurrentAt"`
}

type seedOverride struct {
	ID         string `json:"id"`
	RotationID string `json:"rotationID"`
	MemberID   string `json:"memberID"`
	Start      string `json:"start"`
	End        string `json:"end"`
}

func main() {
	dbPath := flag.String("db", "rota.db", "path to the SQLite database file")
	seedFilePath := flag.String("seed-file", "seed.json", "path to the JSON seed file")
	flag.Parse()

	raw, err := os.ReadFile(*seedFilePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading seed file: %v\n", err)
		os.Exit(1)
	}

	var sf seedFile
	if err := json.Unmarshal(raw, &sf); err != nil {
		fmt.Fprintf(os.Stderr, "error parsing seed file: %v\n", err)
		os.Exit(1)
	}

	db, err := sqlite.Open(*dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error opening database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close() //nolint:errcheck

	if err := seedUsers(db, sf.Users); err != nil {
		fmt.Fprintf(os.Stderr, "error seeding users: %v\n", err)
		os.Exit(1)
	}
	if err := seedRotations(db, sf.Rotations); err != nil {
		fmt.Fprintf(os.Stderr, "error seeding rotations: %v\n", err)
		os.Exit(1)
	}
	if err := seedMembers(db, sf.Members); err != nil {
		fmt.Fprintf(os.Stderr, "error seeding members: %v\n", err)
		os.Exit(1)
	}
	if err := seedOverrides(db, sf.Overrides); err != nil {
		fmt.Fprintf(os.Stderr, "error seeding overrides: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("done: %d user(s), %d rotation(s), %d member(s), %d override(s) seeded\n",
		len(sf.Users), len(sf.Rotations), len(sf.Members), len(sf.Overrides))
}

func seedUsers(db *sql.DB, users []seedUser) error {
	for _, u := range users {
		if u.ID == "" {
			return fmt.Errorf("user missing id (name=%q)", u.Name)
		}
		data, err := json.Marshal(map[string]string{"name": u.Name})
		if err != nil {
			return err
		}
		_, err = db.ExecContext(context.Background(),
			`INSERT INTO users (id, email, data) VALUES (?, ?, ?) ON CONFLICT(id) DO UPDATE SET email=excluded.email, data=excluded.data`,
			u.ID, u.Email, string(data),
		)
		if err != nil {
			return fmt.Errorf("upsert user %q: %w", u.ID, err)
		}
		fmt.Printf("upserted user %q (%s)\n", u.Name, u.ID)
	}
	return nil
}

func seedRotations(db *sql.DB, rotations []seedRotation) error {
	for _, rot := range rotations {
		if rot.ID == "" {
			return fmt.Errorf("rotation missing id (name=%q)", rot.Name)
		}
		rec := map[string]any{"name": rot.Name}
		if rot.Cadence.Weekly != nil {
			rec["cadence"] = map[string]any{
				"weekly": map[string]string{
					"day":      rot.Cadence.Weekly.Day,
					"time":     rot.Cadence.Weekly.Time,
					"timeZone": rot.Cadence.Weekly.TimeZone,
				},
			}
		}
		data, err := json.Marshal(rec)
		if err != nil {
			return err
		}
		_, err = db.ExecContext(context.Background(),
			`INSERT INTO rotations (id, data) VALUES (?, ?) ON CONFLICT(id) DO UPDATE SET data=excluded.data`,
			rot.ID, string(data),
		)
		if err != nil {
			return fmt.Errorf("upsert rotation %q: %w", rot.ID, err)
		}
		fmt.Printf("upserted rotation %q (%s)\n", rot.Name, rot.ID)
	}
	return nil
}

func seedMembers(db *sql.DB, members []seedMember) error {
	for _, m := range members {
		if m.ID == "" {
			return fmt.Errorf("member missing id (rotationID=%q)", m.RotationID)
		}
		data, err := json.Marshal(map[string]any{"order": m.Order, "color": m.Color})
		if err != nil {
			return err
		}
		isCurrent := 0
		if m.IsCurrent {
			isCurrent = 1
		}
		var becameCurrentAt any
		if m.BecameCurrentAt != "" {
			becameCurrentAt = m.BecameCurrentAt
		}
		_, err = db.ExecContext(context.Background(),
			`INSERT INTO members (id, rotation_id, user_id, data, is_current, became_current_at)
			 VALUES (?, ?, ?, ?, ?, ?)
			 ON CONFLICT(id) DO UPDATE SET
			   rotation_id=excluded.rotation_id,
			   user_id=excluded.user_id,
			   data=excluded.data,
			   is_current=excluded.is_current,
			   became_current_at=excluded.became_current_at`,
			m.ID, m.RotationID, m.UserID, string(data), isCurrent, becameCurrentAt,
		)
		if err != nil {
			return fmt.Errorf("upsert member %q: %w", m.ID, err)
		}
		fmt.Printf("upserted member %q (rotation=%s)\n", m.ID, m.RotationID)
	}
	return nil
}

func seedOverrides(db *sql.DB, overrides []seedOverride) error {
	for _, o := range overrides {
		if o.ID == "" {
			return fmt.Errorf("override missing id (rotationID=%q)", o.RotationID)
		}
		_, err := db.ExecContext(context.Background(),
			`INSERT INTO overrides (id, rotation_id, member_id, start_time, end_time)
			 VALUES (?, ?, ?, ?, ?)
			 ON CONFLICT(id) DO UPDATE SET
			   rotation_id=excluded.rotation_id,
			   member_id=excluded.member_id,
			   start_time=excluded.start_time,
			   end_time=excluded.end_time`,
			o.ID, o.RotationID, o.MemberID, o.Start, o.End,
		)
		if err != nil {
			return fmt.Errorf("upsert override %q: %w", o.ID, err)
		}
		fmt.Printf("upserted override %q (rotation=%s)\n", o.ID, o.RotationID)
	}
	return nil
}
