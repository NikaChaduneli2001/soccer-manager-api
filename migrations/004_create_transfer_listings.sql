-- Transfer listings: player on market with asking price (one listing per player)
CREATE TABLE IF NOT EXISTS transfer_listings (
    id BIGSERIAL PRIMARY KEY,
    player_id BIGINT NOT NULL UNIQUE REFERENCES players(id) ON DELETE CASCADE,
    asking_price BIGINT NOT NULL,
    listed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_transfer_listings_player_id ON transfer_listings(player_id);
