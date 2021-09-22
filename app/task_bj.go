package app

import service2 "task_client/service"

// 处理百家云数据
func ProcessBjy() {
	service := service2.BjyH{}
	var url = "https://e83301793.at.baijiayun.com/web/room/enter?room_id=21092260861493&user_number=373653&user_name=Peng&user_role=2&user_avatar=&status=1&sign=59c27cfb81fe7d0335a323d37d9ebccc"

	// 解析链接
	service.ParseP(url)

	// 设置连接 && 注册函数
	service.SetReqClient()

	// 发送一条ws信息 获取鉴权数据
	service.SetToken()
}
