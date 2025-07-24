

# Use official Postgres image
FROM postgres:16-alpine

# Copy schema/data into initialization directory
COPY up.sql /docker-entrypoint-initdb.d/1.sql

# Entrypoint already starts postgres; no need to override CMD
# So this is optional and not needed unless you’re customizing

# CMD ["postgres"]