package main

import (
    "log"
    "net/http"
    "goSpider/webserver"
)

func main() {
    http.HandleFunc("/run", webserver.RunSpiderHandler)

    log.Println("Starting server on :8080")
    if err := http.ListenAndServe(":8080", nil); err != nil {
        log.Fatalf("could not start server: %v\n", err)
    }
}
