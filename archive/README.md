# Immutable migration inputs

Original SQLite database and legacy historical exports, preserved byte-for-byte.
The application never serves this directory. Import with `cmd/import-history`.
Do not update these files with the ingestion job. Keep them until production
migration and backups have been verified. Checksums are listed in SHA256SUMS.
