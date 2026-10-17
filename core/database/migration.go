package database

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

func RunMigrations(ctx context.Context, db *pgxpool.Pool) error {
	migrationDir := "migration"

	files, err := os.ReadDir(migrationDir)
	if err != nil {
		return fmt.Errorf("failed to read migration directory: %w", err)
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].Name() < files[j].Name()
	})

	for _, file := range files {

		if file.IsDir() || !strings.HasSuffix(file.Name(), ".sql") {
			continue
		}
		path := filepath.Join(migrationDir, file.Name())

		sql, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %w", file.Name(), err)
		}

		if _, err := db.Exec(ctx, string(sql)); err != nil {
			return fmt.Errorf("failed to run migration %s: %w", file.Name(), err)
		}

		fmt.Println("Migration applied:", file.Name())

	}

	return nil

}

//we have made this file to run the migration files in the migration folder.
// we have used the pgxpool package to connect to the database and run the migrations.
// we have used the os package to read the migration files from the migration folder.
// we have used the sort package to sort the migration files in ascending order.
//  we have used the strings package to check if the file is a sql file or not.
