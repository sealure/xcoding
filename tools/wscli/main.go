package main

import (
    "flag"
    "fmt"
    "log"
    "net/url"
    "os"
    "os/signal"
    "syscall"

    "github.com/gorilla/websocket"
)

func main() {
    defURL := "ws://localhost:5175/ci_service/api/v1/executor/ws/builds/40"
    raw := getenv("WS_URL", defURL)
    off := getenv("WS_OFFSET", "0")

    u, err := url.Parse(raw)
    if err != nil {
        log.Fatalf("invalid url: %v", err)
    }
    q := u.Query()
    if q.Get("offset") == "" {
        q.Set("offset", off)
    }
    u.RawQuery = q.Encode()

    log.Printf("connecting: %s", u.String())
    conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
    if err != nil {
        log.Fatalf("dial error: %v", err)
    }
    defer conn.Close()

    done := make(chan struct{})
    go func() {
        defer close(done)
        for {
            var msg map[string]any
            if err := conn.ReadJSON(&msg); err != nil {
                log.Printf("read error: %v", err)
                return
            }
            fmt.Printf("%v\n", msg)
        }
    }()

    // wait for interrupt to exit
    sigc := make(chan os.Signal, 1)
    signal.Notify(sigc, os.Interrupt, syscall.SIGTERM)
    <-sigc
    log.Printf("closing")
}

func getenv(k, def string) string {
    if v := os.Getenv(k); v != "" {
        return v
    }
    // allow -ws-url and -ws-offset overrides
    if k == "WS_URL" {
        urlFlag := flag.String("ws-url", def, "websocket url")
        flag.Parse()
        return *urlFlag
    }
    if k == "WS_OFFSET" {
        offFlag := flag.String("ws-offset", def, "offset id")
        flag.Parse()
        return *offFlag
    }
    return def
}

