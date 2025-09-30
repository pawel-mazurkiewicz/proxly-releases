DROP TRIGGER IF EXISTS update_licenses_updated_at ON licenses;
DROP FUNCTION IF EXISTS update_updated_at_column();

DROP INDEX IF EXISTS idx_audit_logs_entity;
DROP INDEX IF EXISTS idx_audit_logs_created_at;
DROP INDEX IF EXISTS idx_activations_occurred_at;
DROP INDEX IF EXISTS idx_activations_license_id;
DROP INDEX IF EXISTS idx_licenses_status;
DROP INDEX IF EXISTS idx_licenses_key;

DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS webhooks;
DROP TABLE IF EXISTS activations;
DROP TABLE IF EXISTS licenses;

DROP EXTENSION IF EXISTS "uuid-ossp";