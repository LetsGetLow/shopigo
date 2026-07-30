CREATE TABLE IF NOT EXISTS catalog_categories (
    category_id UUID PRIMARY KEY,
    parent_id UUID REFERENCES catalog_categories (category_id),
    name TEXT NOT NULL,
    description TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by UUID NOT NULL,
    updated_at TIMESTAMPTZ NULL,
    updated_by UUID NULL,
    deleted_at TIMESTAMPTZ NULL,
    deleted_by UUID NULL
);
