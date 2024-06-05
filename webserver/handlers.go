package webserver

import (
    "encoding/json"
    "net/http"
)

func RunSpiderHandler(w http.ResponseWriter, r *http.Request) {
    err := RunSpider()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{"message": "Spider iniciado com sucesso!"})
}
