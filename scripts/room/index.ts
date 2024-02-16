import * as htmx from 'htmx.org';
import 'htmx.org';

const mhistory = htmx.find('#mhistory');
const textbar = htmx.find('#textbar') as HTMLInputElement;
htmx.on("htmx:wsConnecting", () => {
    window.location.href = '/';
});
htmx.on('htmx:wsOpen', () => {
    console.log("Connected!");
    alert("Room is {{room}}")
    mhistory.removeChild(mhistory.lastChild);
    textbar.hidden = false;
});
htmx.on('htmx:wsBeforeSend', () => {
    const li = document.createElement('li');
    const chat = htmx.find('#chatm') as HTMLInputElement;
//  Me: <message>
    li.innerHTML = `<div class="px-2 py-1 text-green-500">[me]: ${chat.value}</div>`;
    mhistory.append(li);
    chat.value = '';
});
htmx.on('htmx:wsClose', () => {
    window.location.href = '/';
});
htmx.on('htmx:wsError', () => {
    window.location.href = '/';
});
htmx.on('htmx:wsAfterMessage', (evt: any) => {
//    Insert into messages
    const msg = JSON.parse(evt.detail.message);
    var li = document.createElement('li');
//    <user>: <message>
    li.innerHTML = `<div class="px-2 py-1 text-green-500">[${msg.author}]: ${msg.content}</div>`;
    mhistory.append(li);
});