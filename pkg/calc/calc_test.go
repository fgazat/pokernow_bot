package calc

import (
	"fmt"
	"regexp"
	"testing"
)

// for gh search availability
const exampleURL = "https://www.pokernow.club/games/UniQueID"

func TestCalcuclate(t *testing.T) {
	dumb := fakeDownloader{}
	setFakeUsers()
	type args struct {
		url        string
		downloader Downloader
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "ok",
			args: args{
				url:        exampleURL,
				downloader: dumb,
			},
			want: `#table
Date: 2024-02-18

lucas -> scorsese 3100 руб на номер 789
lucas -> tarantino 1900 руб на номер 456
@lucas`,
			wantErr: false,
		},
	}
	for _, tt := range tests {

		r, err := regexp.Compile(`https://[^\s$]*`)
		if err != nil {
			t.Fatal(err)
		}
		result := r.FindAllStringSubmatch("https://www.pokernow.club/games/pgl7DnGH3M0Sq3Ny3a6mYIj9l", 1)[0][0]
		fmt.Println(result)
		t.Run(tt.name, func(t *testing.T) {
			got, err := Calcuclate(tt.args.url, "2024-02-18", "", tt.args.downloader)
			if (err != nil) != tt.wantErr {
				t.Errorf("Calcuclate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Calcuclate() = %v, want %v", got, tt.want)
			}
		})
	}
}

type fakeDownloader struct{}

func (d fakeDownloader) Download(url string) ([]TransationInfo, error) {
	if url == exampleURL {
		r := []TransationInfo{
			{PlayerNickname: "lucas", BuyIn: 1000, BuyOut: 0, Stack: 0},
			{PlayerNickname: "lucas", BuyIn: 2000, BuyOut: 0, Stack: 0},
			{PlayerNickname: "lucasfilm", BuyIn: 2000, BuyOut: 0, Stack: 0},
			{PlayerNickname: "scorsese", BuyIn: 1000, BuyOut: 4100, Stack: 0},
			{PlayerNickname: "tarantino", BuyIn: 1000, BuyOut: 0, Stack: 2900},
		}
		return r, nil
	}
	return []TransationInfo{}, nil
}

func setFakeUsers() {
	users = []UserFromTable{
		{
			Nicknames:   []string{"lucas", "lucasfilm"},
			Login:       "@lucas",
			PaymentInfo: "123",
		},
		{
			Nicknames:   []string{"tarantino", "mr.feetlover"},
			Login:       "@tarantino",
			PaymentInfo: "456",
		},
		{
			Nicknames:   []string{"scorsese"},
			Login:       "@scorsese",
			PaymentInfo: "789",
		},
	}
}
