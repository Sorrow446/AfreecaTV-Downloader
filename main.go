package main

import (
	"bufio"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/grafov/m3u8"
)

var (
	jar, _ = cookiejar.New(nil)
	client = &http.Client{Jar: jar, Transport: &MyTransport{}}
)

const outFolder = "Afreecatv downloads"

type MyTransport struct{}

func (t *MyTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add(
		"User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 "+
			"(KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
	)
	req.Header.Add(
		"Referer", "https://www.afreecatv.com/",
	)
	return http.DefaultTransport.RoundTrip(req)
}

func getScriptDir() (string, error) {
	var (
		ok    bool
		err   error
		fname string
	)
	if filepath.IsAbs(os.Args[0]) {
		_, fname, _, ok = runtime.Caller(0)
		if !ok {
			return "", errors.New("Failed to get script filename.")
		}
	} else {
		fname, err = os.Executable()
		if err != nil {
			return "", err
		}
	}
	scriptDir := filepath.Dir(fname)
	return scriptDir, nil
}

func parseCookies() ([]*http.Cookie, error) {
	var cookies []*http.Cookie
	f, err := os.Open("cookies.txt")
	if err != nil {
		return nil, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") {
			continue
		}
		splitLine := strings.Split(line, "\t")
		secure, err := strconv.ParseBool(splitLine[3])
		if err != nil {
			return nil, err
		}
		cookie := &http.Cookie{
			Domain: splitLine[0],
			Name:   splitLine[5],
			Path:   splitLine[2],
			Secure: secure,
			Value:  splitLine[6],
		}
		cookies = append(cookies, cookie)
	}
	if scanner.Err() != nil {
		return nil, scanner.Err()
	}
	return cookies, nil
}

func setCookies(cookies []*http.Cookie) {
	urlObj, _ := url.Parse("https://www.afreecatv.com/")
	client.Jar.SetCookies(urlObj, cookies)
}

func readTxtFile(path string) ([]string, error) {
	var lines []string
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, strings.TrimSpace(scanner.Text()))
	}
	if scanner.Err() != nil {
		return nil, scanner.Err()
	}
	return lines, nil
}

func contains(lines []string, value string) bool {
	for _, line := range lines {
		if strings.EqualFold(line, value) {
			return true
		}
	}
	return false
}

func processUrls(urls []string) ([]string, error) {
	var (
		processed []string
		txtPaths  []string
	)
	for _, url := range urls {
		url = strings.Split(url, "?")[0]
		if strings.HasSuffix(url, ".txt") && !contains(txtPaths, url) {
			txtLines, err := readTxtFile(url)
			if err != nil {
				return nil, err
			}
			for _, txtLine := range txtLines {
				if !contains(processed, txtLine) {
					processed = append(processed, txtLine)
				}
			}
			txtPaths = append(txtPaths, url)
		} else {
			if !contains(processed, url) {
				processed = append(processed, url)
			}
		}
	}
	return processed, nil
}

func getParams(url string) (string, error) {
	req, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer req.Body.Close()
	if req.StatusCode != http.StatusOK {
		return "", errors.New(req.Status)
	}
	bodyBytes, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return "", err
	}
	bodyString := string(bodyBytes)
	regex := regexp.MustCompile(`document.VodParameter = '([\w=&]+)';`)
	match := regex.FindStringSubmatch(bodyString)
	if match == nil {
		return "", errors.New("No regex match.")
	}
	unix := time.Now().UnixNano() / int64(time.Millisecond)
	params := match[1] + fmt.Sprintf("&adultView=ADULT_VIEW&_=%d", unix)
	return params, nil
}

func getMeta(params string) (*Meta, error) {
	req, err := http.NewRequest(http.MethodGet,
		"https://stbbs.afreecatv.com/api/video/get_video_info.php", nil,
	)
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = params
	do, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer do.Body.Close()
	if do.StatusCode != http.StatusOK {
		return nil, nil
	}
	var obj Meta
	err = xml.NewDecoder(do.Body).Decode(&obj)
	if err != nil {
		return nil, err
	}
	if obj.Track.Flag != "SUCCEED" {
		return nil, errors.New("Bad response.")
	}
	return &obj, nil
}

func sanitize(filename string) string {
	regex := regexp.MustCompile(`[\/:*?"><|]`)
	sanitized := regex.ReplaceAllString(filename, "_")
	return sanitized
}

func fileExists(path string) (bool, error) {
	f, err := os.Stat(path)
	if err == nil {
		return !f.IsDir(), nil
	} else if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func checkUrl(url string) bool {
	regex := regexp.MustCompile(
		`^https://vod.afreecatv.com/(?:PLAYER/STATION|ST)/\d{8}$`,
	)
	match := regex.MatchString(url)
	if match {
		return true
	}
	return false
}

func extractMasterUrls(_meta *Meta) ([]string, error) {
	var masterUrls []string
	meta := _meta.Track.Video
	if meta.File == nil {
		masterUrls = append(masterUrls, meta.Text)
	} else {
		for _, file := range meta.File {
			masterUrls = append(masterUrls, file.Text)
		}
	}
	return masterUrls, nil
}

func extractSegmentUrls(masterUrls []string) ([]string, error) {
	var segmentUrls []string
	for num, masterUrls := range masterUrls {
		req, err := client.Get(masterUrls)
		if err != nil {
			return nil, err
		}
		if req.StatusCode != http.StatusOK {
			return nil, errors.New(req.Status)
		}
		defer req.Body.Close()
		playlist, listType, err := m3u8.DecodeFrom(req.Body, true)
		if err != nil {
			return nil, err
		}
		if listType == m3u8.MASTER {
			master := playlist.(*m3u8.MasterPlaylist)
			sort.Slice(master.Variants, func(x, y int) bool {
				return master.Variants[x].Bandwidth > master.Variants[y].Bandwidth
			})
			if num == 0 {
				fmt.Println(master.Variants[0].Resolution)
			}
			mediaManifestUrl := master.Variants[0].URI
			req, err = client.Get(mediaManifestUrl)
			if err != nil {
				return nil, err
			}
			if req.StatusCode != http.StatusOK {
				return nil, errors.New(req.Status)
			}
			defer req.Body.Close()
			playlist, _, err = m3u8.DecodeFrom(req.Body, true)
			if err != nil {
				return nil, err
			}
		} else {
			fmt.Println("Unknown resolution.")
		}
		media := playlist.(*m3u8.MediaPlaylist)
		for _, segment := range media.Segments {
			if segment == nil {
				break
			}
			segmentUrls = append(segmentUrls, segment.URI)
		}
	}
	return segmentUrls, nil
}

func download(segmentUrls []string, outPath string) error {
	var written int64
	f, err := os.OpenFile(outPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		fmt.Println("Failed to open output file for writing.")
		return err
	}
	defer f.Close()
	total := len(segmentUrls)
	for num, segmentUrl := range segmentUrls {
		num++
		percentage := int(float64(num) / float64(total) * float64(100))
		fmt.Printf("\rSegment %d of %d - %s, %d%%.", num, total, humanize.Bytes(uint64(written)), percentage)
		req, err := client.Get(segmentUrl)
		if err != nil {
			return err
		}
		defer req.Body.Close()
		if req.StatusCode != http.StatusOK {
			return errors.New(req.Status)
		}
		n, err := io.Copy(f, req.Body)
		if err != nil {
			return err
		}
		written += n
	}
	return nil
}

func callFfmpeg(preOutPath, outPath string) error {
	cmd := exec.Command("ffmpeg",
		"-loglevel", "error",
		"-i", preOutPath,
		"-c", "copy",
		outPath,
	)
	cmd.Stderr = os.Stdout
	err := cmd.Run()
	os.Remove(preOutPath)
	return err
}

func init() {
	fmt.Println(`																
 _____ ___                     _____ _____    ____                _           _
|  _  |  _|___ ___ ___ ___ ___|_   _|  |  |  |    \ ___ _ _ _ ___| |___ ___ _| |___ ___ 
|     |  _|  _| -_| -_|  _| .'| | | |  |  |  |  |  | . | | | |   | | . | .'| . | -_|  _|
|__|__|_| |_| |___|___|___|__,| |_|  \___/   |____/|___|_____|_|_|_|___|__,|___|___|_|			
	`)
	if len(os.Args) == 1 {
		fmt.Println("At least one URL or text file filename/path is required.")
		os.Exit(1)
	}
	scriptDir, err := getScriptDir()
	if err != nil {
		panic(err)
	}
	err = os.Chdir(scriptDir)
	if err != nil {
		panic(err)
	}
	err = os.Mkdir(outFolder, os.ModePerm)
	if err != nil && !errors.Is(err, os.ErrExist) {
		panic(err)
	}
	cookies, err := parseCookies()
	if err != nil {
		fmt.Println("Failed to parse cookies.")
		panic(err)
	}
	setCookies(cookies)
}

func main() {
	urls, err := processUrls(os.Args[1:])
	if err != nil {
		fmt.Println("Failed to process URLs.")
		panic(err)
	}
	total := len(urls)
	for num, url := range urls {
		fmt.Printf("URL %d of %d:\n", num+1, total)
		ok := checkUrl(url)
		if !ok {
			fmt.Println("Invalid URL:", url)
			continue
		}
		params, err := getParams(url)
		if err != nil {
			fmt.Println("Failed to parse params", err)
			continue
		}
		meta, err := getMeta(params)
		if err != nil {
			fmt.Println("Failed to get metadata.", err)
			continue
		}
		fname := meta.Track.Nickname + " - " + meta.Track.Title
		fmt.Println(fname)
		sanitizedFname := sanitize(fname)
		preOutPath := filepath.Join(outFolder, sanitizedFname+".ts")
		outPath := filepath.Join(outFolder, sanitizedFname+".mp4")
		exists, err := fileExists(outPath)
		if err != nil {
			fmt.Println("Failed to check if file already exists locally.", err)
			continue
		}
		if exists {
			fmt.Println("File already exists locally.")
			continue
		}
		masterUrls, err := extractMasterUrls(meta)
		if err != nil {
			fmt.Println("Failed to extract info.", err)
			continue
		}
		segmentUrls, err := extractSegmentUrls(masterUrls)
		if err != nil {
			fmt.Println("Failed to extract segment URLs.", err)
			continue
		}
		err = download(segmentUrls, preOutPath)
		if err != nil {
			fmt.Println("Failed to process segments.", err)
			continue
		}
		err = callFfmpeg(preOutPath, outPath)
		if err != nil {
			fmt.Println("Failed to put into MP4 container.", err)
		}
	}
}
