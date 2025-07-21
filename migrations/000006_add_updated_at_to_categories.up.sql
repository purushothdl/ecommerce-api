   -- migrations/000006_add_updated_at_to_categories.up.sql
   ALTER TABLE categories
   ADD COLUMN updated_at timestamp(0) with time zone NOT NULL DEFAULT NOW();