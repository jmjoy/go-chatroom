var ws = null;
var wsUrl = null;
var userName = null;
var msgTpl = null;
var userTpl = null;
var systemTpl = null;
var singleCC = null;
var waitGroup = 0;

var delims = {
    left:  "(#####)",
    right: "(%%%%%)",

    reg: "\\(#####\\).*?\\(%%%%%\\)"
};

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
    systemTpl = Handlebars.compile($("#systemTpl").html());

    // websocket
    ws = new WebSocket(wsUrl);
    ws.onopen = wsOnOpen;
    ws.onclose = wsOnClose;
    ws.onmessage = wsOnMessage;
    ws.onerror = wsOnError;
});

function initUIAndEvent() {
    $("#editor").wysiwyg({
        hotKeys: {
            'return': ''
        }
    });
    $("#editor").focus();

    $("#pictureBtn").click(function() {
        $("#pictureFile").click();
    });

    $("#submitBtn").click(function() {
        // check if there is unfinish request
        if (waitGroup > 0) {
            return;
        }

        // get content
        var content = $("#editor").html();
        if (content == "") {
            return;
        }

        // parse content
        var cc = $("<cc>" + content + "</cc>");
        cc.find("br").after("\n").remove();
        cc.find("div").each(function() {
            var html = $(this).html();
            $(this).after(html + "\n").remove();
        });

        cc.find("img[data-type=emotion]").each(function() {
            var index = $(this).attr("data-index");
            $(this).after(delims.left + "E" + index + delims.right).remove();
        });

        var imgDatas = [];
        // handle img upload
        cc.find("img").each(function() {
            var srcData = $(this).attr("src");
            if (srcData.indexOf("data:image") != 0) {
                return;
            }
            var index = srcData.indexOf(",");
            if (index == -1) {
                return;
            }
            //var base64Data = srcData.substr(index + 1);
            //imgDatas[waitGroup] = atob(base64Data);
            imgDatas[waitGroup] = srcData.substr(index + 1);

            //var blob = new Blob(imgDatas[waitGroup]);
            //console.log(blob.size);
            //throw '';

            $(this).attr("data-index", waitGroup);
            waitGroup++;
        });

        // no images upload
        if (waitGroup == 0) {
            sendMessage(cc.html());
            return;
        }

        singleCC = cc;
        for (var i = 0; i < imgDatas.length; i++) {
            wsSendMessage("image", imgDatas[i], {"index": i});
        }

        // use setInterval to simulate channel
        var waitTimer = setInterval(function(data) {
            if (waitGroup > 0) {
                return;
            }

            // has finished all image upload
            clearInterval(waitTimer);

            sendMessage(singleCC.html());

        }, 350);
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
        img.attr("src", "data:image/gif;base64,R0lGODlhAQABAIAAAAAAAP///yH5BAEAAAAALAAAAAABAAEAAAIBRAA7");
        img.attr("data-type", "emotion");
        img.attr("data-index", index);

        var offset = 25 * index;
        img.addClass("emotionBlock");
        img.css("background-position", "0 -" + offset + "px");

        $("#editor").append(img);
        $("#editor").focus();
    });

    $("#editor").bind('keyup', 'return', function() {
        $("#submitBtn").click();
    });
}

function sendMessage(content) {
    wsSendMessage("message", htmlspecialchars_decode(content), {});
    $("#editor").html("");
    $("#editor").focus();
}

function wsOnMessage(e) {
    var index = e.data.indexOf("\n");
    var queryString = e.data.substring(0, index);
    var obj = getQueryParameters(queryString);

    switch (obj.type) {
    case "join":
    case "leave":
        displaySystem(decodeURIComponent(obj.message), "warning");
        displayUsers($.parseJSON(e.data.substring(index+1)));
        break;

    case "error":
        displaySystem(decodeURIComponent(obj.message), "danger");
        break;

    case "message":
        var body = e.data.substring(index+1);
        displayMessage(decodeURIComponent(obj.userName), decodeURIComponent(obj.time), body);
        break;

    case "image":
        singleCC.find("img[data-index="+obj.index+"]").
            after(delims.left + "I" + decodeURIComponent(obj.pathid) + delims.right).
            remove();
        waitGroup--;
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

function displayMessage(userName, time, content) {
    content = htmlspecialchars(content);
    content = nl2br(content);

    // handle emotions and images
    content = content.replace(new RegExp(delims.reg, "g"), function(word) {
        var media = word.slice(delims.left.length, -delims.right.length);
        switch (media[0]) {
        case "E":
            var span = $("<span>");
            var index = media.substr(1);
            var offset = 25 * index;
            span.addClass("emotionBlock");
            span.css("background-position", "0 -" + offset + "px");
            return $("<div>").html(span).html();

        case "I":
            var img = $("<img>");
            var path = media.substr(1);
            img.attr("src", "/upload/" + path);
            img.css("max-width", "100%");
            return $("<div>").html(img).html();
        }
    });

    var data = {
        "user_name":  userName,
        "time":       time,
        "content":    content,
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

function displaySystem(msg, color) {
    var data = {
        "msg":   msg,
        "color": color,
    };
    var html = systemTpl(data);
    $("#systemPanel").append(html);
    scrollToButtom($("#systemPanel"));
}

function wsSendMessage(type, body, data) {
    var values = $.extend(data, {"type": type})
    var querys = $.param(values);
    var content = "\n" + querys + "\n" + body;
    var sends = sprintf("%08d", getByteLen(content)) + content;

    ws.send(sends);
}

function scrollToButtom(dom) {
    if(dom[0].scrollHeight - dom.scrollTop() <= dom.outerHeight() * 3) {
        dom.scrollTop(dom[0].scrollHeight);
    }

    if (dom.find(".alert").size() > 250) {
        dom.find(".alert")[0].remove();
    }
}

function checkSupportHtml5() {
  return !!document.createElement('canvas').getContext;
}

/**
 * Count bytes in a string's UTF-8 representation.
 * @param   string
 * @return  int
 */
function getByteLen(normal_val) {
    // Force string type
    normal_val = String(normal_val);

    var byteLen = 0;
    for (var i = 0; i < normal_val.length; i++) {
        var c = normal_val.charCodeAt(i);
        byteLen += c < (1 <<  7) ? 1 :
                   c < (1 << 11) ? 2 :
                   c < (1 << 16) ? 3 :
                   c < (1 << 21) ? 4 :
                   c < (1 << 26) ? 5 :
                   c < (1 << 31) ? 6 : Number.NaN;
    }
    return byteLen;
}

/**
 * Parse Query string to object
 * @param str
 * @return 
 */
function getQueryParameters(str) {
  return (str || document.location.search).replace(/(^\?)/,'').split("&").map(function(n){return n = n.split("="),this[n[0]] = n[1],this}.bind({}))[0];
}

function htmlspecialchars(str) {
    if (typeof(str) == "string") {
        str = str.replace(/&/g, "&amp;"); /* must do &amp; first */
        str = str.replace(/"/g, "&quot;");
        str = str.replace(/'/g, "&#039;");
        str = str.replace(/</g, "&lt;");
        str = str.replace(/>/g, "&gt;");
    }
    return str;
}

function htmlspecialchars_decode(str) {
    if (typeof(str) == "string") {
        str = str.replace(/&gt;/ig, ">");
        str = str.replace(/&lt;/ig, "<");
        str = str.replace(/&#039;/g, "'");
        str = str.replace(/&quot;/ig, '"');
        str = str.replace(/&amp;/ig, '&'); /* must do &amp; last */
    }
    return str;
}

function nl2br (str, is_xhtml) {
    var breakTag = (is_xhtml || typeof is_xhtml === 'undefined') ? '<br />' : '<br>';
    return (str + '').replace(/([^>\r\n]?)(\r\n|\n\r|\r|\n)/g, '$1' + breakTag + '$2');
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

