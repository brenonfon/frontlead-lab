-- Add phone column to existing contacts table (if it doesn't exist)
ALTER TABLE contacts 
ADD COLUMN IF NOT EXISTS phone TEXT;

-- Create index on phone column for better query performance
CREATE INDEX IF NOT EXISTS idx_contacts_phone ON contacts(phone);
