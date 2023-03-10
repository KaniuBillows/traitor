const ele = document.getElementById("editor");
const editor = CodeMirror.fromTextArea(ele, {
    lineNumbers: true, mode: "javascript", theme: "idea", lineWrapping: true, styleActiveLine: true, matchBrackets: true
});

editor.setSize('100%', '100%')


function getId() {
    let p = window.location.pathname
    return p.substring(6, p.length)
}

function saveScript() {
    let sc = editor.getValue()
    let id = getId()
    console.log(id)
    $.post(`/api/script?id=${id}`, JSON.stringify({
        "script": sc
    }), () => {

    })
}

const debug_out = []


function response() {
    let v = debug_out.join('\n')
    $('#debug-out').html(v)
}

function debugScript() {
    debug_out.length = 0
    let host = window.location.host
    let id = getId()

    const ws = new WebSocket(`ws://${host}/api/debug?id=${id}`)
    ws.addEventListener('message', e => {
        debug_out.push(e.data)
        response()
    })
}

