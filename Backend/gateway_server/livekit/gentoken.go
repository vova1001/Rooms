package livekit

import (
	"fmt"

	"github.com/livekit/protocol/auth"
)

type TokenService struct {
	url       string
	apiKey    string
	apiSecret string
}

type ConnectionData struct {
	URL   string `json:"url"`
	Token string `json:"token"`
}

func NewTokenService(url, apiKey, apiSecret string) *TokenService {
	return &TokenService{
		url:       url,
		apiKey:    apiKey,
		apiSecret: apiSecret,
	}
}

func (s *TokenService) CreateConnectionData(roomId, userId, userName string) (*ConnectionData, error) {
	accsessToken := auth.NewAccessToken(s.apiKey, s.apiSecret)
	accsessToken.SetIdentity(userId)
	accsessToken.SetName(userName)
	grant := &auth.VideoGrant{
		RoomJoin: true,
		Room:     roomId,
	}
	grant.SetCanPublish(true)
	grant.SetCanSubscribe(true)
	grant.SetCanPublishData(true)
	accsessToken.SetVideoGrant(grant)

	signedToken, err := accsessToken.ToJWT()
	if err != nil {
		return nil, fmt.Errorf("generate LiveKit token: %w", err)
	}

	return &ConnectionData{
		URL:   s.url,
		Token: signedToken,
	}, nil

}
