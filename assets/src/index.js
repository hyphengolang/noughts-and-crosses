
const app = document.getElementById("app");

app.innerHTML = `
    <h1>Welcome to my Go project</h1>
    <button>click me!</button>
    <div class="game"></div>
`;

document.getElementsByTagName('button')[0].addEventListener("click", async e => {
    const r = await fetch("/api/game/create");
    const text = await r.text();

    document.querySelector(".game").textContent = text;
});

// const ws = new WebSocket('ws://localhost:8080/ws');

// ws.onopen = (e) => { console.log('Connected to server', e); };

// ws.onclose = () => {
//     button.disabled = true;
// };

// ws.onmessage = e => { console.log(e.data); };

// let n = 0;
// button.addEventListener('click', () => {
//     n > 10 ? (n = 0) : (ws.send(++n));
// });