#!/bin/sh

MIGRATED_FILE="/app/migrated"

# check migration status
if [ ! -f "$MIGRATED_FILE" ] || [ "$(cat $MIGRATED_FILE)" != "1" ]; then
  # run migration
  echo "Running migrations..."
  /app/bin/migrate -d up
  echo "Migrations completed."

  # update migration status
  echo "1" > "$MIGRATED_FILE"
fi

# run app
exec /app/bin/app
