package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

var (
	username  = flag.String("username", "", "username")
	password  = flag.String("password", "", "password")
	baseUrl   = flag.String("url", "https://kindergarden.flexkids.nl", "url of the flexkids web site")
	outputDir = flag.String("o", "output", "output directory")
)

const (
	loginUrl         = "/login/login"
	photoAlbumUrl    = "/ouder/fotoalbum"
	standardAlbumUrl = "/ouder/fotoalbum/standaardalbum"
	mediaUrl         = "/ouder/media/download/media/"
)

type monthYear struct {
	year  int
	month int
}

type photo struct {
	monthYear
	photoId int
}

func main() {
	flag.Parse()
	if *username == "" || *password == "" {
		flag.Usage()
		return
	}

	loginCookie, err := login(*username, *password)
	if err != nil {
		log.Fatal("Could not login")
	}

	done := startDownloadPhotos(loginCookie, startGetAlbums(loginCookie, startGetMonths(loginCookie)))
	<-done
	log.Println("Downloaded all photos")

}

func startDownloadPhotos(loginCookie *http.Cookie, in chan *photo) chan interface{} {
	result := make(chan interface{})
	numWorkers := 2
	var wg sync.WaitGroup
	wg.Add(numWorkers)

	for i := 0; i < numWorkers; i++ {
		go func(n int) {
			defer wg.Done()
			downloadPhotos(loginCookie, in)
			log.Printf("Photo downloader %d is done\n", n)
		}(i + 1)
	}

	go func() {
		wg.Wait()
		close(result)
	}()
	return result
}

func startGetAlbums(loginCookie *http.Cookie, in chan *monthYear) chan *photo {
	channel := make(chan *photo)
	numWorkers := 12
	var wg sync.WaitGroup

	wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go func(n int) {
			defer wg.Done()
			getAlbums(loginCookie, in, channel)
			log.Printf("Album retriever %d is done\n", n)
		}(i + 1)
	}

	go func() {
		wg.Wait()
		close(channel)
	}()
	return channel
}

func startGetMonths(loginCookie *http.Cookie) chan *monthYear {
	channel := make(chan *monthYear)
	go func() {
		getMonths(loginCookie, channel)
		close(channel)
	}()
	return channel
}

func downloadPhotos(loginCookie *http.Cookie, in chan *photo) {
	for p := range in {
		req, err := http.NewRequest("GET", *baseUrl+mediaUrl+strconv.Itoa(p.photoId), nil)
		if err != nil {
			log.Println("Error creating request for photo")
			continue
		}
		req.AddCookie(loginCookie)
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Println("Error calling photo")
			continue
		}

		content, err := ioutil.ReadAll(res.Body)
		res.Body.Close()
		log.Printf("Downloading photo %d-%02d %d.jpg\n", p.year, p.month, p.photoId)
		err = ioutil.WriteFile(
			filepath.Join(
				*outputDir,
				fmt.Sprintf("%d-%02d", p.year, p.month),
				fmt.Sprintf("%d.jpg", p.photoId)),
			content, 0644)

		if err != nil {
			log.Println(err.Error())
			log.Printf("error while downloading  %+v", p)
		}
	}
}

func getAlbums(loginCookie *http.Cookie, in chan *monthYear, channel chan *photo) {
	photoIds := regexp.MustCompile("\"(\\d+)\"")
	for pair := range in {
		body := make(url.Values)
		body.Add("month", strconv.Itoa(pair.month))
		body.Add("year", strconv.Itoa(pair.year))

		req, err := http.NewRequest("POST", *baseUrl+standardAlbumUrl, strings.NewReader(body.Encode()))
		if err != nil {
			log.Println("Error creating request for albums")
			return
		}

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
		req.AddCookie(loginCookie)
		res, err := http.DefaultClient.Do(req)

		if err != nil {
			log.Printf("Error retrieving the album for %d-%d\n", pair.year, pair.month)
			continue
		}
		content, err := ioutil.ReadAll(res.Body)
		res.Body.Close()

		matches := photoIds.FindAllSubmatch(content, -1)
		for _, match := range matches {
			photoId, _ := strconv.Atoi(string(match[1]))
			photo := &photo{}
			photo.month = pair.month
			photo.year = pair.year
			photo.photoId = photoId
			channel <- photo
		}
	}

}

func getMonths(loginCookie *http.Cookie, channel chan *monthYear) {
	req, err := http.NewRequest("GET", *baseUrl+photoAlbumUrl, nil)
	if err != nil {
		log.Println("Error creating request for months")
		return
	}
	req.AddCookie(loginCookie)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("Error calling months")
		return
	}

	content, err := ioutil.ReadAll(res.Body)
	res.Body.Close()

	months := regexp.MustCompile("option data-month='(\\d+)' data-year='(\\d+)'")
	matches := months.FindAllSubmatch(content, -1)
	for _, match := range matches {
		month, _ := strconv.Atoi(string(match[1]))
		year, _ := strconv.Atoi(string(match[2]))

		filePath := filepath.Join(*outputDir, fmt.Sprintf("%d-%02d", year, month))
		if err := os.MkdirAll(filePath, os.ModePerm); err != nil {
			log.Printf("Could not create directory: %s\n", filePath)
		}
		channel <- &monthYear{month: month, year: year}
	}
}

func login(username, password string) (*http.Cookie, error) {
	data := make(url.Values)
	data.Add("username", username)
	data.Add("password", password)
	data.Add("role", "7")

	res, err := http.PostForm(*baseUrl+loginUrl, data)
	if err != nil {
		log.Println("Cannot login to flexkids")
		return nil, err
	}

	log.Println("Login successfull", res.Cookies())
	cookies := res.Cookies()
	return cookies[0], nil
}
