package app

import "task_client/service/bjy"

// 处理百家云数据
func ProcessBjy() {
	s := &bjy.BjyH{}
	var url = "https://e83301793.at.baijiayun.com/web/room/enter?room_id=21092473707700&user_number=373653&user_name=Peng&user_role=2&user_avatar=&status=1&sign=fe7958566d74068aaf255ce4c2c9ba52"
	s.Serve(url)
}
