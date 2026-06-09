package main

import (
	"github.com/sirupsen/logrus"
	"net/http"
)

func main() {
	log := logrus.New()
	log.SetFormatter(&logrus.JSONFormatter{})

	resp, err := http.Get("https://api.github.com")
	if err != nil {
		log.WithError(err).Fatal("요청 실패")
	}
	defer resp.Body.Close()

	log.WithFields(logrus.Fields{
		"status": resp.StatusCode,
	}).Info("응답 수신")
}
