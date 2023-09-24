package console

import (
	"github.com/irvankadhafi/go-point-of-sales/internal/config"
	"github.com/irvankadhafi/go-point-of-sales/internal/db"
	"github.com/irvankadhafi/go-point-of-sales/utils"
	migrate "github.com/rubenv/sql-migrate"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"strconv"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "migrate database",
	Long:  `This subcommand used to migrate database`,
	Run:   processMigration,
}

func init() {
	migrateCmd.PersistentFlags().Int("step", 0, "maximum migration steps")
	migrateCmd.PersistentFlags().String("direction", "up", "migration direction")
	RootCmd.AddCommand(migrateCmd)
}

func processMigration(cmd *cobra.Command, args []string) {
	log.Info("Process migration!")

	direction := cmd.Flag("direction").Value.String()
	stepStr := cmd.Flag("step").Value.String()
	step, err := strconv.Atoi(stepStr)
	if err != nil {
		log.WithField("stepStr", stepStr).Fatal("Failed to parse step to int: ", err)
	}

	db.InitializePostgresConn()
	migration(direction, step)
}

func migration(direction string, step int) {
	var (
		n                  int
		migrationDirection migrate.MigrationDirection = migrate.Up
	)

	migrations := &migrate.FileMigrationSource{
		Dir: "db/migration/",
	}

	migrate.SetTable("migrations")

	postgresDB, err := db.PostgreSQL.DB()
	if err != nil {
		log.WithField("DatabaseDSN", config.DatabaseDSN()).Fatal("Failed to connect database: ", err)
	}

	if direction == "down" {
		migrationDirection = migrate.Down
	}

	n, err = migrate.ExecMax(postgresDB, "postgres", migrations, migrationDirection, step)
	if err != nil {
		log.WithFields(log.Fields{
			"migrations": utils.Dump(migrations),
			"direction":  direction,
		}).Fatal("Failed to migrate database: ", err)
	}

	log.Infof("Applied %d migrations!\n", n)
}
