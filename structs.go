package main

type Meta struct {
	Text  string `xml:",chardata"`
	Track struct {
		Text          string `xml:",chardata"`
		BjID          string `xml:"bj_id"`
		Nickname      string `xml:"nickname"`
		CopyrightID   string `xml:"copyright_id"`
		CopyrightNick string `xml:"copyright_nick"`
		Flag          string `xml:"flag"`
		ReadCnt       string `xml:"read_cnt"`
		Logo          string `xml:"logo"`
		Title         string `xml:"title"`
		Duration      string `xml:"duration"`
		RegDate       string `xml:"reg_date"`
		TitleImage    string `xml:"titleImage"`
		Video         struct {
			Text         string `xml:",chardata"`
			Chat         string `xml:"chat,attr"`
			ChatDuration string `xml:"chat_duration,attr"`
			Thumbnail    string `xml:"thumbnail,attr"`
			Log          string `xml:"log"`
			LevelsLabel  struct {
				Text string `xml:",chardata"`
				Info []struct {
					Text            string `xml:",chardata"`
					Level           string `xml:"level"`
					Label           string `xml:"label"`
					LabelResolution string `xml:"label_resolution"`
					Name            string `xml:"name"`
					Resolution      string `xml:"resolution"`
					Bitrate         string `xml:"bitrate"`
				} `xml:"info"`
			} `xml:"levels_label"`
			File []struct {
				Text        string `xml:",chardata"`
				Duration    string `xml:"duration,attr"`
				Key         string `xml:"key,attr"`
				LevelsLabel struct {
					Text string `xml:",chardata"`
					Info []struct {
						Text            string `xml:",chardata"`
						Level           string `xml:"level"`
						Label           string `xml:"label"`
						LabelResolution string `xml:"label_resolution"`
						Name            string `xml:"name"`
						Resolution      string `xml:"resolution"`
						Bitrate         string `xml:"bitrate"`
					} `xml:"info"`
				} `xml:"levels_label"`
			} `xml:"file"`
		} `xml:"video"`
		ChatURL                 string `xml:"chat_url"`
		ThumbURL                string `xml:"thumb_url"`
		UccType                 string `xml:"ucc_type"`
		WaterMark               string `xml:"water_mark"`
		SubscriptionPersonalcon string `xml:"subscription_personalcon"`
		FileType                string `xml:"file_type"`
		RegionType              string `xml:"region_type"`
		VodSeekTime             string `xml:"vod_seek_time"`
		ClipType                string `xml:"clip_type"`
		IsSubscribe             string `xml:"is_subscribe"`
	} `xml:"track"`
	Share struct {
		Text   string `xml:",chardata"`
		Link   string `xml:"link"`
		Iframe string `xml:"iframe"`
	} `xml:"share"`
	Watchtime struct {
		Text     string `xml:",chardata"`
		Useable  string `xml:"useable"`
		LastTime string `xml:"last_time"`
	} `xml:"watchtime"`
	Ppv struct {
		Text    string `xml:",chardata"`
		IsPPV   string `xml:"isPPV"`
		PPVNote string `xml:"PPVNote"`
	} `xml:"ppv"`
	Editangle     string `xml:"editangle"`
	LoadingBanner struct {
		Text     string `xml:",chardata"`
		AdType   string `xml:"ad_type"`
		ImageURL string `xml:"image_url"`
	} `xml:"loading_banner"`
	MidrollPointExposure string `xml:"midroll_point_exposure"`
	PathKey              string `xml:"path_key"`
}
