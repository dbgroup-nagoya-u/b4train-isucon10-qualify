CREATE SCHEMA IF NOT EXISTS isuumo;

CREATE EXTENSION IF NOT EXISTS postgis;
CREATE EXTENSION IF NOT EXISTS pg_stat_statements;

DROP TABLE IF EXISTS isuumo.estate CASCADE;

CREATE TABLE isuumo.estate
(
  id
    INT4
    NOT NULL
    PRIMARY KEY,
  name
    VARCHAR(64)
    NOT NULL,
  description
    VARCHAR(4096)
    NOT NULL,
  thumbnail
    VARCHAR(128)
    NOT NULL,
  address
    VARCHAR(128)
    NOT NULL,
  latitude
    FLOAT8
    NOT NULL,
  longitude
    FLOAT8
    NOT NULL,
  rent
    INT4
    NOT NULL,
  door_height
    INT4
    NOT NULL,
  door_width
    INT4
    NOT NULL,
  features
    VARCHAR(64)
    NOT NULL,
  popularity
    INT4
    NOT NULL
);

DROP TABLE IF EXISTS isuumo.chair CASCADE;

CREATE TABLE isuumo.chair
(
  id
    INT4
    NOT NULL
    PRIMARY KEY,
  name
    VARCHAR(64)
    NOT NULL,
  description
    VARCHAR(4096)
    NOT NULL,
  thumbnail
    VARCHAR(128)
    NOT NULL,
  price
    INT4
    NOT NULL,
  height
    INT4
    NOT NULL,
  width
    INT4
    NOT NULL,
  depth
    INT4
    NOT NULL,
  color
    VARCHAR(64)
    NOT NULL,
  features
    VARCHAR(64)
    NOT NULL,
  kind
    VARCHAR(64)
    NOT NULL,
  popularity
    INT4
    NOT NULL,
  stock
    INT4
    NOT NULL
);
