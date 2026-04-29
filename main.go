package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"time"

	"stan/auth"
	"stan/calendar"
	"stan/internal"
	"stan/tasks"
)

func main() {
	if err := run(context.Background(), os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string) error {
	cfg, err := internal.LoadConfig()
	if err != nil {
		return err
	}

	filteredArgs, output, err := extractGlobalFlags(args)
	if err != nil {
		return err
	}
	if len(filteredArgs) == 0 {
		printRootHelp()
		return nil
	}

	authManager := auth.NewManager(cfg)

	switch filteredArgs[0] {
	case "auth":
		return runAuth(ctx, authManager, filteredArgs[1:], output)
	case "calendar":
		return runCalendar(ctx, authManager, filteredArgs[1:], output)
	case "tasks":
		return runTasks(ctx, authManager, filteredArgs[1:], output)
	case "help", "-h", "--help":
		printRootHelp()
		return nil
	default:
		return fmt.Errorf("unknown command %q", filteredArgs[0])
	}
}

func runAuth(ctx context.Context, mgr *auth.Manager, args []string, output internal.OutputOptions) error {
	if len(args) == 0 {
		printAuthHelp()
		return nil
	}

	switch args[0] {
	case "login":
		if len(args) > 1 {
			return fmt.Errorf("auth login does not accept extra arguments")
		}
		return mgr.Login(ctx)
	case "status":
		if len(args) > 1 {
			return fmt.Errorf("auth status does not accept extra arguments")
		}
		status, err := mgr.Status(ctx)
		if err != nil {
			return err
		}
		if output.JSON {
			return internal.PrintJSON(os.Stdout, status)
		}
		if output.Quiet {
			fmt.Fprintln(os.Stdout, status.Email)
			return nil
		}
		fmt.Fprintf(os.Stdout, "Logged in as: %s\n", status.Email)
		fmt.Fprintf(os.Stdout, "Token expires in: %s\n", internal.RelativeExpiry(status.Expiry, time.Now()))
		fmt.Fprintf(os.Stdout, "Storage: %s\n", status.Storage)
		return nil
	case "logout":
		if len(args) > 1 {
			return fmt.Errorf("auth logout does not accept extra arguments")
		}
		return mgr.Logout(ctx)
	case "help", "-h", "--help":
		printAuthHelp()
		return nil
	default:
		return fmt.Errorf("unknown auth command %q", args[0])
	}
}

func runCalendar(ctx context.Context, mgr *auth.Manager, args []string, output internal.OutputOptions) error {
	if len(args) == 0 {
		printCalendarHelp()
		return nil
	}

	switch args[0] {
	case "list":
		client, err := mgr.AuthorizedHTTPClient(ctx)
		if err != nil {
			return err
		}
		fs := flag.NewFlagSet("calendar list", flag.ContinueOnError)
		fs.SetOutput(os.Stderr)
		days := fs.Int("days", 7, "Number of days ahead")
		startStr := fs.String("start", "", "Start date")
		endStr := fs.String("end", "", "End date")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if fs.NArg() != 0 {
			return fmt.Errorf("calendar list does not accept positional arguments")
		}
		opts, err := calendar.ParseListOptions(*days, *startStr, *endStr, time.Now())
		if err != nil {
			return err
		}
		events, err := calendar.List(ctx, client, opts)
		if err != nil {
			return err
		}
		if output.JSON {
			return internal.PrintJSON(os.Stdout, events)
		}
		printCalendarEvents(events, output)
		return nil
	case "add":
		client, err := mgr.AuthorizedHTTPClient(ctx)
		if err != nil {
			return err
		}
		fs := flag.NewFlagSet("calendar add", flag.ContinueOnError)
		fs.SetOutput(os.Stderr)
		when := fs.String("when", "", "Start date/time")
		duration := fs.Duration("duration", time.Hour, "Duration")
		endStr := fs.String("end", "", "End date/time")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if *when == "" {
			return errors.New("calendar add requires --when")
		}
		if fs.NArg() != 1 {
			return errors.New(`calendar add requires exactly one title, e.g. stan calendar add "Meeting" --when "10:00"`)
		}
		event, err := calendar.Add(ctx, client, fs.Arg(0), *when, *duration, *endStr, time.Now())
		if err != nil {
			return err
		}
		if output.JSON {
			return internal.PrintJSON(os.Stdout, event)
		}
		if output.Quiet {
			fmt.Fprintln(os.Stdout, internal.CompactLine(event.ID, event.Title))
			return nil
		}
		fmt.Fprintf(os.Stdout, "Created event: %s\n", event.Title)
		fmt.Fprintf(os.Stdout, "%s - %s\n", event.Start.Format(time.RFC3339), event.End.Format(time.RFC3339))
		return nil
	case "help", "-h", "--help":
		printCalendarHelp()
		return nil
	default:
		return fmt.Errorf("unknown calendar command %q", args[0])
	}
}

func runTasks(ctx context.Context, mgr *auth.Manager, args []string, output internal.OutputOptions) error {
	if len(args) == 0 {
		printTasksHelp()
		return nil
	}

	switch args[0] {
	case "list":
		client, err := mgr.AuthorizedHTTPClient(ctx)
		if err != nil {
			return err
		}
		fs := flag.NewFlagSet("tasks list", flag.ContinueOnError)
		fs.SetOutput(os.Stderr)
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if fs.NArg() != 0 {
			return fmt.Errorf("tasks list does not accept positional arguments")
		}
		items, err := tasks.List(ctx, client)
		if err != nil {
			return err
		}
		if output.JSON {
			return internal.PrintJSON(os.Stdout, items)
		}
		printTasks(items, "@default", output)
		return nil
	case "lists":
		client, err := mgr.AuthorizedHTTPClient(ctx)
		if err != nil {
			return err
		}
		fs := flag.NewFlagSet("tasks lists", flag.ContinueOnError)
		fs.SetOutput(os.Stderr)
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		lists, err := tasks.ListTaskLists(ctx, client)
		if err != nil {
			return err
		}
		if output.JSON {
			return internal.PrintJSON(os.Stdout, lists)
		}
		for _, list := range lists {
			if output.Quiet {
				fmt.Fprintln(os.Stdout, list.ID)
				continue
			}
			fmt.Fprintf(os.Stdout, "%s\t%s\n", list.ID, list.Title)
		}
		return nil
	case "add":
		client, err := mgr.AuthorizedHTTPClient(ctx)
		if err != nil {
			return err
		}
		fs := flag.NewFlagSet("tasks add", flag.ContinueOnError)
		fs.SetOutput(os.Stderr)
		listName := fs.String("list", "", "Task list name")
		due := fs.String("due", "", "Due date")
		notes := fs.String("notes", "", "Notes")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if fs.NArg() != 1 {
			return errors.New(`tasks add requires exactly one title, e.g. stan tasks add "Buy milk"`)
		}
		item, err := tasks.Add(ctx, client, tasks.AddOptions{
			Title:    fs.Arg(0),
			ListName: *listName,
			Due:      *due,
			Notes:    *notes,
			Now:      time.Now(),
		})
		if err != nil {
			return err
		}
		if output.JSON {
			return internal.PrintJSON(os.Stdout, item)
		}
		if output.Quiet {
			fmt.Fprintln(os.Stdout, internal.CompactLine(item.ID, item.Title))
			return nil
		}
		fmt.Fprintf(os.Stdout, "Created task: %s\n", item.Title)
		return nil
	case "help", "-h", "--help":
		printTasksHelp()
		return nil
	default:
		return fmt.Errorf("unknown tasks command %q", args[0])
	}
}

func extractGlobalFlags(args []string) ([]string, internal.OutputOptions, error) {
	var output internal.OutputOptions
	filtered := make([]string, 0, len(args))

	for _, arg := range args {
		switch arg {
		case "--json":
			output.JSON = true
		case "-q":
			output.Quiet = true
		case "--no-color":
			output.NoColor = true
		default:
			filtered = append(filtered, arg)
		}
	}
	if output.JSON && output.Quiet {
		return nil, internal.OutputOptions{}, errors.New("--json and -q cannot be used together")
	}
	return filtered, output, nil
}

func printCalendarEvents(events []calendar.Event, output internal.OutputOptions) {
	if output.Quiet {
		for _, event := range events {
			fmt.Fprintln(os.Stdout, internal.CompactLine(event.ID, event.Title))
		}
		return
	}

	now := time.Now()
	lastHeader := ""
	for _, event := range events {
		header := internal.HumanDayLabel(event.Start, now)
		if header != lastHeader {
			if lastHeader != "" {
				fmt.Fprintln(os.Stdout)
			}
			fmt.Fprintf(os.Stdout, "📅 %s\n", header)
			lastHeader = header
		}
		line := fmt.Sprintf("%s  %s", event.Start.Format("15:04"), event.Title)
		fmt.Fprintln(os.Stdout, internal.CalendarLineColor(!output.NoColor, event.Start, now, line))
	}
	if len(events) == 0 {
		fmt.Fprintln(os.Stdout, "No events found.")
	}
}

func printTasks(items []tasks.Task, listName string, output internal.OutputOptions) {
	now := time.Now()
	if output.Quiet {
		for _, item := range items {
			fmt.Fprintln(os.Stdout, internal.CompactLine(item.ID, item.Title))
		}
		return
	}

	fmt.Fprintf(os.Stdout, "Tasks (%s)\n\n", listName)
	for _, item := range items {
		checkbox := "[ ]"
		if item.Completed {
			checkbox = "[x]"
		}
		line := fmt.Sprintf("%s %s", checkbox, item.Title)
		fmt.Fprintln(os.Stdout, internal.TaskLineColor(!output.NoColor, item.Completed, item.Due, now, line))
	}
	if len(items) == 0 {
		fmt.Fprintln(os.Stdout, "No tasks found.")
	}
}

func printRootHelp() {
	fmt.Println("Stan CLI")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  stan auth <login|status|logout>")
	fmt.Println("  stan calendar <list|add>")
	fmt.Println("  stan tasks <list|lists|add>")
	fmt.Println("")
	fmt.Println("Global flags:")
	fmt.Println("  --json")
	fmt.Println("  -q")
	fmt.Println("  --no-color")
}

func printAuthHelp() {
	fmt.Println("Usage: stan auth <login|status|logout>")
}

func printCalendarHelp() {
	fmt.Println("Usage:")
	fmt.Println("  stan calendar list [--days N] [--start YYYY-MM-DD] [--end YYYY-MM-DD]")
	fmt.Println(`  stan calendar add "Meeting" --when "10:00" [--duration 30m] [--end 2026-03-20T10:30]`)
}

func printTasksHelp() {
	fmt.Println("Usage:")
	fmt.Println("  stan tasks list")
	fmt.Println("  stan tasks lists")
	fmt.Println(`  stan tasks add "Buy milk" [--list Personal] [--due 2026-03-25] [--notes "..."]`)
}

func init() { flag.CommandLine.SetOutput(os.Stderr) }
