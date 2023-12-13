PRAGMA journal_mode = WAL;
PRAGMA synchronous = NORMAL; -- https://news.ycombinator.com/item?id=34247738
PRAGMA cache_size = 20000; -- https://news.ycombinator.com/item?id=34247738
PRAGMA foreign_keys = ON;
PRAGMA strict = ON;
PRAGMA busy_timeout = 5000;

CREATE TABLE uniprot (
    id TEXT PRIMARY KEY,
    seqhash TEXT NOT NULL,
    entry TEXT NOT NULL
);
