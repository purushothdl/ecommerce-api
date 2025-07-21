   -- migrations/000006_add_updated_at_to_categories.down.sql
   ALTER TABLE categories
   DROP COLUMN updated_at;