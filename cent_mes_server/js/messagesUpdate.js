let lastMesID = -1;
async function loadInitMes() {
    const box = document.getElementById('messages');
    box.innerHTML = '';
    const chatID = document.getElementById('chatIDField').value;
    if (chatID == 0) {
        return;
    }
    lastMesID = -1;
    const res = await fetch('/api/messages?chat_id=' + encodeURIComponent(chatID));
    // const res = await fetch('/api/messages', {
    //     method: 'GET',
    //     headers: {
    //         'Content-Type': 'application/json',
    //     },
    //     body: JSON.stringify({
    //         chatID: Number(chatID),
    //         afterID: 0,
    //     }),
    // });
    const msgs = await res.json();
    msgs.forEach(m => {
        appendMessage(m);
    });
    if (msgs.length > 0) {
        lastMesID = msgs[msgs.length - 1].id;
    }
}

function appendMessage(m) {
    const box = document.getElementById('messages');
    const p = document.createElement('p');
    p.innerHTML = `<strong>${m.sender_id}:</strong> ${m.body}`;
    box.appendChild(p);
    box.scrollTop = box.scrollHeight;
}

async function pollNewMessages() {
    const chatId = document.getElementById('chatIDField').value;
    if (chatId == 0) return;
    try {
        const res = await fetch('/api/messages?chat_id=' + encodeURIComponent(chatId) +
                                '&after_id=' + encodeURIComponent(lastMesID));
        // const res = await fetch('/api/messages', {
        //     method: 'GET',
        //     headers: {
        //         'Content-Type': 'application/json',
        //     },
        //     body: JSON.stringify({
        //         chatID: Number(chatId),
        //         afterID: Number(lastMesID),
        //     }),
        // });
        const msgs = await res.json();
        if (Array.isArray(msgs) && msgs.length > 0) {
            msgs.forEach(m => {
                appendMessage(m);
            })
            lastMesID = msgs[msgs.length - 1].id;
        }
    } catch (e) {
        console.error('poll error', e);
    }
}

const messageForm = document.getElementById('msgSendForm');
if (messageForm) {
    messageForm.addEventListener('submit', async (event) => {
        event.preventDefault();
        const chatId = document.getElementById('chatIDField').value;
        const login = document.getElementById('login').value;
        const mes = document.getElementById('message').value;

        if (chatId == 0 && login == "") {
            alert("Choose chat before");
            return
        }
        if (!mes.trim()) {
            return;
        }
        document.getElementById('message').value = '';
        document.getElementById('login').value = '';
        try {
            const res = await fetch('/api/messages', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    chat_id: Number(chatId),
                    getter_id: login,
                    body: mes,
                }),
            });
            const msg = await res.json();
            if (login != "") {
                document.getElementById('chatIDField').value = 0;
                loadInitMes();
            }
            else {
                appendMessage(msg);
                lastMesID = msg.id;
                console.log(msg.id);
            }
        }
        catch (err) {
            console.error('send error', err);
        }
    })
}

loadInitMes();
setInterval(pollNewMessages, 1000);