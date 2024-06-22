package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

type Request struct {
	URL string `json:"url"`
}

func executeGoCode(w http.ResponseWriter, r *http.Request) {
	var req Request
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	goCode := fmt.Sprintf(`
package main

import (
  "fmt"
  "net/http"
  "io/ioutil"
)

func main() {
  resp, err := http.Get("%s")
  if err != nil {
    fmt.Println("Error:", err)
    return
  }
  defer resp.Body.Close()
  body, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    fmt.Println("Error:", err)
    return
  }
  fmt.Println(string(body))
}
  `, req.URL)

	err = ioutil.WriteFile("main.go", []byte(goCode), 0644)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	cmd := exec.Command("go", "run", "main.go")
	output, err := cmd.CombinedOutput()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(output)
}

func main() {
	http.HandleFunc("/execute", executeGoCode)
	fmt.Println("Server started at :8080")
	http.ListenAndServe(":8080", nil)
}
