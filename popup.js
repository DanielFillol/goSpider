document.getElementById('run-script').addEventListener('click', () => {
    chrome.tabs.query({ active: true, currentWindow: true }, (tabs) => {
        chrome.scripting.executeScript({
            target: { tabId: tabs[0].id },
            function: runGoSpider
        });
    });
});

function runGoSpider() {
    // Chama a API do goSpider
    fetch('http://localhost:8080/run', { 
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({ /* Dados necessários para a API */ })
    })
    .then(response => response.json())
    .then(data => {
        console.log(data);
        alert('goSpider iniciado com sucesso!');
    })
    .catch(error => console.error('Error:', error));
}
