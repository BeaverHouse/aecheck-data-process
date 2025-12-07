-- AECheck Database Schema
-- Migrated from existing aecheck-backend database
-- Schema: aecheck
-- Generated for sqlc compatibility with tinyclover-ae-analyzer structure

-- Create schema
CREATE SCHEMA IF NOT EXISTS aecheck;

-- Characters table
CREATE TABLE IF NOT EXISTS aecheck.characters (
    character_id VARCHAR(20) NOT NULL,
    character_code VARCHAR(20) NOT NULL,
    category VARCHAR(20) NOT NULL,
    style VARCHAR(10) NOT NULL,
    light_shadow VARCHAR(10) NOT NULL,
    max_manifest INT4 NOT NULL,
    is_awaken BOOL NOT NULL,
    is_alter BOOL NOT NULL,
    alter_character VARCHAR(20) NULL,
    seesaa_url VARCHAR(500) NULL,
    aewiki_url VARCHAR(500) NULL,
    update_date DATE NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NULL,
    deleted_at TIMESTAMP NULL,
    custom_manifest BOOL DEFAULT false NULL,
    personalities_data JSONB DEFAULT '[]'::jsonb NULL,
    dungeons_data JSONB DEFAULT '[]'::jsonb NULL,
    buddy_data JSONB DEFAULT NULL,
    CONSTRAINT ae_character_pk PRIMARY KEY (character_id)
);

-- Buddies table
CREATE TABLE IF NOT EXISTS aecheck.buddies (
    buddy_id VARCHAR(10) NOT NULL,
    character_id VARCHAR(10) NULL,
    get_path VARCHAR(20) NULL,
    seesaa_url VARCHAR(500) NULL,
    aewiki_url VARCHAR(500) NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NULL,
    deleted_at TIMESTAMP NULL,
    CONSTRAINT ae_buddy_pk PRIMARY KEY (buddy_id)
);

-- Dungeons table
CREATE TABLE IF NOT EXISTS aecheck.dungeons (
    dungeon_id VARCHAR(20) NOT NULL,
    altema_url VARCHAR(500) NULL,
    aewiki_url VARCHAR(500) NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NULL,
    deleted_at TIMESTAMP NULL,
    CONSTRAINT ae_dungeon_pk PRIMARY KEY (dungeon_id)
);

-- Translations table
CREATE TABLE IF NOT EXISTS aecheck.translations (
    key VARCHAR(50) NOT NULL,
    ko VARCHAR(500) NOT NULL,
    en VARCHAR(500) NOT NULL,
    ja VARCHAR(500) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NULL,
    deleted_at TIMESTAMP NULL,
    CONSTRAINT ae_i18n_pk PRIMARY KEY (key)
);

-- Critical performance indexes only
-- For JOIN operations (character -> buddies)
CREATE INDEX IF NOT EXISTS idx_buddies_character_id ON aecheck.buddies(character_id);

-- For related character lookups (character_code and alter_character queries)
CREATE INDEX IF NOT EXISTS idx_characters_character_code ON aecheck.characters(character_code);
CREATE INDEX IF NOT EXISTS idx_characters_alter_character ON aecheck.characters(alter_character) WHERE alter_character IS NOT NULL;

-- Note: No indexes on JSONB columns needed
-- GIN indexes are only useful when querying inside JSONB (e.g., WHERE personalities_data @> ...)
-- Our queries only SELECT these columns for unmarshaling, so indexes provide no benefit