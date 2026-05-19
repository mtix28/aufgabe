# Adding an htmx Frontend (Option A)

One binary, web served as a cobra subcommand. Build it step by step and try each piece before moving on.

## 1. Refactor `db/connect.go` so the connection is reusable

Right now `UpdateQuery` opens a new DB connection every call. Split it:

- A function `Connect() (*sqlx.DB, error)` that opens the connection.
- A function `UpdateQuery(db *sqlx.DB, query string, amount int) error` that takes an existing `*sqlx.DB`.
- A new function `GetData(db *sqlx.DB) (incValue, decValue int, err error)` so the web page can show current values.

Also: move the connection string into env vars (`DB_HOST`, `DB_PASSWORD`, …) or a config — don't ship a hardcoded password.

Full file:

```go
package connection

import (
    "fmt"
    "log"
    "os"

    "github.com/jmoiron/sqlx"
    _ "github.com/lib/pq"
)

func Connect() (*sqlx.DB, error) {
    dsn := fmt.Sprintf(
        "host=%s user=%s dbname=%s password=%s sslmode=disable",
        env("DB_HOST", "localhost"),
        env("DB_USER", "postgres"),
        env("DB_NAME", "demo"),
        env("DB_PASSWORD", "secret"),
    )

    db, err := sqlx.Connect("postgres", dsn)
    if err != nil {
        return nil, fmt.Errorf("connect: %w", err)
    }
    if err := db.Ping(); err != nil {
        db.Close()
        return nil, fmt.Errorf("ping: %w", err)
    }
    log.Println("[+] Successfully connected to database")
    return db, nil
}

func UpdateQuery(db *sqlx.DB, query string, amount int) error {
    log.Println("[*] running query..")
    if _, err := db.Exec(query, amount); err != nil {
        return fmt.Errorf("exec: %w", err)
    }
    log.Println("[+] Data successfully updated")
    return nil
}

func GetData(db *sqlx.DB) (incValue, decValue int, err error) {
    row := db.QueryRow(`SELECT "IncValue", "DecValue" FROM "Data" LIMIT 1`)
    if err = row.Scan(&incValue, &decValue); err != nil {
        return 0, 0, fmt.Errorf("scan: %w", err)
    }
    return incValue, decValue, nil
}

func env(key, fallback string) string {
    if v := os.Getenv(key); v != "" {
        return v
    }
    return fallback
}
```

Then update `cmd/increment.go` and `cmd/decrement.go` to call `Connect()` themselves and pass the `*sqlx.DB` in. For example:

```go
Run: func(cmd *cobra.Command, args []string) {
    i, err := strconv.Atoi(args[0])
    if err != nil {
        log.Fatalln(err)
    }

    db, err := connection.Connect()
    if err != nil {
        log.Fatalln(err)
    }
    defer db.Close()

    if err := connection.UpdateQuery(db, incQuery, i); err != nil {
        log.Fatalln(err)
    }
},
```

## 2. Update `cmd/increment.go` and `cmd/decrement.go`

They'll now call `db.Connect()` themselves, then pass the `*sqlx.DB` into `UpdateQuery`. Defer `db.Close()`.

## 3. Add `cmd/serve.go`

A new cobra command. Skeleton:

```go
var serveCmd = &cobra.Command{
    Use:   "serve",
    Short: "Start the htmx web frontend",
    Run: func(cmd *cobra.Command, args []string) {
        // 1. open db once
        // 2. register http handlers (closures that capture db)
        // 3. http.ListenAndServe(":8080", mux)
    },
}
```

Register it in `cmd/root.go` next to the other two: `rootCmd.AddCommand(serveCmd)`.

## 4. Create `web/templates/index.html`

Two buttons that POST to your endpoints with htmx, and a `<div id="values">` that gets swapped:

```html
<!doctype html>
<html>
<head><script src="https://unpkg.com/htmx.org@2"></script></head>
<body>
  <div id="values">{{template "values" .}}</div>
  <button hx-post="/increment" hx-target="#values">+1</button>
  <button hx-post="/decrement" hx-target="#values">-1</button>
</body>
</html>

{{define "values"}}
  <p>Inc: {{.Inc}} — Dec: {{.Dec}}</p>
{{end}}
```

Parse it once at startup with `html/template`.

## 5. Wire up handlers in `serve.go`

Three of them:
a
- `GET /` — render the whole page (calls `GetData`, executes the full template).
- `POST /increment` — call `UpdateQuery(db, incQuery, 1)`, then render **only** the `"values"` block. htmx swaps that fragment into `#values`. No full page reload.
- `POST /decrement` — same pattern.

Use `http.NewServeMux()` and `mux.HandleFunc("/increment", ...)`.

## 6. Try it

```
go run . increment 5      # CLI still works
go run . serve            # open http://localhost:8080
```

## 7. Full `cmd/serve.go` for reference

If you get stuck, here's the whole file written out. It assumes step 1 is done: `db.Connect()`, `db.UpdateQuery(db, query, amount)`, and `db.GetData(db)` exist.

```go
package cmd

import (
    "html/template"
    "log"
    "net/http"

    connection "example.de/demo/db"
    "github.com/jmoiron/sqlx"
    "github.com/spf13/cobra"
)

const (
    incQuery = `UPDATE "Data" SET "IncValue" = "IncValue" + $1`
    decQuery = `UPDATE "Data" SET "DecValue" = "DecValue" + $1`
)

var serveCmd = &cobra.Command{
    Use:   "serve",
    Short: "Start the htmx web frontend",
    Run: func(cmd *cobra.Command, args []string) {
        db, err := connection.Connect()
        if err != nil {
            log.Fatalln(err)
        }
        defer db.Close()

        tmpl := template.Must(template.ParseFiles("web/templates/index.html"))

        mux := http.NewServeMux()
        mux.HandleFunc("GET /", handleIndex(db, tmpl))
        mux.HandleFunc("POST /increment", handleUpdate(db, tmpl, incQuery))
        mux.HandleFunc("POST /decrement", handleUpdate(db, tmpl, decQuery))

        log.Println("[+] listening on :8080")
        log.Fatal(http.ListenAndServe(":8080", mux))
    },
}

type viewData struct {
    Inc int
    Dec int
}

func handleIndex(db *sqlx.DB, tmpl *template.Template) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        inc, dec, err := connection.GetData(db)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        if err := tmpl.Execute(w, viewData{Inc: inc, Dec: dec}); err != nil {
            log.Println(err)
        }
    }
}

func handleUpdate(db *sqlx.DB, tmpl *template.Template, query string) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if err := connection.UpdateQuery(db, query, 1); err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        inc, dec, err := connection.GetData(db)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        // render ONLY the "values" block — htmx swaps this into #values
        if err := tmpl.ExecuteTemplate(w, "values", viewData{Inc: inc, Dec: dec}); err != nil {
            log.Println(err)
        }
    }
}
```

Don't forget to register it in `cmd/root.go`:

```go
rootCmd.AddCommand(serveCmd)
```

## Gotchas to watch for

- **Template parsing**: parse once at startup, not per request. `template.Must(template.ParseFiles(...))`.
- **Embedding templates**: once it works, use `//go:embed web/templates/*` so the binary is self-contained.
- **Concurrent DB access**: `*sqlx.DB` is a pool and is safe for concurrent use — don't put a mutex around it.
- **htmx response = HTML, not JSON**: write the rendered fragment directly to `w`.
