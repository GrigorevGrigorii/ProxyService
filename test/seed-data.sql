BEGIN;

-- Clean all tables
TRUNCATE TABLE targets CASCADE;
TRUNCATE TABLE services CASCADE;

-- Fill services
INSERT INTO services (name, host, scheme, timeout, retry_count, retry_interval, version) VALUES
    ('mock', 'localhost:8081', 'http', 10, 3, 0.1, 0),
    ('thecatapi', 'api.thecatapi.com', 'https', 10, 3, 0.5, 0),
    ('spacex', 'api.spacexdata.com', 'https', 10, 3, 0.1, 0),
    ('mock-docker', 'mock:8081', 'http', 10, 3, 0.1, 0);

-- Fill targets
INSERT INTO targets (service_name, path, method, query, cache_interval) VALUES
    -- targets for service 'mock'
    ('mock', '/mock', 'GET', 'query=param', NULL),
    ('mock', '/mock', 'POST', '', NULL),
    ('mock', '/mock', 'PUT', '', NULL),
    ('mock', '/mock', 'DELETE', '', NULL),

    -- targets for service 'thecatapi'
    ('thecatapi', '/v1/images/search', 'GET', '*', NULL),

    -- targets for service 'spacex'
    ('spacex', '/v3', 'GET', '', '1m'),

    -- targets for service 'mock-docker'
    ('mock-docker', '/mock', 'GET', 'query=param', '1m'),
    ('mock-docker', '/mock', 'POST', '', NULL),
    ('mock-docker', '/mock', 'PUT', '', NULL),
    ('mock-docker', '/mock', 'DELETE', '', NULL);

COMMIT;
