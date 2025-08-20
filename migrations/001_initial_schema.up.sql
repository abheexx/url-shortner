-- Create short_urls table
CREATE TABLE short_urls (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(16) UNIQUE NOT NULL,
    long_url TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expire_at TIMESTAMPTZ NULL,
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    custom_alias BOOLEAN NOT NULL DEFAULT FALSE,
    created_by VARCHAR(255) NULL,
    metadata JSONB NULL
);

-- Create click_stats table
CREATE TABLE click_stats (
    code VARCHAR(16) PRIMARY KEY REFERENCES short_urls(code) ON DELETE CASCADE,
    total_clicks BIGINT NOT NULL DEFAULT 0,
    last_access_at TIMESTAMPTZ NULL,
    first_access_at TIMESTAMPTZ NULL
);

-- Create click_events table for analytics
CREATE TABLE click_events (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(16) NOT NULL REFERENCES short_urls(code) ON DELETE CASCADE,
    ts TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    user_agent TEXT NULL,
    ip_address INET NULL,
    referer TEXT NULL,
    country VARCHAR(2) NULL,
    device_type VARCHAR(20) NULL
);

-- Create indexes for performance
CREATE INDEX idx_short_urls_code ON short_urls(code);
CREATE INDEX idx_short_urls_expire_at ON short_urls(expire_at) WHERE expire_at IS NOT NULL;
CREATE INDEX idx_short_urls_is_deleted ON short_urls(is_deleted);
CREATE INDEX idx_click_events_code_ts ON click_events(code, ts);
CREATE INDEX idx_click_events_ts ON click_events(ts);

-- Create function to update click stats
CREATE OR REPLACE FUNCTION update_click_stats()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO click_stats (code, total_clicks, last_access_at, first_access_at)
    VALUES (NEW.code, 1, NEW.ts, NEW.ts)
    ON CONFLICT (code) DO UPDATE SET
        total_clicks = click_stats.total_clicks + 1,
        last_access_at = NEW.ts;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger to automatically update click stats
CREATE TRIGGER trigger_update_click_stats
    AFTER INSERT ON click_events
    FOR EACH ROW
    EXECUTE FUNCTION update_click_stats();
