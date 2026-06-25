[working-directory: './internal/dba']
db-gen-dba:
  bobgen-psql

watch:
  watchexec -r -e go -- go run ./cmd/server/main.go
