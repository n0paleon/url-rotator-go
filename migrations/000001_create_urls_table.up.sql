CREATE TABLE urls (
                      id SERIAL PRIMARY KEY,
                      shortcode VARCHAR(255) NOT NULL,
                      total_hit INT DEFAULT 0,
                      original TEXT NOT NULL,
                      created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                      updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
