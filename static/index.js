var ws = null;
var wsUrl = "ws://" + window.location.host + "/message";
var userName = null;
var msgTpl = null;
var userTpl = null;

$(function() {
    if (!checkSupportHtml5()) {
        alert("您的浏览器不支持HTML5");
        return;
    }

    // adjust ui height
    //$(window).resize(resetPanelHeight);
    //resetPanelHeight();

    $("#editor").wysiwyg();

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

    // event
    $("#msgForm").submit(submitMessage);

    // websocket
    ws = new WebSocket(wsUrl);
    ws.onopen = wsOnOpen;
    ws.onclose = wsOnClose;
    ws.onmessage = wsOnMessage;
    ws.onerror = wsOnError;
});

function submitMessage() {
    var content = $("#msgInput").val();
    $("#msgInput").val("");
    wsSendMessage("message", {"content": content});
    return false;
}

function wsOnMessage(e) {
    var data = $.parseJSON(e.data);
    switch (data.type) {
    case "auth":
    case "close":
        displayMessage("系统消息", data.message, "danger");
        displayUsers(data.userNames);
        break;
    case "message":
        displayMessage(data.userName, data.content, "default");
    }
}

function wsOnOpen(e) {
    wsSendMessage("auth", {
        "userName": userName
    });
}

function wsOnClose(e) {
    if (!confirm("聊天室连接已中断，是否重新加载页面？")) {
        return false;
    }
    location.reload();
}

function wsOnError(e) {
    alert("出现异常：", e);
}

function displayMessage(userName, content, color) {
    var data = {
        "user_name": userName,
        "content": content,
        "color": color,
    };
    var html = msgTpl(data);
    $("#msgPanel").append(html);
    scrollToButtom($("#msgPanel"));
}

function displayUsers(users) {
    var data = {"users": users};
    var html = userTpl(data);
    $("#userPanel").html(html);
    $("#numUser").html(users.length);
}

function wsSendMessage(type, data) {
    var json = {
        "type": type,
        "data": data
    };
    ws.send(JSON.stringify(json));
}

function scrollToButtom(dom) {
    if(dom[0].scrollHeight - dom.scrollTop() <= dom.outerHeight() * 3) {
        dom.scrollTop(dom[0].scrollHeight);
    }
}

function checkSupportHtml5() {
  return !!document.createElement('canvas').getContext;
}

function resetPanelHeight() {
    var winHeight = $(document).height();
    var userPanelHeight = winHeight - 110;
    var msgPanelHeight = winHeight - 123;

    if (userPanelHeight > 10) {
        $("#userPanel").css("height", userPanelHeight);
    }
    if (msgPanelHeight > 10) {
        $("#msgPanel").css("height", msgPanelHeight);
    }
}
