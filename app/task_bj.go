package app

import (
	"task_client/service/bjy"
)

// 处理百家云数据
func ProcessBjy() {
	s := &bjy.BjyH{}
	var url = "https://e83301793.at.baijiayun.com/web/room/enter?room_id=21092394733076&user_number=373653&user_name=Peng&user_role=2&user_avatar=&status=1&sign=4a746b16fae18b9ab63b4d2f0f5b0fcb"

	// 解析链接
	s.ParseP(url)

	// 设置连接 && 注册函数
	s.SetReqClient()

	// 发送一条ws信息 获取鉴权数据
	s.MessageToken()

	select {}
}
