-- Photos this application hosts, as opposed to the ones ERP serves.
--
-- Deliberately a separate table rather than more entries in buildings.images: that
-- column is owned by the ERP photo sync, which rewrites it whenever a path moves, so
-- anything added there would be erased within a cycle. Keeping the two apart also
-- makes "whose photo is this?" answerable, and lets a missing photo mean "nobody
-- uploaded one" rather than "the sync removed it".
--
-- The serving route prefers a row here and falls back to the ERP path, so a building
-- keeps its ERP photo until someone replaces it.
CREATE TABLE IF NOT EXISTS building_images (
    id BIGSERIAL PRIMARY KEY,
    building_id BIGINT NOT NULL,

    -- The same four names ERP uses, so the fallback is one-for-one: asking for the
    -- 'front' photo returns ours if we have one and ERP's if we do not.
    slot VARCHAR(20) NOT NULL,

    -- Relative to BUILDING_IMAGES_DIR. Stored rather than derived so the file can be
    -- renamed or moved without the database disagreeing with the disk.
    path VARCHAR(512) NOT NULL,
    content_type VARCHAR(100) NOT NULL,
    size_bytes BIGINT NOT NULL,

    uploaded_by_user_id BIGINT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- CASCADE: a photo of a building that no longer exists is not a record of
    -- anything, unlike a change row or a quotation snapshot.
    FOREIGN KEY (building_id) REFERENCES buildings(id) ON DELETE CASCADE,
    FOREIGN KEY (uploaded_by_user_id) REFERENCES users(id) ON DELETE SET NULL,

    -- One photo per slot per building. Replacing a photo overwrites the row rather
    -- than accumulating versions nobody asked for.
    CONSTRAINT unique_building_image_slot UNIQUE (building_id, slot),
    CONSTRAINT building_image_slot_check CHECK (
        slot IN ('front', 'back', 'left', 'right_side')),
    CONSTRAINT building_image_size_positive CHECK (size_bytes > 0)
);

CREATE INDEX IF NOT EXISTS idx_building_images_building ON building_images(building_id);
