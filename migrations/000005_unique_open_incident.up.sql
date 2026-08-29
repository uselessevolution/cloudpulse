CREATE UNIQUE INDEX idx_incidents_one_open_per_service
    ON incidents (service_id)
    WHERE status = 'OPEN';