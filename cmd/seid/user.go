package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"golang.org/x/term"

	"github.com/statusengine/interface/internal/auth"
	"github.com/statusengine/interface/internal/config"
	"github.com/statusengine/interface/internal/database"
	"github.com/statusengine/interface/internal/logging"
	"github.com/statusengine/interface/internal/migrate"
)

func userCommand(args []string) error {
	if len(args) == 0 {
		return errors.New(`user: expected one of create, passwd, role, list`)
	}
	sub, rest := args[0], args[1:]

	switch sub {
	case "create":
		return userCreate(rest)
	case "passwd":
		return userPasswd(rest)
	case "role":
		return userRole(rest)
	case "list":
		return userList(rest)
	default:
		return fmt.Errorf("user: unknown subcommand %q; expected create, passwd, role or list", sub)
	}
}

// withStore opens the database, applies migrations and hands a store to fn.
// Every user subcommand needs exactly this, including the very first one
// run against an empty database.
func withStore(args []string, register func(*flag.FlagSet), fn func(context.Context, *auth.Store) error) error {
	fs := flag.NewFlagSet("seid user", flag.ContinueOnError)
	register(fs)
	// The database flags have to be available here too; the same file and
	// environment apply, so only -config and -mysql-dsn are worth repeating.
	var configPath, dsn string
	fs.StringVar(&configPath, "config", os.Getenv("SEI_CONFIG"), "path to a YAML configuration file")
	fs.StringVar(&dsn, "mysql-dsn", "", "MySQL data source name (overrides the config file)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	cfgArgs := []string{}
	if configPath != "" {
		cfgArgs = append(cfgArgs, "-config", configPath)
	}
	if dsn != "" {
		cfgArgs = append(cfgArgs, "-mysql-dsn", dsn)
	}
	cfg, err := config.Load(cfgArgs)
	if err != nil {
		return err
	}

	log, err := logging.New(os.Stderr, "warn", cfg.LogFormat)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	db, err := database.Open(ctx, database.Options{DSN: cfg.MySQLDSN, MaxOpenConns: 2, ConnMaxLife: cfg.MySQLConnMaxLife})
	if err != nil {
		return err
	}
	defer db.Close()

	if err := migrate.Run(ctx, db, log); err != nil {
		return err
	}
	store := auth.NewStore(db)
	if err := auth.Bootstrap(ctx, store, log, cfg.DemoMode, cfg.DemoUser); err != nil {
		return err
	}
	return fn(ctx, store)
}

func userCreate(args []string) error {
	var username, roleName, display, email, password string
	return withStore(args, func(fs *flag.FlagSet) {
		fs.StringVar(&username, "username", "", "login name (required)")
		fs.StringVar(&roleName, "role", auth.RoleOperator, "role: admin, operator or guest")
		fs.StringVar(&display, "display-name", "", "name shown in the UI")
		fs.StringVar(&email, "email", "", "email address")
		fs.StringVar(&password, "password", "", "password; omit to be prompted, which keeps it out of your shell history")
	}, func(ctx context.Context, store *auth.Store) error {
		if username == "" {
			return errors.New("user create: -username is required")
		}
		role, err := store.RoleByName(ctx, roleName)
		if errors.Is(err, auth.ErrNotFound) {
			return fmt.Errorf("user create: no role named %q", roleName)
		}
		if err != nil {
			return err
		}

		if password == "" {
			password, err = promptPassword(true)
			if err != nil {
				return err
			}
		}
		if err := checkPassword(password); err != nil {
			return err
		}

		hash, err := auth.HashPassword(password)
		if err != nil {
			return err
		}
		if display == "" {
			display = username
		}
		if _, err := store.CreateUser(ctx, auth.User{
			Username:     username,
			PasswordHash: hash,
			DisplayName:  display,
			Email:        email,
			RoleID:       role.ID,
			IsActive:     true,
		}); err != nil {
			return err
		}
		fmt.Printf("created user %q with role %q\n", username, role.Name)
		return nil
	})
}

func userPasswd(args []string) error {
	var username, password string
	return withStore(args, func(fs *flag.FlagSet) {
		fs.StringVar(&username, "username", "", "login name (required)")
		fs.StringVar(&password, "password", "", "new password; omit to be prompted")
	}, func(ctx context.Context, store *auth.Store) error {
		if username == "" {
			return errors.New("user passwd: -username is required")
		}
		u, err := store.UserByUsername(ctx, username)
		if errors.Is(err, auth.ErrNotFound) {
			return fmt.Errorf("user passwd: %w: %s", errNoSuchUser, username)
		}
		if err != nil {
			return err
		}

		if password == "" {
			password, err = promptPassword(true)
			if err != nil {
				return err
			}
		}
		if err := checkPassword(password); err != nil {
			return err
		}
		hash, err := auth.HashPassword(password)
		if err != nil {
			return err
		}
		if err := store.SetPassword(ctx, u.ID, hash); err != nil {
			return err
		}
		// A password change should end every other session, or a stolen
		// cookie outlives the password it was obtained with.
		if err := store.DeleteUserSessions(ctx, u.ID); err != nil {
			return err
		}
		fmt.Printf("password updated for %q; existing sessions were ended\n", username)
		return nil
	})
}

func userRole(args []string) error {
	var username, roleName string
	return withStore(args, func(fs *flag.FlagSet) {
		fs.StringVar(&username, "username", "", "login name (required)")
		fs.StringVar(&roleName, "role", "", "new role: admin, operator or guest (required)")
	}, func(ctx context.Context, store *auth.Store) error {
		if username == "" || roleName == "" {
			return errors.New("user role: -username and -role are required")
		}
		u, err := store.UserByUsername(ctx, username)
		if errors.Is(err, auth.ErrNotFound) {
			return fmt.Errorf("user role: %w: %s", errNoSuchUser, username)
		}
		if err != nil {
			return err
		}
		role, err := store.RoleByName(ctx, roleName)
		if errors.Is(err, auth.ErrNotFound) {
			return fmt.Errorf("user role: no role named %q", roleName)
		}
		if err != nil {
			return err
		}
		if err := store.SetRole(ctx, u.ID, role.ID); err != nil {
			return err
		}
		fmt.Printf("%q is now on role %q\n", username, role.Name)
		return nil
	})
}

func userList(args []string) error {
	return withStore(args, func(fs *flag.FlagSet) {}, func(ctx context.Context, store *auth.Store) error {
		users, err := store.ListUsers(ctx)
		if err != nil {
			return err
		}
		if len(users) == 0 {
			fmt.Println("no accounts yet; create one with: seid user create -username <name> -role admin")
			return nil
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "USERNAME\tROLE\tACTIVE\tLAST LOGIN")
		for _, u := range users {
			last := "never"
			if u.LastLoginAt != nil {
				last = time.Unix(*u.LastLoginAt, 0).Format(time.RFC3339)
			}
			fmt.Fprintf(w, "%s\t%s\t%t\t%s\n", u.Username, u.RoleName, u.IsActive, last)
		}
		return w.Flush()
	})
}

// checkPassword enforces a floor, not a policy. Composition rules push
// people toward predictable substitutions; length is what actually costs
// an attacker something.
func checkPassword(p string) error {
	if len([]rune(p)) < 12 {
		return errors.New("password must be at least 12 characters")
	}
	return nil
}

func promptPassword(confirm bool) (string, error) {
	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		// Not a terminal: read one line, so the password can be piped in
		// from a secret store without ending up in the process table.
		line, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil && line == "" {
			return "", fmt.Errorf("reading password from stdin: %w", err)
		}
		return strings.TrimRight(line, "\r\n"), nil
	}

	fmt.Fprint(os.Stderr, "Password: ")
	first, err := term.ReadPassword(fd)
	fmt.Fprintln(os.Stderr)
	if err != nil {
		return "", fmt.Errorf("reading password: %w", err)
	}
	if !confirm {
		return string(first), nil
	}

	fmt.Fprint(os.Stderr, "Repeat: ")
	second, err := term.ReadPassword(fd)
	fmt.Fprintln(os.Stderr)
	if err != nil {
		return "", fmt.Errorf("reading password: %w", err)
	}
	if string(first) != string(second) {
		return "", errors.New("the two passwords do not match")
	}
	return string(first), nil
}
