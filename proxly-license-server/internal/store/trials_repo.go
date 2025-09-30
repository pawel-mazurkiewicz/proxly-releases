package store

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pawel-mazurkiewicz/proxly-license-server/pkg/types"
)

type TrialsRepository struct {
	db *sql.DB
}

func NewTrialsRepository(db *sql.DB) *TrialsRepository {
	return &TrialsRepository{db: db}
}

func (r *TrialsRepository) Create(trial *types.Trial) error {
	query := `
        INSERT INTO trials (id, client_id, device_fingerprint, status, started_at, expires_at, ended_at, metadata, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())`

	_, err := r.db.Exec(query,
		trial.ID,
		trial.ClientID,
		trial.DeviceFingerprint,
		trial.Status,
		trial.StartedAt,
		trial.ExpiresAt,
		trial.EndedAt,
		nullableJSON(trial.Metadata),
	)
	if err != nil {
		return fmt.Errorf("failed to create trial: %w", err)
	}
	return nil
}

func (r *TrialsRepository) GetByFingerprint(fingerprint string) (*types.Trial, error) {
	query := `
        SELECT id, client_id, device_fingerprint, status, started_at, expires_at, ended_at, metadata, created_at, updated_at
        FROM trials
        WHERE device_fingerprint = $1`

	trial := &types.Trial{}
	var metadata sql.NullString
	var endedAt sql.NullTime
	err := r.db.QueryRow(query, fingerprint).Scan(
		&trial.ID,
		&trial.ClientID,
		&trial.DeviceFingerprint,
		&trial.Status,
		&trial.StartedAt,
		&trial.ExpiresAt,
		&endedAt,
		&metadata,
		&trial.CreatedAt,
		&trial.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get trial: %w", err)
	}
	if metadata.Valid {
		trial.Metadata = []byte(metadata.String)
	}
	if endedAt.Valid {
		trial.EndedAt = &endedAt.Time
	}
	return trial, nil
}

func (r *TrialsRepository) UpdateStatus(id uuid.UUID, status types.TrialStatus, endedAt *time.Time) error {
	query := `
        UPDATE trials
        SET status = $2, ended_at = $3, updated_at = NOW()
        WHERE id = $1`

	_, err := r.db.Exec(query, id, status, endedAt)
	if err != nil {
		return fmt.Errorf("failed to update trial status: %w", err)
	}
	return nil
}

func (r *TrialsRepository) List(limit, offset int) ([]*types.Trial, error) {
	query := `
        SELECT id, client_id, device_fingerprint, status, started_at, expires_at, ended_at, metadata, created_at, updated_at
        FROM trials
        ORDER BY created_at DESC
        LIMIT $1 OFFSET $2`

	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list trials: %w", err)
	}
	defer rows.Close()

	var trials []*types.Trial
	for rows.Next() {
		trial := &types.Trial{}
		var metadata sql.NullString
		var endedAt sql.NullTime
		if err := rows.Scan(
			&trial.ID,
			&trial.ClientID,
			&trial.DeviceFingerprint,
			&trial.Status,
			&trial.StartedAt,
			&trial.ExpiresAt,
			&endedAt,
			&metadata,
			&trial.CreatedAt,
			&trial.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan trial: %w", err)
		}
		if metadata.Valid {
			trial.Metadata = []byte(metadata.String)
		}
		if endedAt.Valid {
			trial.EndedAt = &endedAt.Time
		}
		trials = append(trials, trial)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return trials, nil
}

func (r *TrialsRepository) DeleteOlderThan(threshold time.Time) error {
	query := `DELETE FROM trials WHERE expires_at < $1`
	if _, err := r.db.Exec(query, threshold); err != nil {
		return fmt.Errorf("failed to delete old trials: %w", err)
	}
	return nil
}

func (r *TrialsRepository) GetByID(id uuid.UUID) (*types.Trial, error) {
	query := `
        SELECT id, client_id, device_fingerprint, status, started_at, expires_at, ended_at, metadata, created_at, updated_at
        FROM trials
        WHERE id = $1`

	trial := &types.Trial{}
	var metadata sql.NullString
	var endedAt sql.NullTime
	if err := r.db.QueryRow(query, id).Scan(
		&trial.ID,
		&trial.ClientID,
		&trial.DeviceFingerprint,
		&trial.Status,
		&trial.StartedAt,
		&trial.ExpiresAt,
		&endedAt,
		&metadata,
		&trial.CreatedAt,
		&trial.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get trial: %w", err)
	}

	if metadata.Valid {
		trial.Metadata = []byte(metadata.String)
	}
	if endedAt.Valid {
		trial.EndedAt = &endedAt.Time
	}
	return trial, nil
}

func (r *TrialsRepository) Update(trial *types.Trial) error {
	query := `
        UPDATE trials
        SET status = $2, expires_at = $3, ended_at = $4, metadata = $5, updated_at = NOW()
        WHERE id = $1`

	_, err := r.db.Exec(query,
		trial.ID,
		trial.Status,
		trial.ExpiresAt,
		trial.EndedAt,
		nullableJSON(trial.Metadata),
	)
	if err != nil {
		return fmt.Errorf("failed to update trial: %w", err)
	}
	return nil
}

func (r *TrialsRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM trials WHERE id = $1`
	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete trial: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to determine deleted rows: %w", err)
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}
