-- Données de démonstration.
INSERT INTO users (email, password_hash, name)
VALUES (
        'alice@example.com',
        '$2a$10$Op6Ix/7GBsHRsqi5JIetle1roHfCiX39IkuRvj24EEtyqExmhWVzG',
        'Alice'
    ),
    (
        'bob@example.com',
        '$2a$10$xXhQrNtQhpZPnGILc5ttieutfZrwwrI8Wj2JSOmDoQFksufFSz6ny',
        'Bob'
    );
-- Espaces et notes d'Alice
INSERT INTO spaces (user_id, name, description)
SELECT id,
    'Devoirs',
    'Travaux scolaires et rendus à préparer'
FROM users
WHERE email = 'alice@example.com';
INSERT INTO spaces (user_id, name, description)
SELECT id,
    'Jobs',
    'Candidatures et recherches d''alternance'
FROM users
WHERE email = 'alice@example.com';
INSERT INTO notes (space_id, title, content, status)
SELECT s.id,
    'Test technique Go',
    'Développer le serveur et le client de l''application de notes.',
    'in_progress'
FROM spaces s
    JOIN users u ON u.id = s.user_id
WHERE u.email = 'alice@example.com'
    AND s.name = 'Devoirs';
INSERT INTO notes (space_id, title, content, status)
SELECT s.id,
    'Réviser les bases de données',
    'Revoir les jointures, les index et les transactions.',
    'todo'
FROM spaces s
    JOIN users u ON u.id = s.user_id
WHERE u.email = 'alice@example.com'
    AND s.name = 'Devoirs';
INSERT INTO notes (space_id, title, content, status)
SELECT s.id,
    'Mettre à jour le CV',
    'Ajouter les projets réalisés cette année.',
    'done'
FROM spaces s
    JOIN users u ON u.id = s.user_id
WHERE u.email = 'alice@example.com'
    AND s.name = 'Jobs';
-- Espace et note de Bob (ne doivent jamais apparaître pour Alice)
INSERT INTO spaces (user_id, name, description)
SELECT id,
    'Personnel',
    'Notes privées de Bob'
FROM users
WHERE email = 'bob@example.com';
INSERT INTO notes (space_id, title, content, status)
SELECT s.id,
    'Liste de courses',
    'Pain, café, fruits.',
    'todo'
FROM spaces s
    JOIN users u ON u.id = s.user_id
WHERE u.email = 'bob@example.com'
    AND s.name = 'Personnel';