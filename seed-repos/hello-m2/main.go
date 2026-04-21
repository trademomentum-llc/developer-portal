package main

import (
    "fmt"
    "log"
    "net/http"
    "os"
)

func main() {
    http.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
        fmt.Fprintf(w, "hello from hello-m2 env=%s secret=%s\n",
            os.Getenv("ENVIRONMENT"),
            redact(os.Getenv("EXAMPLE_SECRET")),
        )
    })
    addr := ":8080"
    log.Printf("listening on %s", addr)
    log.Fatal(http.ListenAndServe(addr, nil))
}

func redact(s string) string {
    if len(s) <= 4 {
        return "****"
    }
    return s[:2] + "****" + s[len(s)-2:]
}
