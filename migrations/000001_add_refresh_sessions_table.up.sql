-- Create the refresh_sessions table
CREATE TABLE refresh_sessions
(
    id         VARCHAR(255)             NOT NULL
        PRIMARY KEY,
    user_id    UUID                     NOT NULL,
    token_hash VARCHAR(255)             NOT NULL,
    user_ip    VARCHAR(45)              NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL
);
