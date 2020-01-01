package main

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"golang.org/x/net/html/charset"
)

var output = "output.txt"

type VideoData struct {
	XMLName xml.Name    `xml:"video-data"`
	Text    string      `xml:",chardata"`
	Title   string      `xml:"title"`
	Serials []XMLSerial `xml:"group"`
}

type XMLSerial struct {
	Text     string `xml:",chardata"`
	GUID     string `xml:"guid,attr"`
	Type     string `xml:"type,attr"`
	MetaInfo struct {
		Text  string `xml:",chardata"`
		Title []struct {
			Text string `xml:",chardata"`
			Type string `xml:"type,attr"`
		} `xml:"title"`
		Description struct {
			Text string `xml:",chardata"`
			Type string `xml:"type,attr"`
		} `xml:"description"`
		Restriction     string `xml:"restriction"`
		Category        string `xml:"category"`
		Year            string `xml:"year"`
		Location        string `xml:"location"`
		Available       string `xml:"available"`
		Featured        string `xml:"featured"`
		Priority        string `xml:"priority"`
		ImdbID          string `xml:"imdb_id"`
		ExternalAllowed string `xml:"external_allowed"`
		Credits         struct {
			Text   string `xml:",chardata"`
			Credit []struct {
				Text  string `xml:",chardata"`
				Role  string `xml:"role,attr"`
				Award []struct {
					Text string `xml:",chardata"`
					Type string `xml:"type,attr"`
					Year string `xml:"year,attr"`
				} `xml:"award"`
			} `xml:"credit"`
		} `xml:"credits"`
		KinopoiskID string `xml:"kinopoisk_id"`
		Quote       struct {
			Text   string `xml:",chardata"`
			Author string `xml:"author,attr"`
		} `xml:"quote"`
		Slogan struct {
			Text string `xml:",chardata"`
			Type string `xml:"type,attr"`
		} `xml:"slogan"`
		StudioRestrictions struct {
			Text            string `xml:",chardata"`
			EpisodesAllowed string `xml:"episodes_allowed"`
		} `xml:"studio_restrictions"`
	} `xml:"meta-info"`
	Seasons []XMLSeason `xml:"group"`
}

type XMLSeason struct {
	Text     string `xml:",chardata"`
	Number   string `xml:"number,attr"`
	Type     string `xml:"type,attr"`
	MetaInfo struct {
		Text  string `xml:",chardata"`
		Title struct {
			Text string `xml:",chardata"`
			Type string `xml:"type,attr"`
		} `xml:"title"`
		Available struct {
			Text  string `xml:",chardata"`
			Start string `xml:"start,attr"`
		} `xml:"available"`
		Year        string `xml:"year"`
		Description struct {
			Text string `xml:",chardata"`
			Type string `xml:"type,attr"`
		} `xml:"description"`
	} `xml:"meta-info"`
	Videos []XMLVideo `xml:"video"`
}

type XMLVideo struct {
	Text            string `xml:",chardata"`
	End             string `xml:"end,attr"`
	Endtitles       string `xml:"endtitles,attr"`
	Episodesinopsys string `xml:"episodesinopsys,attr"`
	GUID            string `xml:"guid,attr"`
	Multilang       string `xml:"multilang,attr"`
	Number          string `xml:"number,attr"`
	Src             string `xml:"src,attr"`
	Start           string `xml:"start,attr"`
	MetaInfo        struct {
		Text  string `xml:",chardata"`
		Title []struct {
			Text string `xml:",chardata"`
			Type string `xml:"type,attr"`
		} `xml:"title"`
		Available struct {
			Text  string `xml:",chardata"`
			Start string `xml:"start,attr"`
			End   string `xml:"end,attr"`
		} `xml:"available"`
		Duration string `xml:"duration"`
		Featured string `xml:"featured"`
	} `xml:"meta-info"`
	Logo struct {
		Text string `xml:",chardata"`
		Src  string `xml:"src,attr"`
	} `xml:"logo"`
	Subtitles struct {
		Text string `xml:",chardata"`
		Src  string `xml:"src,attr"`
	} `xml:"subtitles"`
}

type Series struct {
	TitleOriginal   string
	TitleTranslated string
	Year            int
	Restriction     string
	Seasons         []Season
}

type Season struct {
	Number int
	Year   int

	Episodes []Episode
}

type Episode struct {
	Number    int
	File      string
	Available string
}

func (s Series) String() string {
	output := fmt.Sprintf("%v (%v) [%v %v]\n", s.TitleTranslated, s.TitleOriginal, s.Year, s.Restriction)
	for _, s := range s.Seasons {
		output += fmt.Sprintf("\t%v\n", s)
	}

	return output
}

func (s Season) String() string {
	output := fmt.Sprintf("s%02d %v\n", s.Number, s.Year)
	for _, e := range s.Episodes {
		output += fmt.Sprintf("\ts%02de%02d\t%v\t%v\n", s.Number, e.Number, e.File, e.Available)
	}
	return output
}

func main() {
	f, err := os.Open("amedia_tv_series.xml")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	decoder := xml.NewDecoder(f)
	decoder.CharsetReader = charset.NewReaderLabel
	var vd VideoData
	err = decoder.Decode(&vd)
	if err != nil {
		panic(err)
	}

	if _, err := os.Stat(output); err == nil {
		err = os.Remove(output)
		if err != nil {
			panic(err)
		}
	}

	for _, xmlSerial := range vd.Serials {
		year, err := strconv.Atoi(xmlSerial.MetaInfo.Year)
		if err != nil {
			panic(err)
		}

		serial := Series{
			TitleOriginal:   strings.TrimSpace(xmlSerial.MetaInfo.Title[0].Text),
			TitleTranslated: strings.TrimSpace(xmlSerial.MetaInfo.Title[1].Text),
			Year:            year,
			Restriction:     strings.TrimSpace(xmlSerial.MetaInfo.Restriction),
		}

		for _, xmlSeason := range xmlSerial.Seasons {
			number, err := strconv.Atoi(xmlSeason.Number)
			if err != nil {
				panic(err)
			}
			year, err := strconv.Atoi(xmlSeason.MetaInfo.Year)
			if err != nil {
				panic(err)
			}

			season := Season{
				Number: number,
				Year:   year,
			}

			for _, xmlEpisode := range xmlSeason.Videos {
				number, err := strconv.Atoi(xmlEpisode.Number)
				if err != nil {
					panic(err)
				}

				episode := Episode{
					Number:    number,
					File:      strings.TrimSpace(filepath.Base(xmlEpisode.Src)),
					Available: strings.TrimSpace(xmlEpisode.MetaInfo.Available.Start),
				}

				season.Episodes = append(season.Episodes, episode)
			}

			serial.Seasons = append(serial.Seasons, season)
		}

		s := fmt.Sprintf("%v\n", serial)
		fmt.Print(s)
		writeStringToFile(output, s)
	}
}

func writeStringToFile(filename string, str string) error {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0775)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err = f.WriteString(str); err != nil {
		return err
	}

	return nil
}
