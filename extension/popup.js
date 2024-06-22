document.getElementById('url-form').addEventListener('submit', function(event) {
  event.preventDefault();
  const url = document.getElementById('url-input').value;
  const output = document.getElementById('output');
  const goCode = `
package main

import (
  "fmt"
  "net/http"
  "io/ioutil"
)

func main() {
  resp, err := http.Get("${url}")
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
  `;
  output.textContent = goCode;
});

document.getElementById('execute-button').addEventListener('click', function() {
  const url = document.getElementById('url-input').value;
  const output = document.getElementById('output');

  
  fetch('http://localhost:8080/execute', { 
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({ url: url })
  })
  .then(response => response.text())
  .then(result => {
    output.textContent = result;
  })
  .catch(error => {
    console.error('Error:', error);
    output.textContent = 'Error executing Go code';
  });
});
