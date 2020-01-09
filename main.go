package main

import (
	"compress/gzip"
	"crypto/tls"
	"encoding/xml"
	"fmt"
	"flag"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"os"
	"strings"
	"time"

	"golang.org/x/net/publicsuffix"
	"github.com/evilsocket/islazy/log"
)

var murl string
var musername string
var mpassword string
var mdomain string
var mpath string

func init() {
	log.Output = "/dev/stdout"
	log.Level = log.INFO
	log.OnFatal = log.ExitOnFatal
	log.DateTimeFormat = "02-01-2006 15:04:05"
	log.Format = "{datetime} {level:name} {message}"

	flag.StringVar(&murl, "url", "", "WatchGuard IP address")
	flag.StringVar(&musername, "username", "admin", "WatchGuard admin username")
	flag.StringVar(&mpassword, "password", "", "WatchGuard admin password")
	flag.StringVar(&mdomain, "domain", "Firebox-DB", "WatchGuard domain")
	flag.StringVar(&mpath, "path", "", "path to backup")
	flag.Parse()

	if murl == "" ||  mpassword == "" {
		fmt.Println("Usage of watchGuard-backup-exporter:")
		flag.PrintDefaults()
		os.Exit(1)
	}
}

func Do() {

	options := cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	}
	jar, err := cookiejar.New(&options)
	if err != nil {
		log.Fatal("[SETUP] Cannot setup Cookie JAR: %s", err.Error())
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := http.Client{Timeout: 15 * time.Second, Jar: jar, Transport: tr}

	payload := "<methodCall>"
	payload += "<methodName>login</methodName>"
	payload += "<params>"
	payload += "	<param><value><struct>"
	payload += "	<member><name>user</name><value><string>" + musername +"</string></value></member>"
	payload += "	<member><name>password</name><value><string>" + mpassword +"</string></value></member>"
	payload += "	<member><name>domain</name><value><string>" + mdomain + "</string></value></member>"
	payload += "	<member><name>uitype</name><value><string>2</string></value></member>"
	payload += "	</struct></value></param>"
	payload += "</params>"
	payload += "</methodCall>"

	log.Info("Logging in as " + musername + " on " + murl)

	url := murl + "/agent/login"
	log.Debug("[LOGIN] POST " + url)
	log.Debug("[LOGIN] Payload: " + payload)
	res, err := client.Post(url, "text/xml", strings.NewReader(payload))
	if err != nil {
		log.Fatal("[LOGIN] HTTP error: %s", err.Error())
	}
	defer res.Body.Close()

	if res.StatusCode == 403 {
		log.Fatal("[LOGIN] Invalid username or password or domain")
	}

	byteValue, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal("[LOGIN] XML parse error: %s", err.Error())
	}
    var loginResponse LoginResponse
    xml.Unmarshal(byteValue, &loginResponse)

    sid := loginResponse.Params.Param.Value.Struct.Member[0].Value.Text
    csrf := loginResponse.Params.Param.Value.Struct.Member[1].Value.Text

    payload  = "<methodCall>"
	payload += "<methodName>/agent/file_action</methodName>"
	payload += "<params>"
	payload += "	<param><value><struct>"
	payload += "		<member><name>action</name><value><string>config</string></value></member>"
	payload += "	</struct></value></param>"
	payload += "	</params>"
	payload += "</methodCall>"

	log.Info("Generating configuration file")
	url = murl + "/agent/file_action"
	log.Debug("[FILE-ACTION] POST " + url)
	log.Debug("[FILE-ACTION] Payload: " + payload)
    req, err := http.NewRequest("POST", url, strings.NewReader(payload))
	if err != nil {
		log.Fatal("[FILE-ACTION] Failed to parse HTTP request: %s", err.Error())
	}
	req.Header.Set("Cookie", "sessionid=" + sid)
	req.Header.Set("X-CSRFToken", csrf)
	res, err = client.Do(req)
	if err != nil {
		log.Fatal("[FILE-ACTION] HTTP error: %s", err.Error())
	}
	defer res.Body.Close()

	if res.StatusCode == 404 {
		log.Fatal("[FILE-ACTION] Cannot generate file")
	}

	byteValue, err = ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal("[FILE-ACTION] XML parse error: %s", err.Error())
	}
    var fileActionresponse FileActionResponse
    xml.Unmarshal(byteValue, &fileActionresponse)
    log.Debug("Filename: " + fileActionresponse.Params.Param.Value.String)

    log.Info("Downloading configuration file")

    url = murl + "/agent/download?action=config"
	log.Debug("[DL] GET " + url)
	req, err = http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal("[DL] Failed to parse HTTP request: %s", err.Error())
	}
	req.Header.Set("Cookie", "sessionid=" + sid)
	req.Header.Set("X-CSRFToken", csrf)
	req.Header.Add("Accept-Encoding", "gzip")
	res, err = client.Do(req)
	if err != nil {
		log.Fatal("[DL] HTTP error: %s", err.Error())
	}
	defer res.Body.Close()

	gr, err := gzip.NewReader(res.Body)
	if err != nil {
		log.Fatal("[POST-DL] Cannot gunzip resonse: %s", err.Error())
	}
	defer gr.Close()

	content, err := ioutil.ReadAll(gr)
	if err != nil {
		log.Fatal("[POST-DL] Cannot decode resonse: %s", err.Error())
	}
	c := string(content)
	i := strings.Index(c, "<system-name>")
	j := strings.Index(c, "</system-name>")
	filename := strings.ReplaceAll(c[i:j], "<system-name>", "")
	now := time.Now()
	filename = filename + "-" + now.Format("200601021504") + ".xml"

	if mpath != "" {
		mpath = strings.TrimRight(mpath, "\\")
		mpath = strings.TrimRight(mpath, "/")
		filename = mpath + "/" + filename
	}

	f, err := os.Create(filename)
	if err != nil {
		log.Fatal("[POST-DL] Cannot create file: %s", err.Error())
	}
	defer f.Close()
	_, err = f.WriteString(c)
	if err != nil {
		log.Fatal("[POST-DL] Cannot write file: %s", err.Error())
	}

	log.Info("Downloaded configuration file: " + filename)

}

func main() {

	log.Info("---------------------------------------------------------")
	log.Info("|               WatchGuard Backup Exporter               |")
	log.Info("---------------------------------------------------------")

	Do()

}
