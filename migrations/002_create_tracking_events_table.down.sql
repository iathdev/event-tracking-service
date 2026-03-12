SELECT remove_retention_policy('tracking_events', if_exists => TRUE);
DROP TABLE IF EXISTS tracking_events;
