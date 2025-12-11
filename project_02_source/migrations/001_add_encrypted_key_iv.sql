-- Migration: Add encrypted_key_iv column to notes table
-- Date: 2025-12-09
-- Purpose: Support proper key encryption with IV

-- Add new column
ALTER TABLE notes ADD COLUMN encrypted_key_iv TEXT;

-- For existing notes without encrypted_key_iv, they cannot be decrypted
-- Users need to re-upload their notes with the new encryption scheme
-- Or you can set a dummy value and mark them as "legacy" notes

-- Update existing notes with a placeholder (optional)
-- UPDATE notes SET encrypted_key_iv = 'MIGRATION_REQUIRED' WHERE encrypted_key_iv IS NULL;

-- Make it required for new notes
-- ALTER TABLE notes MODIFY encrypted_key_iv TEXT NOT NULL;
