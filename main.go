package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
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
		verbose := fs.Bool("verbose", false, "Include subtasks")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if fs.NArg() != 0 {
			return fmt.Errorf("tasks list does not accept positional arguments")
		}
		items, err := tasks.List(ctx, client, tasks.ListOptions{IncludeSubtasks: *verbose})
		if err != nil {
			return err
		}
		if output.JSON {
			return internal.PrintJSON(os.Stdout, tasks.ToJSONTasks(items))
		}
		printTasks(items, tasks.DefaultListTitle, output)
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
	case "show":
		client, err := mgr.AuthorizedHTTPClient(ctx)
		if err != nil {
			return err
		}
		fs := flag.NewFlagSet("tasks show", flag.ContinueOnError)
		fs.SetOutput(os.Stderr)
		listName := fs.String("list", "", "Task list name")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if fs.NArg() != 1 {
			return errors.New(`tasks show requires exactly one task id or title, e.g. stan tasks show "Agent47"`)
		}
		detail, err := tasks.Show(ctx, client, tasks.ShowOptions{
			Query:    fs.Arg(0),
			ListName: *listName,
		})
		if err != nil {
			return err
		}
		if output.JSON {
			return internal.PrintJSON(os.Stdout, tasks.ToJSONTaskDetail(detail))
		}
		if output.Quiet {
			fmt.Fprintln(os.Stdout, internal.CompactLine(detail.Task.ID, detail.Task.Title))
			return nil
		}
		printTaskDetail(detail, output)
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
			return internal.PrintJSON(os.Stdout, tasks.ToJSONTasks([]tasks.Task{*item})[0])
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
		indent := strings.Repeat("  ", item.Depth)
		line := fmt.Sprintf("%s%s %s", indent, checkbox, item.Title)
		fmt.Fprintln(os.Stdout, internal.TaskLineColor(!output.NoColor, item.Completed, item.Due, now, line))
	}
	if len(items) == 0 {
		fmt.Fprintln(os.Stdout, "No tasks found.")
	}
}

func printTaskDetail(detail *tasks.TaskDetail, output internal.OutputOptions) {
	now := time.Now()
	printTaskDetailNode(*detail, 0, output, now)
}

func printTaskDetailNode(detail tasks.TaskDetail, depth int, output internal.OutputOptions, now time.Time) {
	item := detail.Task
	checkbox := "[ ]"
	if item.Completed {
		checkbox = "[x]"
	}
	indent := strings.Repeat("  ", depth)
	line := fmt.Sprintf("%s%s %s", indent, checkbox, item.Title)
	fmt.Fprintln(os.Stdout, internal.TaskLineColor(!output.NoColor, item.Completed, item.Due, now, line))
	if item.Notes != "" {
		printTaskField(indent, "notes", item.Notes)
	}
	for _, subtask := range detail.Subtasks {
		fmt.Fprintln(os.Stdout)
		printTaskDetailNode(subtask, depth+1, output, now)
	}
}

func printTaskField(indent string, name string, value string) {
	prefix := indent + "  " + name + ": "
	lines := strings.Split(value, "\n")
	for i, line := range lines {
		if i == 0 {
			fmt.Fprintln(os.Stdout, prefix+line)
			continue
		}
		fmt.Fprintln(os.Stdout, strings.Repeat(" ", len(prefix))+line)
	}
}

func printRootHelp() {
	fmt.Println("Stan CLI")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  stan auth <login|status|logout>")
	fmt.Println("  stan calendar <list|add>")
	fmt.Println("  stan tasks <list|lists|show|add>")
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
	fmt.Println(`  stan calendar add --when "10:00" [--duration 30m] [--end 2026-03-20T10:30] "Meeting"`)
}

func printTasksHelp() {
	fmt.Println("Usage:")
	fmt.Println("  stan tasks list [--verbose]")
	fmt.Println("  stan tasks lists")
	fmt.Println(`  stan tasks show [--list Stan] "Agent47"`)
	fmt.Println(`  stan tasks add [--list Personal] [--due 2026-03-25] [--notes "..."] "Buy milk"`)
}

func init() { flag.CommandLine.SetOutput(os.Stderr) }
