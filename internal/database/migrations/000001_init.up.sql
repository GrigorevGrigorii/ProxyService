CREATE TYPE httpscheme AS ENUM('http', 'https');
CREATE TYPE httpmethod AS ENUM('GET', 'POST', 'PUT', 'DELETE');

CREATE TABLE services(
    name VARCHAR(32) PRIMARY KEY,
    host VARCHAR(128) UNIQUE NOT NULL,
    scheme httpscheme NOT NULL,
    timeout NUMERIC(4, 2) NOT NULL,
    retry_count INTEGER DEFAULT 0 NOT NULL,
    retry_interval NUMERIC(4, 2) DEFAULT 0.0 NOT NULL
);

CREATE TABLE targets(
    service_name VARCHAR(32) REFERENCES services(name) ON DELETE CASCADE,
    path VARCHAR(128),
    method httpmethod,
    PRIMARY KEY (service_name, path, method)
);
