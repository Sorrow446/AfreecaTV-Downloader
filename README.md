# AfreecaTV-Downloader
AfreecaTV downloader written in Go.

![](https://i.imgur.com/Sq6DcR9.png)
[Windows binaries](https://github.com/Sorrow446/AfreecaTV-Downloader/releases)

# Setup
1. **Put FFmpeg in tool's directory.**
2. Login to [AfreecaTV](https://dereferer.me/?https://afreecatv.com/).
3. Install [EditThisCookie Chrome extension](https://chrome.google.com/webstore/detail/editthiscookie/fngmhnnpilhplaeedifhccceomclgfbg?hl=en) (any other Netscape extensions will also work).
4. Dump cookies to txt file named "cookies.txt" (https://afreecatv.com/ tab only).
5. Move cookies to tool's directory.

# Usage
The best available quality will be automatically selected. 

Download two videos:   
`atv_downloader_x64.exe https://vod.afreecatv.com/PLAYER/STATION/75107955 https://vod.afreecatv.com/PLAYER/STATION/75107951`    
Download a single video and from two text files:   
`atv_downloader_x64.exe https://vod.afreecatv.com/PLAYER/STATION/75107955 G:\1.txt G:\2.txt`

If building or running from source, you'll need to include the structs.   
`go run main.go structs.go <urls>...`
