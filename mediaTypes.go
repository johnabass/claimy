package main

import (
	"mime"
	"net/http"
)

type MediaType string

const (
	JWKMediaType    MediaType = "application/jwk+json"
	JWKSetMediaType MediaType = "application/jwk-set+json"
	PEMMediaType    MediaType = "application/x-pem-file"
	JWTMediaType    MediaType = "application/jwt"
)

func (mt MediaType) SetContentType(h http.Header) {
	h.Set("Content-Type", string(mt))
}

func ParseMediaType(mediaType string) (MediaType, map[string]string, error) {
	mt, params, err := mime.ParseMediaType(mediaType)
	return MediaType(mt), params, err
}
