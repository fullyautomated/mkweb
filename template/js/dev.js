const socket = new WebSocket('ws://' + location.host + '/ws');

socket.addEventListener('message', function (event) {
    var url = window.location.pathname;
    var filename = url.substring(url.lastIndexOf('/') + 1);

    console.log(filename)
    console.log(event.data)

    if ((filename.length == 0 && event.data === "index")
        || filename.replace(/\.[^/.]+$/, "") === event.data) {
        console.log("source file changed, reloading...")
        window.location.reload();
    }
});
