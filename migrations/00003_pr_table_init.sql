-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS pull_requests (
                                             id VARCHAR(255) PRIMARY KEY,
                                             name VARCHAR(500) NOT NULL,
                                             author_id VARCHAR(255) NOT NULL,
                                             status VARCHAR(20) NOT NULL DEFAULT 'OPEN',
                                             created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                                             merged_at TIMESTAMP,
                                             CONSTRAINT fk_pull_requests_author FOREIGN KEY (author_id) REFERENCES users(id) ON DELETE CASCADE,
                                             CONSTRAINT chk_status CHECK (status IN ('OPEN', 'MERGED'))
);

CREATE INDEX IF NOT EXISTS idx_pull_requests_author_id ON pull_requests(author_id);
CREATE INDEX IF NOT EXISTS idx_pull_requests_status ON pull_requests(status);
CREATE INDEX IF NOT EXISTS idx_pull_requests_created_at ON pull_requests(created_at);
CREATE INDEX IF NOT EXISTS idx_pull_requests_author_status ON pull_requests(author_id, status);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS pull_requests;
DROP INDEX IF EXISTS idx_pull_requests_author_status;
DROP INDEX IF EXISTS idx_pull_requests_created_at;
DROP INDEX IF EXISTS idx_pull_requests_status;
DROP INDEX IF EXISTS idx_pull_requests_author_id;
-- +goose StatementEnd
