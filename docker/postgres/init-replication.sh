#!/bin/bash
# Runs once, only on first initialization of the primary's data volume.
# Enables streaming replication for the db-replica service.
set -e

# Allow replication connections from any host on the Docker network.
echo "host replication all all scram-sha-256" >> "$PGDATA/pg_hba.conf"

# Create the physical replication slot the replica connects with.
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-SQL
	SELECT pg_create_physical_replication_slot('replication_slot');
SQL
