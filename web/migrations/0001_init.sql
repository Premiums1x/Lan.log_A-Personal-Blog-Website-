-- lancer.log initial schema
-- deploy: psql -f migrations/0001_init.sql

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ===== admin users =====
CREATE TABLE IF NOT EXISTS users (
  id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  username      text NOT NULL UNIQUE,
  password_hash text NOT NULL,
  display_name  text NOT NULL DEFAULT '',
  created_at    timestamptz NOT NULL DEFAULT now()
);

-- ===== posts =====
CREATE TABLE IF NOT EXISTS posts (
  id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  slug          text NOT NULL UNIQUE,
  title         text NOT NULL,
  excerpt       text NOT NULL DEFAULT '',
  body_md       text NOT NULL DEFAULT '',
  body_html     text NOT NULL DEFAULT '',
  cover_url     text NOT NULL DEFAULT '',
  section       text NOT NULL DEFAULT 'posts',
  status        text NOT NULL DEFAULT 'draft',
  commit_hash   text NOT NULL DEFAULT '',
  read_minutes  integer NOT NULL DEFAULT 1,
  words         integer NOT NULL DEFAULT 0,
  pinned        boolean NOT NULL DEFAULT false,
  published_at  timestamptz,
  created_at    timestamptz NOT NULL DEFAULT now(),
  updated_at    timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS posts_status_published_idx ON posts (published_at DESC) WHERE status = 'published';
CREATE INDEX IF NOT EXISTS posts_section_idx ON posts (section);

-- ===== tags =====
CREATE TABLE IF NOT EXISTS tags (
  id   uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  slug text NOT NULL UNIQUE,
  name text NOT NULL
);
CREATE TABLE IF NOT EXISTS post_tags (
  post_id uuid NOT NULL REFERENCES posts (id) ON DELETE CASCADE,
  tag_id  uuid NOT NULL REFERENCES tags (id)  ON DELETE CASCADE,
  PRIMARY KEY (post_id, tag_id)
);

-- ===== site settings / content sections (jsonb) =====
-- key: hero / about / stack / now / footer / nav / branding / archive / shelf
CREATE TABLE IF NOT EXISTS settings (
  section_key text PRIMARY KEY,
  value       jsonb NOT NULL DEFAULT '{}'::jsonb,
  updated_at   timestamptz NOT NULL DEFAULT now()
);

-- ===== default content =====
INSERT INTO settings (section_key, value) VALUES
('branding', '{"brand":"lancer.log","footer_tag":"Frontend practice, agent notes, and backend experiments.","since_year":2026,"commit_hash":"a1f3c2d","build_badge":"BUILT QUIETLY / NO TRACKING / NO ADS"}'::jsonb),
('hero', '{"eyebrow_cmd":"cat about-this-blog.md","title":"A learning log for frontend work,","title_accent":"agents","title_tail":"and backend experiments.","sub":"Notes from frontend internship work, React practice, Go and backend experiments, and the slow process of turning concepts into working projects.","meta":[{"k":"role","v":"frontend intern"},{"k":"major","v":"computer science"},{"k":"next","v":"agent / backend"},{"k":"ads","v":"never"}],"corner":[{"label":"focus","val":"react / go"},{"label":"mode","val":"learning in public"},{"label":"build","val":"passing"}]}'::jsonb),
('nav', '{"links":[{"label":"posts","href":"/"},{"label":"about","href":"/about"},{"label":"archive","href":"/archive"},{"label":"shelf","href":"/shelf"}]}'::jsonb),
('stack', '{"cells":[{"ic":"01","title":"React","desc":"Components, state, effects, data flow, and enough practice to stop freezing at a blank editor."},{"ic":"02","title":"Go","desc":"Backend basics, APIs, database work, and small services that make frontend work less mysterious."},{"ic":"03","title":"Agents","desc":"Notes on tool use, workflows, and how AI coding assistants change daily engineering practice."},{"ic":"04","title":"Internship","desc":"Frontend implementation details, review notes, and the rough edges that only appear in real projects."}]}'::jsonb),
('about', '{"title":"A CS student moving from frontend toward agents and backend.","title_accent":"agents","intro":["I am a computer science undergraduate about to enter senior year, currently working as a frontend intern.","This blog records the process of turning docs, internships, small projects, and self-study into practical engineering judgment."],"meta":[{"k":"role","v":"frontend intern"},{"k":"major","v":"computer science"},{"k":"next","v":"agent / backend"}],"bio_yml":{"name":"Lan","role":"frontend intern / CS student","stack":"react / go / agent","writes":"learning in public","based":"china","hosting":"self-hosted blog"},"uptime":"still learning","body_md":"## About me\n\nI am in that useful stage where concepts are starting to make sense, but real implementation still exposes gaps. This blog keeps those gaps visible.\n\n## About this site\n\nThis is not a resume and not a tutorial archive. It is an engineering log for frontend internship notes, React practice, Go and backend experiments, agent ideas, and route changes in my learning plan.\n\n## Interests\n\nI like Max Verstappen consistency and attack, and Stephen Curry way of turning long practice into instinct. Coding has a bit of that too: repeat the basics, then stay calm in complex situations."}'::jsonb),
('now', '{"lines":[{"is_cmd":true,"f":"cat","args":"now.txt","c":"# current focus"},{"is_cmd":false,"arrow":"->","k":"internship","v":"frontend","is_string":true},{"is_cmd":false,"arrow":"->","k":"learning","v":"react / go / agents","is_string":true},{"is_cmd":false,"arrow":"->","k":"next","v":"ship small projects","is_string":true}]}'::jsonb),
('footer', '{"cols":[{"h":"browse","links":[{"label":"posts","href":"/"},{"label":"archive","href":"/archive"},{"label":"tags","href":"/tags"},{"label":"shelf","href":"/shelf"}]},{"h":"about","links":[{"label":"whoami","href":"/about"}]}]}'::jsonb),
('archive', '{"eyebrow_cmd":"git log --reverse --stat","title":"Archive","title_accent":"Archive","intro":"All published posts, grouped by time. This is a long-running learning log for frontend work, agents, backend practice, and ideas still taking shape.","meta":[{"k":"mode","v":"timeline"},{"k":"order","v":"newest first"},{"k":"status","v":"published only"}]}'::jsonb),
('shelf', '{"eyebrow_cmd":"ls shelf/learning-stack","title":"Shelf","title_accent":"Shelf","intro":"A living shelf for books, docs, tools, and courses I am reading, using, or planning to study more deeply.","meta":[{"k":"type","v":"books / tools / courses"},{"k":"update","v":"manual"},{"k":"bias","v":"small useful things"}],"groups":[{"title":"Learning","eyebrow":"reading queue","desc":"Resources that strengthen the engineering base.","items":[{"title":"React Docs","desc":"Rebuild muscle memory around components, state, effects, and data flow.","meta":"frontend","href":"https://react.dev/","status":"re-reading","tags":["react","frontend"]},{"title":"Go by Example","desc":"Small examples for Go syntax and standard library practice.","meta":"backend","href":"https://gobyexample.com/","status":"active","tags":["go","backend"]}]},{"title":"Tools","eyebrow":"daily kit","desc":"Tools used in internship work and personal projects.","items":[{"title":"VS Code / Cursor","desc":"Main entry for frontend work, reading Go projects, and agent collaboration.","meta":"editor","href":"#","status":"daily","tags":["editor","agent"]},{"title":"Postman / Apifox","desc":"API debugging and frontend/backend integration practice.","meta":"api","href":"#","status":"daily","tags":["api","backend"]}]}]}'::jsonb)
ON CONFLICT (section_key) DO NOTHING;

-- Clean legacy subscribe/contact surfaces if this migration runs on an older DB.
UPDATE settings
SET value = jsonb_set(value - 'cta_label' - 'cta_href', '{links}', COALESCE(value->'links', '[]'::jsonb)), updated_at = now()
WHERE section_key = 'nav'
  AND ((value->>'cta_label') = 'subscribe' OR (value->>'cta_href') = '/about#contact');

UPDATE settings
SET value = jsonb_build_object(
  'cols', COALESCE((
    SELECT jsonb_agg(
      jsonb_set(col.item - 'links', '{links}', COALESCE((
        SELECT jsonb_agg(link.item)
        FROM jsonb_array_elements(COALESCE(col.item->'links', '[]'::jsonb)) AS link(item)
        WHERE link.item->>'href' <> '/about#contact'
          AND lower(COALESCE(link.item->>'label', '')) NOT LIKE '%subscribe%'
      ), '[]'::jsonb))
    )
    FROM jsonb_array_elements(COALESCE(value->'cols', '[]'::jsonb)) AS col(item)
    WHERE lower(COALESCE(col.item->>'h', '')) <> 'subscribe'
  ), '[]'::jsonb)
), updated_at = now()
WHERE section_key = 'footer';
