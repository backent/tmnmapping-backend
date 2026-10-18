-- Brand contact details.
--
-- A quotation is addressed to a person, not to a brand: the printed document carries
-- "Attention to", their job title, phone and email. Those four were typed afresh into
-- every quotation, which meant retyping the same person for every campaign of the
-- same brand.
--
-- They now live on the brand as the master record, and the quotation wizard prefills
-- from here. The quotation KEEPS ITS OWN COPY: migration 019 stores the contact on
-- the quotation row, and that stays the historical record. Editing a brand's contact
-- must never change what an already-approved quotation prints -- the same reason
-- building names and prices are snapshotted rather than joined.
--
-- Nullable, not NOT NULL: brands already exist without contact details, and a NOT
-- NULL column would refuse to add. The requirement is enforced by the API on create
-- and update, so existing brands keep working until someone edits them.
ALTER TABLE brands ADD COLUMN IF NOT EXISTS attention_to VARCHAR(255);
ALTER TABLE brands ADD COLUMN IF NOT EXISTS job_title VARCHAR(255);
ALTER TABLE brands ADD COLUMN IF NOT EXISTS contact_phone VARCHAR(50);
ALTER TABLE brands ADD COLUMN IF NOT EXISTS contact_email VARCHAR(255);
