package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"text/template"
	"time"

	"github.com/Innsmouth-trip/go-spotify-github/auth"
	"github.com/joho/godotenv"
	"github.com/zmb3/spotify"
)

func init() {
	// грузим .env
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
}

// глобальные переменные
var trackName string
var artistName string
var imagedata string
var myuserUrl string
var client *spotify.Client

type TrackData struct {
	TrackName  string
	ArtistName string
	Image      string
	UserUrl    string
}

func main() {

	// получаем данные из env
	clientID := os.Getenv("SPOTIFY_ID")
	clientSecret := os.Getenv("SPOTIFY_SECRET")
	refreshToken := os.Getenv("REFRESH_TOKEN")
	// авторизовываемся, получаем клиент
	client = auth.SpotifyAuth(clientID, clientSecret, refreshToken)

	http.HandleFunc("/spotify", func(w http.ResponseWriter, r *http.Request) {

		// Получаем ссылку на юзера
		getName, _ := client.CurrentUser()
		for _, elem := range getName.ExternalURLs {
			myuserUrl = elem
		}

		// Делаем запрос играет ли что-либо в данный момент
		currentlyResult := GetDataPlayerCurrentlyPlaying()

		// если не играет, выполняем сценарий Recently
		if !currentlyResult.Playing {
			recentlyResult := GetDataPlayerRecentlyPlayed()

			// получаем id трека и запрашиваем информацию о нём
			trackId := recentlyResult.ID
			resentlyTrack, _ := client.GetTrack(trackId)

			// получаем имя артиста
			for _, elem := range resentlyTrack.Artists {
				artistName = elem.Name
			}

			// получаем имя трека
			trackName = resentlyTrack.SimpleTrack.Name

			// получаем изображение 300х300
			image := resentlyTrack.Album.Images[1]

			// Создаем буфер и пишем в него изображение
			var resentlyImage bytes.Buffer
			image.Download(&resentlyImage)

			// кодируем в base64 и делаем данные строкой
			newimage := Base64Encode(resentlyImage.Bytes())
			imagedata = string(newimage)

			//если играет то оставляем данные currentlyResult
		} else {
			for _, elem := range currentlyResult.Item.Artists {
				artistName = elem.Name
			}
			trackName = currentlyResult.Item.Name

			image := currentlyResult.Item.Album.Images[1]
			var currentImage bytes.Buffer
			image.Download(&currentImage)

			newimage := Base64Encode(currentImage.Bytes())
			imagedata = string(newimage)
		}

		// Создаем дату и сериализовываем её в стуктуру
		data := TrackData{
			TrackName:  trackName,
			ArtistName: artistName,
			Image:      imagedata,
			UserUrl:    myuserUrl,
		}

		// парсим html шаблон
		tmpl, err := template.ParseFiles("templates/index.html")
		if err != nil {
			fmt.Printf("template execution: %s", err)
		}
		// возвращаем заголовок страницы с типом свг, добавляем дату
		w.Header().Set("Content-Type", "image/svg+xml")
		if err := tmpl.Execute(w, data); err != nil {
			log.Printf("Error executing template: %v", err)
		}

	})
	// проверка на ошибки сервера
	err := http.ListenAndServe(":1984", nil)
	if err != nil {
		fmt.Println("ListenAndServe: ", err)
	}

}

// Получаем трек который играет
func GetDataPlayerCurrentlyPlaying() *spotify.CurrentlyPlaying {
	dataRes, err := client.PlayerCurrentlyPlaying()
	if err != nil {
		fmt.Printf("couldn't get features playlists: %v", err)
	}
	return dataRes
}

// Получаем список треков которые играли
func GetDataPlayerRecentlyPlayed() spotify.SimpleTrack {

	var track spotify.SimpleTrack

	dataRes, err := client.PlayerRecentlyPlayed()
	if err != nil {
		fmt.Printf("couldn't get features playlists: %v", err)
	}

	rand.Seed(time.Now().UnixNano())
	getRandomTrack := dataRes[rand.Intn(len(dataRes))]
	track = getRandomTrack.Track

	return track
}

// отдельная функция для энкодинга изображения в base64
func Base64Encode(src []byte) []byte {
	return []byte(base64.StdEncoding.EncodeToString(src))
}
