-- AI-assisted excerpt provenance and review tracking

ALTER TABLE posts
  ADD COLUMN IF NOT EXISTS excerpt_source text NOT NULL DEFAULT 'empty'
    CHECK (excerpt_source IN ('manual', 'ai', 'empty')),
  ADD COLUMN IF NOT EXISTS excerpt_reviewed_body_hash text NOT NULL DEFAULT '';

UPDATE posts
SET excerpt_source = CASE
  WHEN btrim(excerpt) = '' THEN 'empty'
  ELSE 'manual'
END
WHERE excerpt_source = 'empty';