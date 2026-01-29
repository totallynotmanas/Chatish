(function(){
    const loginScreen = document.getElementById('loginScreen');
    const otpScreen = document.getElementById('otpScreen');
    const chatScreen = document.getElementById('chatScreen');
    const usernameInput = document.getElementById('username');
    const passwordInput = document.getElementById('password');
    const otpInput = document.getElementById('otp');
    const messageInput = document.getElementById('text');
    const roomsList = document.getElementById('roomsList');
    const messages = document.getElementById('messages');
    const loginBtn = document.getElementById('loginBtn');
    const otpBtn = document.getElementById('otpBtn');
    const sendBtn = document.getElementById('send');
    const loginStatus = document.getElementById('loginStatus');
    const otpStatus = document.getElementById('otpStatus');
    const welcomeMsg = document.getElementById('welcomeMsg');
    const currentRoomTitle = document.getElementById('currentRoomTitle');

    let ws;
    let username = '';
    let currentRoom = '';
    let rooms = [];

    function showScreen(screen) {
        loginScreen.classList.remove('active');
        otpScreen.classList.remove('active');
        chatScreen.classList.remove('active');
        screen.classList.add('active');
    }

    function connectWebSocket() {
        ws = new WebSocket((location.protocol === 'https:' ? 'wss://' : 'ws://') + location.host + '/ws');
        ws.addEventListener('open', () => { 
            console.log('WebSocket connected');
            showScreen(loginScreen);
        });
        ws.addEventListener('close', () => { 
            console.log('WebSocket disconnected');
            setTimeout(connectWebSocket, 1000); 
        });
        ws.addEventListener('message', (ev) => {
            try {
                const msg = JSON.parse(ev.data);
                handleServerMessage(msg);
            } catch (e) {
                console.log('Raw message:', ev.data);
            }
        });
        ws.addEventListener('error', (err) => {
            console.error('WebSocket error:', err);
            loginStatus.textContent = '❌ Connection error';
            loginStatus.classList.add('error');
        });
    }

    function renderRoomsList() {
        roomsList.innerHTML = '';
        rooms.forEach(room => {
            const el = document.createElement('div');
            el.className = 'room-item' + (room === currentRoom ? ' active' : '');
            el.innerHTML = `
                <div class="room-item-name"># ${room}</div>
                <div class="room-item-status">Tap to join</div>
            `;
            el.addEventListener('click', () => join(room));
            roomsList.appendChild(el);
        });
    }

    function handleServerMessage(msg) {
        console.log('Server message:', msg);
        if (msg.type === 'otp_required') {
            showScreen(otpScreen);
            otpStatus.textContent = 'Enter the OTP from the server console';
        } else if (msg.type === 'auth_success') {
            username = usernameInput.value;
            welcomeMsg.textContent = `👤 ${username}`;
            showScreen(chatScreen);
            ws.send(JSON.stringify({ type: 'list' }));
        } else if (msg.type === 'rooms_list') {
            rooms = msg.rooms;
            renderRoomsList();
            join('General');
        } else if (msg.type === 'error') {
            if (otpScreen.classList.contains('active')) {
                otpStatus.textContent = '❌ ' + (msg.content || msg.error);
                otpStatus.classList.add('error');
            } else {
                loginStatus.textContent = '❌ ' + (msg.content || msg.error);
                loginStatus.classList.add('error');
            }
        } else if (msg.type === 'chat') {
            const el = document.createElement('div');
            el.className = 'msg';
            el.innerHTML = `<span class="msg-user">${msg.username}:</span> <span class="msg-content">${escapeHtml(msg.content)}</span>`;
            messages.appendChild(el);
            messages.scrollTop = messages.scrollHeight;
        } else if (msg.type === 'system') {
            const el = document.createElement('div');
            el.className = 'msg-system';
            el.textContent = '✓ ' + msg.content;
            messages.appendChild(el);
            messages.scrollTop = messages.scrollHeight;
        } else if (msg.type === 'join_success') {
            currentRoom = msg.room;
            currentRoomTitle.textContent = '# ' + msg.room;
            messages.innerHTML = '';
            renderRoomsList();
            const el = document.createElement('div');
el.className = 'msg-system';
            el.textContent = '✓ Joined #' + msg.room;
            messages.appendChild(el);
        }
    }

    function escapeHtml(text) {
        const map = {
            '&': '&amp;',
            '<': '&lt;',
            '>': '&gt;',
            '"': '&quot;',
            "'": '&#039;'
        };
        return text.replace(/[&<>"']/g, m => map[m]);
    }

    function login() {
        const user = usernameInput.value.trim();
        const pass = passwordInput.value.trim();

        if (!user || !pass) {
            loginStatus.textContent = '❌ Please enter username and password';
            loginStatus.classList.add('error');
            return;
        }

        loginStatus.textContent = '⏳ Authenticating...';
        loginStatus.classList.remove('error', 'success');

        ws.send(JSON.stringify({
            type: 'auth',
            username: user,
            password: pass
        }));
    }

    function submitOTP() {
        const otp = otpInput.value.trim();
        if (!otp || otp.length !== 6) {
            otpStatus.textContent = '❌ Please enter a 6-digit OTP';
            otpStatus.classList.add('error');
            return;
        }

        otpStatus.textContent = '⏳ Verifying...';
        otpStatus.classList.remove('error');

        ws.send(JSON.stringify({
            type: 'otp',
            otp: otp
        }));
    }

    function join(room) {
        console.log('Joining room:', room);
        ws.send(JSON.stringify({
            type: 'join',
            room: room
        }));
    }

    function sendMessage() {
        const text = messageInput.value.trim();
        if (!text || !currentRoom) return;

        console.log('Sending message to', currentRoom, ':', text);
        ws.send(JSON.stringify({
            type: 'message',
            content: text,
            room: currentRoom
        }));

        messageInput.value = '';
    }

    loginBtn.addEventListener('click', login);
    otpBtn.addEventListener('click', submitOTP);
    sendBtn.addEventListener('click', sendMessage);
    messageInput.addEventListener('keydown', (e) => { if (e.key === 'Enter') sendMessage(); });
    otpInput.addEventListener('keydown', (e) => { if (e.key === 'Enter') submitOTP(); });
    usernameInput.addEventListener('keydown', (e) => { if (e.key === 'Enter') login(); });
    passwordInput.addEventListener('keydown', (e) => { if (e.key === 'Enter') login(); });

    connectWebSocket();
})();
