package postgres_test

import (
	"context"
	"net/url"
	"os"
	"testing"

	"github.com/jackc/pgx/v4"
	"github.com/tclemos/goit"
	"github.com/tclemos/goit/postgres"
)

const (
	Port     = 5432
	User     = "postgres_user"
	Password = "postgres_password"
	Database = "postgres_database"
)

var dbUrl url.URL

func TestMain(m *testing.M) {

	ctx := context.Background()

	// Prepare container
	c := postgres.NewContainer(postgres.Params{
		Port:     Port,
		User:     User,
		Password: Password,
		Database: Database,
	})

	// Start container
	goit.Start(ctx, c)
	dbUrl = c.Url()

	// Run tests
	code := m.Run()

	// Stop containers
	goit.Stop()

	// finalize test execution
	os.Exit(code)
}

func TestPostgres(t *testing.T) {

	ctx := context.Background()

	// connect to postgres with the url provided by the container
	conn, err := pgx.Connect(ctx, dbUrl.String())
	if err != nil {
		t.Errorf("Unable to connect to database: %v", err)
		return
	}
	defer conn.Close(ctx)

	var sqlCmd string
	// database initialization
	sqlCmd = `CREATE TABLE things (
		id int primary key,
		name varchar(50)
	);`
	if _, err := conn.Exec(ctx, sqlCmd); err != nil {
		t.Errorf("Unable to create things table: %v", err)
		return
	}

	// insert 1
	sqlCmd = "INSERT INTO things VALUES($1, $2)"
	if _, err := conn.Exec(ctx, sqlCmd, 1, "something"); err != nil {
		t.Errorf("Unable to insert 1: %v", err)
		return
	}

	// intert 2
	sqlCmd = "INSERT INTO things VALUES($1, $2)"
	if _, err := conn.Exec(ctx, sqlCmd, 2, "another thing"); err != nil {
		t.Errorf("Unable to insert 2: %v", err)
		return
	}

	// update
	sqlCmd = "UPDATE things SET name = $2 WHERE id=$1;"
	if _, err := conn.Exec(ctx, sqlCmd, 1, "anything"); err != nil {
		t.Errorf("Unable to update: %v", err)
		return
	}

	// select
	var id int
	var name string
	sqlCmd = "SELECT id, name FROM things WHERE id=$1;"
	if err := conn.QueryRow(ctx, sqlCmd, 1).Scan(&id, &name); err != nil {
		t.Errorf("Unable to select: %v", err)
		return
	}

	if id != 1 {
		t.Errorf("Invalid id, expected 1, found: %d", id)
		return
	}

	if name != "anything" {
		t.Errorf("Invalid name, expected anything, found: %s", name)
		return
	}

	// delete
	sqlCmd = "DELETE FROM things WHERE id=$1;"
	if _, err := conn.Exec(ctx, sqlCmd, 1); err != nil {
		t.Errorf("Unable to update: %v", err)
		return
	}

	// count
	var count int
	sqlCmd = "SELECT count(*) FROM things;"
	if err := conn.QueryRow(ctx, sqlCmd).Scan(&count); err != nil {
		t.Errorf("Unable to select count: %v", err)
		return
	}

	if count != 1 {
		t.Errorf("Invalid count, expected 1, found: %d", count)
		return
	}
}
