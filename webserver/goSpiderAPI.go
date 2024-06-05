package webserver

import (
    "fmt"
    "goSpider"
    "log"
)

func RunSpider() error {
    nav := goSpider.NewNavigator()
    defer nav.Close()

    url := "https://www.example.com"
    err := nav.OpenNewTab(url)
    if err != nil {
        log.Printf("Failed to run spider: %v\n", err)
        return fmt.Errorf("failed to run spider: %v", err)
    }

    log.Println("Spider iniciado com sucesso!")
    return nil
}
