-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS pr_reviewers (
                                            pull_request_id VARCHAR(255) NOT NULL,
                                            reviewer_id VARCHAR(255) NOT NULL,
                                            assigned_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                                            PRIMARY KEY (pull_request_id, reviewer_id),
                                            CONSTRAINT fk_pr_reviewers_pr FOREIGN KEY (pull_request_id) REFERENCES pull_requests(id) ON DELETE CASCADE,
                                            CONSTRAINT fk_pr_reviewers_reviewer FOREIGN KEY (reviewer_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_pr_reviewers_reviewer_id ON pr_reviewers(reviewer_id);
CREATE INDEX IF NOT EXISTS idx_pr_reviewers_pull_request_id ON pr_reviewers(pull_request_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS pr_reviewers;
DROP INDEX IF EXISTS idx_pr_reviewers_pull_request_id;
DROP INDEX IF EXISTS idx_pr_reviewers_reviewer_id;
-- +goose StatementEnd
