import * as htmx from 'htmx.org';
import Sortable, { Swap } from 'sortablejs';
import 'htmx.org';

Sortable.mount(new Swap());

const board = htmx.find('#board');
const o_name = htmx.find('#o-name');

const room = (htmx.find('#room-id') as HTMLTableCellElement).innerHTML;
const client = (htmx.find('#client-id') as HTMLTableCellElement).innerHTML;
const protoc = window.location.protocol === 'https:' ? 'wss' : 'ws';

const ws = new WebSocket(`${protoc}:${window.location.host}/chess/ws?room=${room}&user=${client}`);

ws.onopen = () => {
    console.log('Connection opened');
};

ws.onclose = () => {
    console.log('Connection closed');
    window.location.href = '/';
};

ws.onerror = (event) => {
    console.log('Error:', event);
    window.location.href = '/';
};

const swapElements = (obj1: HTMLElement, obj2: HTMLElement) => {
    // create marker element and insert it where obj1 is
    var temp = document.createElement("div");
    obj1.parentNode.insertBefore(temp, obj1);

    // move obj1 to right before obj2
    obj2.parentNode.insertBefore(obj1, obj2);

    // move obj2 to right before where obj1 used to be
    temp.parentNode.insertBefore(obj2, temp);

    // remove temporary marker node
    temp.parentNode.removeChild(temp);
}

ws.onmessage = (event) => {
    const data = JSON.parse(event.data);
    if (data.content.type === 'error') {
        console.log(data.content.msg);
        const src = data.content.src;
        const src_square = board.children[src];
        const src_piece = data.content.src_piece;
        const src_color = data.content.src_color;

        if (src_piece) {
            src_square.className = `bg-${src_color}`;
            src_square.innerHTML = `<input type="hidden" name="square" value="${src}"/><img src="https://upload.wikimedia.org/wikipedia/commons/${src_piece}" class="w-[5dvw] h-[5dvw]">`;
        } else {
            src_square.className = `unswappable w-[5dvw] h-[5dvw] bg-${src_color}`;
            src_square.innerHTML = `<input type="hidden" name="square" value="${src}" disabled/>`;
        }

        const trg = data.content.dst;
        const trg_square = board.children[trg];
        const trg_piece = data.content.dst_piece;
        const trg_color = data.content.dst_color;

        if (trg_piece) {
            trg_square.className = `bg-${trg_color}`;
            trg_square.innerHTML = `<input type="hidden" name="square" value="${trg}"/><img src="https://upload.wikimedia.org/wikipedia/commons/${trg_piece}" class="w-[5dvw] h-[5dvw]">`;
        } else {
            trg_square.className = `unswappable w-[5dvw] h-[5dvw] bg-${trg_color}`;
            trg_square.innerHTML = `<input type="hidden" name="square" value="${trg}" disabled/>`;
        }

    } else if (data.content.type === 'move'){
        // Swap the pieces
        const trg_pos = data.content.src;
        const src_pos = data.content.dst;
        const source = board.children[src_pos] as HTMLElement;
        const target = board.children[trg_pos] as HTMLElement;
        const src_end = source.classList.length - 1;
        const trg_end = target.classList.length - 1;
        const src_bg = source.classList[src_end];
        const trg_bg = target.classList[trg_end];
        source.classList.remove(src_bg);
        source.classList.add(trg_bg);
        target.classList.remove(trg_bg);
        target.classList.add(src_bg);
        swapElements(source, target);
    } else if (data.content.type === 'cmd') {
        if (data.content.msg === 'connected') {
            o_name.innerHTML = data.author;
            const ack = JSON.stringify({
                'type': 'cmd',
                'msg': 'acknowledge',
            });
            ws.send(ack);
        }
        if (data.content.msg === 'acknowledge') {
            o_name.innerHTML = data.author;
        }
    }
};

htmx.onLoad((ctt) => {
    const boards = ctt.querySelectorAll('#board');
    for (let i = 0; i < boards.length; i++) {
        const board = boards[i] as HTMLElement;
        const squareInstance = new Sortable(board, {
            animation: 150,
            swap: true,
            swapClass: 'bg-black',
            filter: '.unswappable',
            onEnd: (evt) => {
                const source = evt.item;
                const target = evt.swapItem;
                const src_pos = evt.oldIndex;
                const trg_pos = evt.newIndex;
                //get bg colors from both, should be in class
                const src_end = source.classList.length - 1;
                const trg_end = target.classList.length - 1;
                const src_bg = source.classList[src_end];
                const trg_bg = target.classList[trg_end];
                //swap bg colors
                source.classList.remove(src_bg);
                source.classList.add(trg_bg);
                target.classList.remove(trg_bg);
                target.classList.add(src_bg);
                const information = JSON.stringify({
                    'from': src_pos,
                    'to': trg_pos,
                    'type': 'move'
                });
                ws.send(information);
            }
        });
        board.addEventListener('htmx:afterSwap', (event) => {
            squareInstance.option("disabled", false);
        });
    }
});