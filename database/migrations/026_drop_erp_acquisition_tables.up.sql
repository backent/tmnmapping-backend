-- Retire the acquisition and building-proposal mirrors.
--
-- Both were copies of ERP DocTypes, refreshed on a timer, and nothing writes or reads
-- them any more: buildings are maintained by spreadsheet, so building_status and the
-- screen count come from the file rather than from these feeds, and the two dashboard
-- tabs that read them were removed with them.
--
-- Dropping rather than keeping them empty: a table nobody writes is a table someone
-- eventually reads and believes. Nothing is lost that ERP does not still hold, and
-- nothing references them -- neither table had a foreign key in either direction.
--
-- letters_of_intent is deliberately untouched. That feed keeps syncing, and
-- /dashboard/loi keeps its report.
DROP TABLE IF EXISTS acquisitions;
DROP TABLE IF EXISTS building_proposals;
