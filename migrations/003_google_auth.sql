-- Connexion via Google
ALTER TABLE users
ADD COLUMN google_id VARCHAR(255);
ALTER TABLE users
ALTER COLUMN password_hash DROP NOT NULL;
CREATE UNIQUE INDEX idx_users_google_id ON users(google_id);
COMMENT ON COLUMN users.google_id IS 'Identifiant Google stable (claim « sub »). NULL pour un compte créé par mot de passe.';
COMMENT ON COLUMN users.password_hash IS 'Empreinte bcrypt. NULL pour un compte créé via Google et sans mot de passe.';