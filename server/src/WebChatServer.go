package server

import (
	"log"
	"net/http"

	"github.com/chanxuehong/wechat.v2/mp/core"
	"github.com/chanxuehong/wechat.v2/mp/menu"
	"github.com/chanxuehong/wechat.v2/mp/message/callback/request"
	"github.com/chanxuehong/wechat.v2/mp/message/callback/response"
	. "Minger/common"
)

const (
	wxAppId     = "wx8eb80f3a6e704f0f"
	wxAppSecret = "34dbb59e2adf994579b626ece660c016"

	wxOriId         = "gh_64f6c6a2ff72"
	wxToken         = "xiaodourobotserver"
	wxEncodedAESKey = "scfj2WrOuHUpjdc3yJghO6fzuNAJn8EUWS02qr7qlBm"
)

type WebChat struct {
	msgServer *core.Server
	messages  *MessageQueue
	name      string
	contexts map[string]*core.Context // record the context of the user, ugly to across webchat does not allow send client message using personal account
	//msgHandler	core.Handler
}

var (
// 下面两个变量不一定非要作为全局变量, 根据自己的场景来选择.
)

func NewWebChat() *WebChat {
	var webchat WebChat
	webchat.name = "gh_64f6c6a2ff72"
	webchat.messages = NewMessageQueue(100)
	webchat.contexts = make(map[string]*core.Context, 10)
	mux := core.NewServeMux()
	mux.DefaultMsgHandleFunc(webchat.defaultMsgHandler)
	mux.DefaultEventHandleFunc(webchat.defaultEventHandler)
	mux.MsgHandleFunc(request.MsgTypeText, webchat.textMsgHandler)
	mux.EventHandleFunc(menu.EventTypeClick, webchat.menuClickEventHandler)

	var msgHandler core.Handler = mux
	webchat.msgServer = core.NewServer(wxOriId, wxAppId, wxToken, wxEncodedAESKey, msgHandler, nil)

	return &webchat
}

func (w *WebChat) textMsgHandler(ctx *core.Context) {
	log.Printf("收到文本消息:\n%s\n", ctx.MsgPlaintext)

	msg := request.GetText(ctx.MixedMsg)
	w.contexts[msg.FromUserName] = ctx
	w.messages.Push(&Message{User: msg.FromUserName, CreateTime:msg.CreateTime, Format: "str", Content: []byte(msg.Content)})
	//resp := response.NewText(msg.FromUserName, msg.ToUserName, msg.CreateTime, msg.Content)
	//ctx.RawResponse(resp) // 明文回复
	//ctx.AESResponse(resp, 0, "", nil) // aes密文回复
}

func (w *WebChat) defaultMsgHandler(ctx *core.Context) {
	log.Printf("收到消息:\n%s\n", ctx.MsgPlaintext)
	ctx.NoneResponse()
}

func (w *WebChat) menuClickEventHandler(ctx *core.Context) {
	log.Printf("收到菜单 click 事件:\n%s\n", ctx.MsgPlaintext)

	event := menu.GetClickEvent(ctx.MixedMsg)
	resp := response.NewText(event.FromUserName, event.ToUserName, event.CreateTime, "收到 click 类型的事件")
	//ctx.RawResponse(resp) // 明文回复
	ctx.AESResponse(resp, 0, "", nil) // aes密文回复
}

func (w *WebChat) defaultEventHandler(ctx *core.Context) {
	log.Printf("收到事件:\n%s\n", ctx.MsgPlaintext)
	ctx.NoneResponse()
}

// wxCallbackHandler 是处理回调请求的 http handler.
//  1. 不同的 web 框架有不同的实现
//  2. 一般一个 handler 处理一个公众号的回调请求(当然也可以处理多个, 这里我只处理一个)
func (wc *WebChat) wxCallbackHandler(w http.ResponseWriter, r *http.Request) {
	wc.msgServer.ServeHTTP(w, r, nil)
}

func (wc *WebChat) Serve() {
	http.HandleFunc("/wx_callback", wc.wxCallbackHandler)
	log.Println(http.ListenAndServe(":4430", nil))
}

//func (wc *WebChat) SetMessageHandler(f func()) {
//	wc.messages.SetHandler(f)
//}

func (wc *WebChat) sendMessage(message *Message) {
	switch message.Format {
	case "str":
		resp := response.NewText(message.User, wc.name, message.CreateTime, string(message.Content))
		if ctx, ok := wc.contexts[message.User]; ok {
			ctx.RawResponse(resp)
		} else {
			log.Println("invalid user", message.User)
		}
	}
}
