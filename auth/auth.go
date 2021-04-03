package auth

import (
	"time"

	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
)

//Делаем авторизацию

func SpotifyAuth(clientID, clientSecret, refreshToken string) *spotify.Client {
	token := new(oauth2.Token)
	token.Expiry = time.Now().Add(time.Second * -1)
	token.RefreshToken = refreshToken
	authenticator := spotify.NewAuthenticator("no-redirect-url")
	authenticator.SetAuthInfo(clientID, clientSecret)
	client := authenticator.NewClient(token)
	return &client
}
