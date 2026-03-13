CREATE TABLE reactions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id TEXT NOT NULL,
    reaction TEXT NOT NULL,
    createdAt  INTEGER DEFAULT(unixepoch('subsec') * 1000)
);

CREATE TABLE reaction_counts (
    reaction TEXT PRIMARY KEY,
    count INTEGER NOT NULL
);
