INSERT INTO profile.languages (code, name) VALUES
  ('en','English'),
  ('ru','Русский'),
  ('es','Español'),
  ('fr','Français'),
  ('zh','中文'),
ON CONFLICT (code) DO NOTHING;
