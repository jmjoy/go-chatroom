var ws = null;
var wsUrl = "ws://" + window.location.host + "/message";
var userName = null;
var msgTpl = null;
var userTpl = null;

$(function() {
    if (!checkSupportHtml5()) {
        alert("您的浏览器不支持HTML5");
    }

    // adjust ui height
    $(window).resize(resetPanelHeight);
    resetPanelHeight();

    // get userName
    userName = localStorage.getItem("user.name");
    if (!userName) {
        while (true) {
            var input = prompt("请输入您的名字");
            if (!input) {
                continue;
            }
            userName = input;
            localStorage.setItem("user.name", userName);
            break;
        }
    }

    // parse tpl
    msgTpl = Handlebars.compile($("#msgTpl").html());
    userTpl = Handlebars.compile($("#userTpl").html());

    // websocket
    ws = new WebSocket(wsUrl);
    ws.onopen = wsOnOpen;
    ws.onclose = wsOnClose;
    ws.onmessage = wsOnMessage;
    ws.onerror = wsOnError;
});

function submitMessage() {

}

function wsOnMessage(e) {
    console.log(e);
}

function wsOnOpen(e) {
    wsSendMessage("open", {
        "userName": userName
    });
}

function wsOnClose(e) {
}

function wsOnError(e) {
}

function displayMessage(user, content, color) {
}

function displayUsers(users) {
    
}

function wsSendMessage(type, data) {
    var json = {
        "type": type,
        "data": data
    };
    ws.send(JSON.stringify(json) + "\n");
}

function checkSupportHtml5() {
  return !!document.createElement('canvas').getContext;
}

function resetPanelHeight() {
    var winHeight = $(document).height();
    var userPanelHeight = winHeight - 110;
    var msgPanelHeight = winHeight - 120;

    if (userPanelHeight > 10) {
        $("#userPanel").css("height", userPanelHeight);
    }
    if (msgPanelHeight > 10) {
        $("#msgPanel").css("height", msgPanelHeight);
    }
}
