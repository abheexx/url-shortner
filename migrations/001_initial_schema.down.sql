-- Drop triggers and functions
DROP TRIGGER IF EXISTS trigger_update_click_stats ON click_events;
DROP FUNCTION IF EXISTS update_click_stats();

-- Drop tables in reverse order
DROP TABLE IF EXISTS click_events;
DROP TABLE IF EXISTS click_stats;
DROP TABLE IF EXISTS short_urls;
