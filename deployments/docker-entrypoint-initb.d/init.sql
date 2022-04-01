-- Make the table to hold our coins
CREATE TABLE IF NOT EXISTS coins (
    id text NOT NULL,
    exchanges varchar ARRAY,
    taskrun INT NOT NULL CHECK (taskrun > 0)
)