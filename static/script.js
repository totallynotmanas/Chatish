(function(){
    const msgs = document.getElementById('messages');
    const input = document.getElementById('text');
    const nameInput = document.getElementById('name');
    const sendBtn = document.getElementById('send');
    const status = document.getElementById('status');

    let ws;
    function connect() {
        ws = new WebSocket((location.protocol === 'https:' ? 'wss://' : 'ws://') + location.host + '/ws');
        ws.addEventListener('open', () => { status.textContent = 'Connected'; });
        ws.addEventListener('close', () => { status.textContent = 'Disconnected — retrying in 1s'; setTimeout(connect, 1000); });
        ws.addEventListener('message', (ev) => {
            const text = ev.data;
            const el = document.createElement('div');
            el.className = 'msg';
            el.textContent = text;
            msgs.appendChild(el);
            msgs.scrollTop = msgs.scrollHeight;
        });
    }

    sendBtn.addEventListener('click', sendMessage);
    input.addEventListener('keydown', (e) => { if (e.key === 'Enter') sendMessage(); });

    function sendMessage(){
        if (!ws || ws.readyState !== WebSocket.OPEN) return;
        const name = nameInput.value.trim() || 'Anonymous';
        const text = input.value.trim();
        if (!text) return;
        const msg = name + ': ' + text;
        ws.send(msg);
        input.value = '';
    }

    connect();
})();
