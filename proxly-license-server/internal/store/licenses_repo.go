package store

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/pawel-mazurkiewicz/proxly-license-server/pkg/types"
)

type LicensesRepository struct {
	db *sql.DB
}

func NewLicensesRepository(db *sql.DB) *LicensesRepository {
	return &LicensesRepository{db: db}
}

func (r *LicensesRepository) Create(license *types.License) error {
        query := `
                INSERT INTO licenses (id, key, max_activations, activation_count, status, metadata, created_at, updated_at)
                VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

        _, err := r.db.Exec(query,
                license.ID,
                license.Key,
                license.MaxActivations,
                license.ActivationCount,
                license.Status,
                nullableJSON(license.Metadata),
                license.CreatedAt,
                license.UpdatedAt,
        )
        if err != nil {
                return fmt.Errorf("failed to create license: %w", err)
	}

	return nil
}

func (r *LicensesRepository) GetByID(id uuid.UUID) (*types.License, error) {
	query := `
		SELECT id, key, max_activations, activation_count, status, metadata, created_at, updated_at
		FROM licenses
		WHERE id = $1`

	license := &types.License{}
	var metadata sql.NullString
	err := r.db.QueryRow(query, id).Scan(
		&license.ID,
		&license.Key,
		&license.MaxActivations,
		&license.ActivationCount,
		&license.Status,
		&metadata,
		&license.CreatedAt,
		&license.UpdatedAt,
	)
	if metadata.Valid {
		license.Metadata = []byte(metadata.String)
	}
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("license not found")
		}
		return nil, fmt.Errorf("failed to get license: %w", err)
	}

	return license, nil
}

func (r *LicensesRepository) GetByKey(key string) (*types.License, error) {
	query := `
		SELECT id, key, max_activations, activation_count, status, metadata, created_at, updated_at
		FROM licenses
		WHERE key = $1`

	license := &types.License{}
	var metadata sql.NullString
	err := r.db.QueryRow(query, key).Scan(
		&license.ID,
		&license.Key,
		&license.MaxActivations,
		&license.ActivationCount,
		&license.Status,
		&metadata,
		&license.CreatedAt,
		&license.UpdatedAt,
	)
	if metadata.Valid {
		license.Metadata = []byte(metadata.String)
	}
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("license not found")
		}
		return nil, fmt.Errorf("failed to get license: %w", err)
	}

	return license, nil
}

func (r *LicensesRepository) Update(license *types.License) error {
        query := `
                UPDATE licenses
                SET key = $2, max_activations = $3, activation_count = $4, status = $5, metadata = $6, updated_at = $7
                WHERE id = $1`

        result, err := r.db.Exec(query,
                license.ID,
                license.Key,
                license.MaxActivations,
                license.ActivationCount,
                license.Status,
                nullableJSON(license.Metadata),
                time.Now(),
        )
	if err != nil {
		return fmt.Errorf("failed to update license: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("license not found")
	}

	return nil
}

func (r *LicensesRepository) SetActivationCount(id uuid.UUID, count int) error {
	query := `
		UPDATE licenses
		SET activation_count = $2, updated_at = NOW()
		WHERE id = $1`

	result, err := r.db.Exec(query, id, count)
	if err != nil {
		return fmt.Errorf("failed to set activation count: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("license not found")
	}

	return nil
}

func (r *LicensesRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM licenses WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete license: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("license not found")
	}

	return nil
}

func (r *LicensesRepository) List(limit, offset int) ([]*types.License, error) {
	query := `
		SELECT id, key, max_activations, activation_count, status, metadata, created_at, updated_at
		FROM licenses
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list licenses: %w", err)
	}
	defer rows.Close()

	var licenses []*types.License
	for rows.Next() {
		license := &types.License{}
		var metadata sql.NullString
		err := rows.Scan(
			&license.ID,
			&license.Key,
			&license.MaxActivations,
			&license.ActivationCount,
			&license.Status,
			&metadata,
			&license.CreatedAt,
			&license.UpdatedAt,
		)
		if metadata.Valid {
			license.Metadata = []byte(metadata.String)
		}
		if err != nil {
			return nil, fmt.Errorf("failed to scan license: %w", err)
		}
		licenses = append(licenses, license)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return licenses, nil
}

func (r *LicensesRepository) IncrementActivationCount(id uuid.UUID) error {
	query := `
		UPDATE licenses
		SET activation_count = activation_count + 1, updated_at = NOW()
		WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to increment activation count: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("license not found")
	}

	return nil
}

type ActivationsRepository struct {
	db *sql.DB
}

func NewActivationsRepository(db *sql.DB) *ActivationsRepository {
	return &ActivationsRepository{db: db}
}

func (r *ActivationsRepository) Create(activation *types.Activation) error {
	query := `
		INSERT INTO activations (id, license_id, client_id, device_fingerprint, ip, user_agent, occurred_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.db.Exec(query,
		activation.ID,
		activation.LicenseID,
		activation.ClientID,
		activation.DeviceFingerprint,
		activation.IP,
		activation.UserAgent,
		activation.OccurredAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create activation: %w", err)
	}

	return nil
}

func (r *ActivationsRepository) GetByLicenseID(licenseID uuid.UUID) ([]*types.Activation, error) {
	query := `
		SELECT id, license_id, client_id, device_fingerprint, ip, user_agent, occurred_at
		FROM activations
		WHERE license_id = $1
		ORDER BY occurred_at DESC`

	rows, err := r.db.Query(query, licenseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get activations: %w", err)
	}
	defer rows.Close()

	var activations []*types.Activation
	for rows.Next() {
		activation := &types.Activation{}
		err := rows.Scan(
			&activation.ID,
			&activation.LicenseID,
			&activation.ClientID,
			&activation.DeviceFingerprint,
			&activation.IP,
			&activation.UserAgent,
			&activation.OccurredAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan activation: %w", err)
		}
		activations = append(activations, activation)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return activations, nil
}

func (r *ActivationsRepository) DeleteByLicenseID(licenseID uuid.UUID) error {
	query := `DELETE FROM activations WHERE license_id = $1`

	if _, err := r.db.Exec(query, licenseID); err != nil {
		return fmt.Errorf("failed to delete activations: %w", err)
	}

	return nil
}

func (r *ActivationsRepository) GetByLicenseIDAndFingerprint(licenseID uuid.UUID, fingerprint string) (*types.Activation, error) {
	query := `
		SELECT id, license_id, client_id, device_fingerprint, ip, user_agent, occurred_at
		FROM activations
		WHERE license_id = $1 AND device_fingerprint = $2
		LIMIT 1`

	activation := &types.Activation{}
	var clientID, deviceFP sql.NullString
	err := r.db.QueryRow(query, licenseID, fingerprint).Scan(
		&activation.ID,
		&activation.LicenseID,
		&clientID,
		&deviceFP,
		&activation.IP,
		&activation.UserAgent,
		&activation.OccurredAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get activation: %w", err)
	}

	if clientID.Valid {
		activation.ClientID = new(string)
		*activation.ClientID = clientID.String
	}
	if deviceFP.Valid {
		activation.DeviceFingerprint = new(string)
		*activation.DeviceFingerprint = deviceFP.String
	}

	return activation, nil
}

func (r *ActivationsRepository) HasActivationForFingerprint(fingerprint string) (bool, error) {
	query := `SELECT 1 FROM activations WHERE device_fingerprint = $1 LIMIT 1`
	var dummy int
	err := r.db.QueryRow(query, fingerprint).Scan(&dummy)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("failed to check activations: %w", err)
	}
	return true, nil
}

type WebhooksRepository struct {
	db *sql.DB
}

func NewWebhooksRepository(db *sql.DB) *WebhooksRepository {
	return &WebhooksRepository{db: db}
}

func (r *WebhooksRepository) Create(webhook *types.Webhook) error {
	query := `
		INSERT INTO webhooks (id, url, secret, events)
		VALUES ($1, $2, $3, $4)`

	_, err := r.db.Exec(query,
		webhook.ID,
		webhook.URL,
		webhook.Secret,
		pq.Array(webhook.Events),
	)
	if err != nil {
		return fmt.Errorf("failed to create webhook: %w", err)
	}

	return nil
}

func (r *WebhooksRepository) GetAll() ([]*types.Webhook, error) {
	query := `
		SELECT id, url, secret, events
		FROM webhooks`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get webhooks: %w", err)
	}
	defer rows.Close()

	var webhooks []*types.Webhook
	for rows.Next() {
		webhook := &types.Webhook{}
		err := rows.Scan(
			&webhook.ID,
			&webhook.URL,
			&webhook.Secret,
			pq.Array(&webhook.Events),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan webhook: %w", err)
		}
		webhooks = append(webhooks, webhook)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return webhooks, nil
}

type AuditLogsRepository struct {
	db *sql.DB
}

func NewAuditLogsRepository(db *sql.DB) *AuditLogsRepository {
	return &AuditLogsRepository{db: db}
}

func (r *AuditLogsRepository) Create(log *types.AuditLog) error {
	query := `
		INSERT INTO audit_logs (id, actor, action, entity, payload, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := r.db.Exec(query,
		log.ID,
		log.Actor,
		log.Action,
		log.Entity,
		log.Payload,
		log.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	return nil
}
