package useractions

import (
	"context"

	"example.com/m/db"
)

func SavePlace(ctx context.Context, userUUID, placeID string) error {
	_, err := db.Conn.Exec(ctx,
		`INSERT INTO user_saved_places (user_uuid, place_id)
		 VALUES ($1, $2) ON CONFLICT (user_uuid, place_id) DO NOTHING`,
		userUUID, placeID,
	)
	return err
}

func RemoveSavedPlace(ctx context.Context, userUUID, placeID string) error {
	_, err := db.Conn.Exec(ctx,
		`DELETE FROM user_saved_places
		 WHERE user_uuid=$1 AND place_id=$2`,
		userUUID, placeID,
	)
	return err
}

func GetSavedPlaces(ctx context.Context, userUUID string) ([]string, error) {
	rows, err := db.Conn.Query(ctx,
		`SELECT place_id FROM user_saved_places
		 WHERE user_uuid=$1 ORDER BY saved_at DESC`,
		userUUID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var places []string
	for rows.Next() {
		var pid string
		if err := rows.Scan(&pid); err == nil {
			places = append(places, pid)
		}
	}
	return places, nil
}

func LogVisit(ctx context.Context, userUUID, placeID string) error {
	_, err := db.Conn.Exec(ctx,
		`INSERT INTO user_visit_history (user_uuid, place_id)
		 VALUES ($1, $2)`,
		userUUID, placeID,
	)
	return err
}

func GetVisitHistory(ctx context.Context, userUUID string) ([]string, error) {
	rows, err := db.Conn.Query(ctx,
		`SELECT place_id FROM user_visit_history
		 WHERE user_uuid=$1 ORDER BY visited_at DESC`,
		userUUID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var visits []string
	for rows.Next() {
		var pid string
		if err := rows.Scan(&pid); err == nil {
			visits = append(visits, pid)
		}
	}
	return visits, nil
}
