import * as htmx from 'htmx.org';
import 'htmx.org';

// After posting to joinroom
htmx.on('htmx:afterRequest', (event: any) => {
    const response = JSON.parse(event.detail.xhr.response);
    if (response.type == "chat") {
        const input = htmx.find('#iroomid') as HTMLInputElement;
        if (input) {
            input.value = '';
        }
        if (response.error) {
            alert(response.error);
            return;
        }
        alert('Joined room');
        window.location.href = `/room?room=${response.room}&user=${response.id}`;
    } else if (response.type == "chess") {
        const input = htmx.find('#ichessid') as HTMLInputElement;
        if (input) {
            input.value = '';
        }
        if (response.error) {
            alert(response.error);
            return;
        }
        alert('Joined game');
        window.location.href = `/chess?room=${response.room}&user=${response.id}`;
    }
});