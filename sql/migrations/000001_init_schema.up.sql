BEGIN;

-- Generic "updated_at" trigger function
CREATE OR REPLACE FUNCTION trigger_set_timestamp()
RETURNS TRIGGER AS $$

BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Books
CREATE TABLE "books" (
    "id" bigserial UNIQUE NOT NULL PRIMARY KEY,
    "code" varchar UNIQUE NOT NULL,
    "name" varchar NOT NULL,
    "auther" varchar,
    "is_archived" boolean NOT NULL DEFAULT FALSE,
    "created_at" timestamptz NOT NULL DEFAULT NOW(),
    "updated_at" timestamptz NOT NULL DEFAULT NOW()
);
CREATE TRIGGER set_timestamp
BEFORE UPDATE ON books
FOR EACH ROW
EXECUTE FUNCTION trigger_set_timestamp();

COMMIT;