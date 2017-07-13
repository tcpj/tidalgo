package tidalgo

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

const API_URL string = "https://api.tidalhifi.com/v1/"

func PrepareEndpoint(baseUrl *url.URL, endpoint string) (string, error) {
	endpointUrl, err := url.Parse(endpoint)
	if err != nil {
		return "", err
	}

	resUrl := baseUrl.ResolveReference(endpointUrl)

	return resUrl.String(), nil
}

func NewSession(username string, password string, apiKey string) (*Session, error) {
	/*
		NOTE: since all FLACs are encrypted by using standard TIDAL api key,
		it's neccessary to use WiMP api key. This is temporary workaround,
		API will work with regular TIDAL key, EXCEPT providing proper
		unencrypted FLAC stream. Which is kinda crucial. BUT! You can
		still use regular TIDAL api key and downgrade sound quality to "HIGH"
		or bellow, unencrypted M4A stream will be provided.
	*/
	var session *Session

	loginEndpoint := "login/username"
	baseUrl, err := url.Parse(API_URL)
	if err != nil {
		return session, err
	}

	loginUrl, err := PrepareEndpoint(baseUrl, loginEndpoint)
	if err != nil {
		return session, err
	}

	loginValues := url.Values{}
	loginValues.Add("username", username)
	loginValues.Add("password", password)

	client := &http.Client{}

	loginRequest, err := http.NewRequest("POST", loginUrl, strings.NewReader(loginValues.Encode()))
	if err != nil {
		return session, err
	}

	loginRequest.Header.Add("X-Tidal-Token", apiKey)
	loginRequest.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	response, err := client.Do(loginRequest)
	if err != nil {
		return session, err
	}

	var loginResponse struct {
		UserId      int
		SessionId   string
		CountryCode string
	}

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return session, err
	}

	if err = json.Unmarshal(data, &loginResponse); err != nil {
		return session, err
	}

	session = &Session{
		Id:          loginResponse.SessionId,
		UserId:      loginResponse.UserId,
		CountryCode: loginResponse.CountryCode,
		BaseUrl:     baseUrl,
		Client:      client,
	}

	return session, nil
}

func (session *Session) ApiRequest(
	method, endpointUrl string, params map[string]string,
	body io.Reader, headers map[string]string) ([]byte, error) {

	requestUrl, err := PrepareEndpoint(session.BaseUrl, endpointUrl)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest(method, requestUrl, body)
	if err != nil {
		return nil, err
	}

	request.Header.Add("X-Tidal-SessionId", session.Id)

	for k, v := range headers {
		request.Header.Add(k, v)
	}

	query := request.URL.Query()
	query.Add("countryCode", session.CountryCode)

	for k, v := range params {
		query.Add(k, v)
	}

	request.URL.RawQuery = query.Encode()

	response, err := session.Client.Do(request)
	if err != nil {
		return nil, err
	}

	return ioutil.ReadAll(response.Body)
}

func (session *Session) MappedApiRequest(
	endpointUrl string, params map[string]string,
	headers map[string]string, structure interface{}) error {

	data, err := session.ApiRequest("GET", endpointUrl, params, nil, headers)
	if err != nil {
		return err
	}

	if err = json.Unmarshal(data, structure); err != nil {
		return err
	}
	return nil
}

func (session *Session) GetAlbum(albumId int) (Album, error) {
	var album Album
	albumEndpoint := fmt.Sprintf("albums/%d", albumId)

	if err := session.MappedApiRequest(albumEndpoint, nil, map[string]string{}, &album); err != nil {
		return album, err
	}

	return album, nil
}

func (session *Session) GetTrack(trackId int) (Track, error) {
	var track Track
	trackEndpoint := fmt.Sprintf("tracks/%d", trackId)

	if err := session.MappedApiRequest(trackEndpoint, nil, map[string]string{}, &track); err != nil {
		return track, err
	}

	return track, nil
}

func (session *Session) GetArtist(artistId int) (Artist, error) {
	var artist Artist
	artistEndpoint := fmt.Sprintf("artists/%d", artistId)

	if err := session.MappedApiRequest(artistEndpoint, nil, map[string]string{}, &artist); err != nil {
		return artist, err
	}

	return artist, nil
}

func (session *Session) GetStreamUrl(trackId int) (string, error) {
	var streamResponse struct {
		Codec         string   `json:"codec"`
		EncryptionKey string   `json:"encryptionKey"`
		Urls          []string `json:"urls"`
	}

	streamEndpoint := fmt.Sprintf("tracks/%d/urlpostpaywall", trackId)
	params := map[string]string{
		"audioquality":      "LOSSLESS",
		"urlusagemode":      "STREAM",
		"assetpresentation": "FULL",
	}

	if err := session.MappedApiRequest(
		streamEndpoint,
		params,
		map[string]string{"Origin": "https://listen.tidal.com"},
		&streamResponse); err != nil {

		return "", err
	}

	if streamResponse.EncryptionKey != "" {
		fmt.Printf(`WARNING: encrypted media in response.
			   But here is your encryption key: %s`,
			streamResponse.EncryptionKey)
	}
	return streamResponse.Urls[0], nil

}

func (session *Session) GetAlbumItems(albumId int) ([]Track, error) {
	items := make([]Track, 0)

	var itemsResponse struct {
		Items []struct {
			Item Track  `json:"item"`
			Type string `json:"type"`
		} `json:"items"`
		Limit              int `json:"limit"`
		Offset             int `json:"offset"`
		TotalNumberOfItems int `json:"totalNumberOfItems"`
	}

	itemsEndpoint := fmt.Sprintf("albums/%d/items", albumId)

	if err := session.MappedApiRequest(itemsEndpoint, nil, map[string]string{}, &itemsResponse); err != nil {
		return items, err
	}

	for _, item := range itemsResponse.Items {
		items = append(items, item.Item)
	}

	return items, nil
}
