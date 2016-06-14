var socket;
 
$("#connect").click(function(event){
    socket = new WebSocket("ws://192.168.5.32:8000");
 
    socket.onopen = function(){
        alert("Socket has been opened");
    }
 
    socket.onmessage = function(msg){
        alert(msg.data);
    }
 
    socket.onclose = function() {
        alert("Socket has been closed");
    }
});
 
$("#send").click(function(event){
    socket.send("sending data to server!");
});
 
$("#close").click(function(event){
    socket.close();
})