package bjy

type User struct {
	ActualName        string     `json:"actual_name"`
	AuditionDuration  int        `json:"audition_duration"`
	Avatar            string     `json:"avatar"`
	EndType           int        `json:"end_type"`
	ExtInfo           string     `json:"ext_info"`
	Group             int        `json:"group"`
	IsAudition        bool       `json:"is_audition"`
	IsBackdoor        int        `json:"is_backdoor"`
	IsRecord          int        `json:"is_record"`
	Name              string     `json:"name"`
	Number            string     `json:"number"`
	ReplaceUserNumber string     `json:"replace_user_number"`
	Status            int        `json:"status"`
	Type              int        `json:"type"`
	WebrtcInfo        WebrtcInfo `json:"webrtc_info"`
	WebrtcSupport     int        `json:"webrtc_support"`
	CameraCover       string     `json:"camera_cover"`
	SpecialGuest      int        `json:"special_guest"`
	IsFakeAssistant   int        `json:"is_fake_assistant"`
}

type WebrtcInfo struct {
	AppId        int       `json:"app_id"`
	FileToken    string    `json:"file_token"`
	FileToken2   string    `json:"file_token2"`
	ScreenToken  string    `json:"screen_token"`
	ScreenToken2 string    `json:"screen_token2"`
	Token        string    `json:"token"`
	Token2       string    `json:"token2"`
	WebrtcExt    WebrtcExt `json:"webrtc_ext"`
}

type WebrtcExt struct {
	AutoSwitch struct {
		Enable     int    `json:"enable"`
		Thresholds string `json:"thresholds"`
	} `json:"auto_switch"`
	Resolution            Resolution `json:"resolution"`
	VideoKeyframeInterval int        `json:"video_keyframe_interval"`
}

type Resolution struct {
	Mobile480x360 Equipment `json:"mobile480x360"`
	MobileFULLHD  Equipment `json:"mobile_f_u_l_l_h_d"`
	MobileHD      Equipment `json:"mobile_h_d"`
	MobileQVGA    Equipment `json:"mobile_q_v_g_a"`
	MobileVGA     Equipment `json:"mobile_v_g_a"`
}

type Equipment struct {
	AudioMaxBitrate int    `json:"audio_max_bitrate"`
	Height          int    `json:"height"`
	MaxBitrate      int    `json:"max_bitrate"`
	MaxFramerate    int    `json:"max_framerate"`
	MinBitrate      int    `json:"min_bitrate"`
	Name            string `json:"name"`
	Width           int    `json:"width"`
}

var (
	Mobile480x360 Equipment
	MobileFULLHD  Equipment
	MobileHD      Equipment
	MobileQVGA    Equipment
	MobileVGA     Equipment
	CResolution   Resolution
)

func init() {
	MobileFULLHD = Equipment{
		AudioMaxBitrate: 64,
		Height:          1080,
		MaxBitrate:      3000,
		MaxFramerate:    25,
		MinBitrate:      2000,
		Name:            "mobileFULLHD",
		Width:           1920,
	}
	MobileHD = Equipment{
		AudioMaxBitrate: 64,
		Height:          720,
		MaxBitrate:      1500,
		MaxFramerate:    25,
		MinBitrate:      1000,
		Name:            "mobileHD",
		Width:           1280,
	}
	MobileQVGA = Equipment{
		AudioMaxBitrate: 32,
		Height:          240,
		MaxBitrate:      200,
		MaxFramerate:    15,
		MinBitrate:      100,
		Name:            "mobileQVGA",
		Width:           320,
	}
	MobileVGA = Equipment{
		AudioMaxBitrate: 32,
		Height:          480,
		MaxBitrate:      300,
		MaxFramerate:    15,
		MinBitrate:      200,
		Name:            "mobileVGA",
		Width:           640,
	}
	Mobile480x360 = Equipment{
		AudioMaxBitrate: 32,
		Height:          360,
		MaxBitrate:      300,
		MaxFramerate:    15,
		MinBitrate:      200,
		Name:            "mobile480x360",
		Width:           480,
	}

	CResolution = Resolution{
		Mobile480x360: Mobile480x360,
		MobileFULLHD:  MobileFULLHD,
		MobileHD:      MobileHD,
		MobileQVGA:    MobileQVGA,
		MobileVGA:     MobileVGA,
	}
}
