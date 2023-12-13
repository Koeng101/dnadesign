-- name: GetEntry :one
SELECT entry FROM uniprot WHERE id = ?;

-- name: InsertEntry :one
INSERT INTO uniprot(id, seqhash, entry) VALUES (?, ?, ?) RETURNING *;
