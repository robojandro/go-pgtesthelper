CREATE TABLE IF NOT EXISTS dogbreed (
    id    bigint GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
    name  character varying(50) NOT NULL
);
