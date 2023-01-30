package common

//ResponseBean 统一封装响应结构体
type ResponseBean struct {
	Status string      `json:"status"`
	Code   int         `json:"code"`
	Msg    string      `json:"msg"`
	Body   interface{} `json:"body"`
}

func GenSuccessData(data interface{}) *ResponseBean {
	return &ResponseBean{"ok", 200, "", data}
}

func GenSuccessMsg(msg string) *ResponseBean {
	return &ResponseBean{"ok", 200, msg, ""}
}

func GenFailedMsg(errMsg string) *ResponseBean {
	return &ResponseBean{"fail", 400, errMsg, ""}
}

type H map[string]interface{}
