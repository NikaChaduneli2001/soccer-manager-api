-- Add fullname and age to users (for DBs created before these columns existed)
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'users' AND column_name = 'fullname') THEN
    ALTER TABLE users ADD COLUMN fullname VARCHAR(255) NOT NULL DEFAULT '';
  END IF;
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'users' AND column_name = 'age') THEN
    ALTER TABLE users ADD COLUMN age INT NOT NULL DEFAULT 0 CHECK (age >= 0 AND age <= 150);
  END IF;
END $$;
