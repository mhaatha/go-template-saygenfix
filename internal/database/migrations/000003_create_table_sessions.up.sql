CREATE TABLE IF NOT EXISTS sessions (
    session_id VARCHAR(100) NOT NULL,
    user_id UUID NOT NULL,
    created_at TIMESTAMP(0) WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY (session_id),
    CONSTRAINT fk_users
        FOREIGN KEY (user_id) 
        REFERENCES users (id) 
        ON DELETE CASCADE
)