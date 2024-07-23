CREATE TABLE IF NOT EXISTS users_posts
(
    id SERIAL PRIMARY KEY,
    email TEXT NOT NULL,
    bucket TEXT NOT NULL,
    key TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS subscriptions
(
    uid    integer,
    sub_id integer,
    PRIMARY KEY (uid, sub_id)
);