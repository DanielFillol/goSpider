package goSpider

import (
	"net/http"
)

// StartTestServer starts a simple HTTP server for testing purposes
func StartTestServer() *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`
			<!DOCTYPE html>
			<html>
			<head><title>Test Page</title></head>
			<body>
				<button id="exampleButton" onclick="showDynamicContent()">Click Me</button>
				<form id="searchForm">
					<input type="text" id="searchBar" />
				</form>
				<form id="loginForm" onsubmit="submitForm(); return false;">
					<input type="text" name="username" />
					<input type="password" name="password" />
					<button type="submit">Login</button>
				</form>
				<select id="dropdown">
					<option value="option1">Option 1</option>
					<option value="option2">Option 2</option>
				</select>
				<input type="checkbox" id="checkbox" />
				<input type="radio" id="radioButton" name="radio" />
				<input type="file" id="fileInput" />
				<div id="dynamicElement" style="display:none;">Dynamic Content</div>
				<div id="nestedElement" style="display:none;">
					<input type="text" id="nestedInput" />
				</div>
				<div id="delayedElement" style="display:none;">Delayed Content</div>
				<script>
					function showDynamicContent() {
						setTimeout(() => {
							document.getElementById('dynamicElement').style.display = 'block';
						}, 1000);
					}

					function submitForm() {
						alert('Form submitted');
					}

					function mockAjax() {
						setTimeout(() => {
							document.getElementById('delayedElement').style.display = 'block';
							console.log('AJAX call completed');
						}, 3000);
					}

					document.getElementById('exampleButton').addEventListener('click', function() {
						setTimeout(() => {
							document.getElementById('nestedElement').style.display = 'block';
						}, 2000);
					});

					mockAjax();
				</script>
			</body>
			</html>
		`))
	})
	server := &http.Server{Addr: ":8080", Handler: mux}
	go server.ListenAndServe()
	return server
}
