const app = document.getElementById("app");

app.innerHTML = `
    <h1>Let's play</h1>
    <button>Go!</button>
`;

const id = window.location.pathname.split("/").at(-1);
const endpoint = `ws://localhost:8080/api/game/play/${id}`;

const ws = new WebSocket(endpoint);

const button = document.querySelector("button");

ws.onopen = (e) => { console.log('Connected to server', e); };

ws.onclose = () => {
    button.disabled = true;
};

ws.onmessage = e => { console.log(e.data); };

let n = 0;
button.addEventListener('click', () => {
    n > 10 ? (n = 0) : (ws.send(++n));
});