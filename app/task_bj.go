package app

import "task_client/service/bjy"

// 处理百家云数据
func ProcessBjy() {
	s := &bjy.BjyH{}
	var url = "https://e83301793.at.baijiayun.com/web/room/enter?room_id=21092469494973&user_number=373653&user_name=Peng&user_role=2&user_avatar=&status=1&sign=a722c118169f13ef433f9fb9e4502deb"
	var x_class_id = "20028024"
	s.Serve(url, x_class_id)
}
