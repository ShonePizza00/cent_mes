let chatsNum = 0;
const chatList = document.getElementById('chat-list');
chatList.addEventListener('click', function(event) {
    const link = event.target.closest('a[data-id]');
    if (!link) return;
    event.preventDefault();
    const num = Number(link.dataset.id);
    chatLoader(num);

    const prevActive = chatList.querySelector('.active');
    if (prevActive) {
        prevActive.classList.remove('active');
    }
    const li = link.closest('li');
    if (li) {
        li.classList.add('active')
    }
});

async function getChats() {
    const res = await fetch('/api/chats');
    const chats = await res.json();
    if (chats.length <= chatsNum) {
        return;
    }
    chatList.innerHTML = '<h2>Чаты</h2>';
    chats.forEach(c => {
        appendChat(c);
    });
    chatsNum = chats.length;
}

function appendChat(c) {
    const ul = document.createElement('ul');
    ul.innerHTML = `<li ><a href="#" data-id="${c.id}">${c.title}</a></li>`;
    chatList.appendChild(ul);
}

function chatLoader(chat_id) {
    const curChat = document.getElementById('chatIDField');
    curChat.value = chat_id;
    loadInitMes();
}
getChats();
setInterval(getChats, 1000);