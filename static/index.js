var ws = null;
var wsUrl = null;
var userName = null;
var msgTpl = null;
var userTpl = null;

$(function() {
    if (!checkSupportHtml5()) {
        alert("您的浏览器不支持HTML5");
        window.close();
        return;
    }

    wsUrl = "ws://" + window.location.hostname + ":" + $("#wsPort").val() + "/ws";

    initUIAndEvent();

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
    msgTpl = Handlebars.compile($("#msgTpl").html(), {"noEscape": true});
    userTpl = Handlebars.compile($("#userTpl").html());

    // websocket
    ws = new WebSocket(wsUrl);
    ws.onopen = wsOnOpen;
    ws.onclose = wsOnClose;
    ws.onmessage = wsOnMessage;
    ws.onerror = wsOnError;
});

function initUIAndEvent() {
    $("#editor").wysiwyg();
    $("#editor").focus();

    $("#pictureBtn").click(function() {
        $("#pictureFile").click();
    });

    $("#submitBtn").click(function() {
        var content = $("#editor").html();
        if (content == "") {
            return;
        }
        wsSendMessage("message", content);
        $("#editor").html("");
        $("#editor").focus();
    });

    // toggle emotion panel 
    $("#emotionBtn").click(function() {
        var emtionBasePanel = $("#emotionBasePanel");
        var msgBasePanelHeight = parseInt($("#msgBasePanel ").css("bottom"), 10);
        if (emtionBasePanel.hasClass("hidden")) {
            emtionBasePanel.removeClass("hidden");
            $("#msgBasePanel ").css("bottom", (msgBasePanelHeight + 71) + 'px');
            return;
        }
        emtionBasePanel.addClass("hidden");
        $("#msgBasePanel ").css("bottom", (msgBasePanelHeight - 71) + 'px');
    });

    // click emotion
    $(".emotionBlock").click(function() {
        var index = $(this).attr("data-index");
        var img = $("<img />");
        img.attr("src", "/static/emotion/"+index+".png");
        $("#editor").append(img);
        $("#editor").focus();
    });

    $("#editor").bind('keyup', 'ctrl+return', function() {
        $("#submitBtn").click();
    });
}

function wsOnMessage(e) {
    console.log(e.data);return;

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
    wsSendMessage("auth", "", {
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

function wsSendMessage(type, body, data) {
    var values = $.extend(data, {"type": type})
    var querys = $.param(values);
    var content = "\n" + querys + "\n" + body;
    var sends = sprintf("%08d", content.length) + content;

    ws.send(sends);
}

function scrollToButtom(dom) {
    if(dom[0].scrollHeight - dom.scrollTop() <= dom.outerHeight() * 3) {
        dom.scrollTop(dom[0].scrollHeight);
    }
}

function checkSupportHtml5() {
  return !!document.createElement('canvas').getContext;
}

// 废弃
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
