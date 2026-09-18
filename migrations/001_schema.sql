-- Schéma initial de la base de données.
-- Utilisateurs
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(100) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON COLUMN users.email IS 'Stocké en minuscules par l''application : la contrainte UNIQUE devient ainsi insensible à la casse.';
COMMENT ON COLUMN users.password_hash IS 'Empreinte bcrypt. Le mot de passe en clair n''est jamais stocké.';
-- Espaces
CREATE TABLE spaces (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_spaces_user_id ON spaces(user_id);
CREATE UNIQUE INDEX idx_spaces_user_id_name ON spaces(user_id, LOWER(name));
-- Notes
CREATE TABLE notes (
    id BIGSERIAL PRIMARY KEY,
    space_id BIGINT NOT NULL REFERENCES spaces(id) ON DELETE CASCADE,
    title VARCHAR(200) NOT NULL,
    content TEXT NOT NULL DEFAULT '',
    status VARCHAR(20) NOT NULL DEFAULT 'todo',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    -- Les trois états possibles d'une note sont contraints par la base.
    -- Une valeur inattendue est rejetée même si la validation applicative
    -- venait à être contournée.
    CONSTRAINT notes_status_valid CHECK (status IN ('todo', 'in_progress', 'done'))
);
CREATE INDEX idx_notes_space_id ON notes(space_id);