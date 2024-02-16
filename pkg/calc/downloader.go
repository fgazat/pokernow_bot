package calc

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/gocarina/gocsv"
)

type TransationInfo struct {
	PlayerNickname string  `csv:"player_nickname"`
	BuyIn          float64 `csv:"buy_in"`
	BuyOut         float64 `csv:"buy_out"`
	Stack          float64 `csv:"stack"`
}
type Downloader interface {
	Download(url string) ([]TransationInfo, error)
}

type DownloadImpl struct{}

func (d DownloadImpl) Download(url string) ([]TransationInfo, error) {
	r := []TransationInfo{}
	urlParts := strings.Split(url, "/")
	gameID := urlParts[len(urlParts)-1]
	ledgerName := fmt.Sprintf("ledger_%s.csv", gameID)
	ledgerUrl := url + "/" + ledgerName
	out, err := os.Create(ledgerName)
	if err != nil {
		return r, err
	}
	defer out.Close()
	defer os.Remove(ledgerName)
	resp, err := http.Get(ledgerUrl)
	if err != nil {
		return r, err
	}
	defer resp.Body.Close()
	if _, err = io.Copy(out, resp.Body); err != nil {
		return r, err
	}
	in, err := os.Open(ledgerName)
	if err != nil {
		return r, err
	}
	defer in.Close()
	if err := gocsv.UnmarshalFile(in, &r); err != nil {
		return r, err
	}
	return r, nil
}
