-- Building prices, one per building, no versions.
--
-- Until now a building's price lived in rate_card_building_prices, keyed to a rate
-- card version that had to be drafted, filled and published before anything could
-- be quoted. Nobody outside the system ever saw the version, and the business does
-- not prepare price lists in advance, so the workflow was ceremony: every price
-- change meant a new draft and a publish.
--
-- The one guarantee versions appeared to provide -- that an approved quotation keeps
-- its prices -- never came from them. A submitted quotation stores its own pricing
-- snapshot and a per-building unit price, and reading it never recalculates. So
-- prices can simply be current, and changed in place.
--
-- A separate table rather than a column on buildings: buildings is owned by the ERP
-- sync, and a commercial price is not ERP data.

CREATE TABLE IF NOT EXISTS building_prices (
    id BIGSERIAL PRIMARY KEY,
    building_id BIGINT NOT NULL,
    price_idr_per_week NUMERIC(18,0) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (building_id) REFERENCES buildings(id) ON DELETE CASCADE,
    CONSTRAINT unique_building_price UNIQUE (building_id),
    CONSTRAINT building_price_non_negative CHECK (price_idr_per_week >= 0)
);

-- Carry across the current rate card so nothing that can be quoted today stops
-- being quotable. The rate card tables are left in place, unused, as history.
INSERT INTO building_prices (building_id, price_idr_per_week)
SELECT rp.building_id, rp.price_idr_per_week
FROM rate_card_building_prices rp
JOIN rate_card_versions v ON v.id = rp.rate_card_version_id
WHERE v.status = 'current'
ON CONFLICT (building_id) DO NOTHING;
