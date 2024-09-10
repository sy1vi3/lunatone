package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/sy1vi3/lunatone/config"
	"github.com/zsefvlol/timezonemapper"
)

type Coords struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

type DragoFortMode struct {
	Workers   int      `json:"workers"`
	Route     []Coords `json:"route"`
	FullRoute []Coords `json:"full_route"`
	PrioRaid  bool     `json:"prio_raid"`
	Showcase  bool     `json:"showcase"`
	Invasion  bool     `json:"invasion"`
}

type DragoQuestMode struct {
	Workers       int      `json:"workers"`
	Hours         []int    `json:"hours"`
	MaxLoginQueue int      `json:"max_login_queue"`
	Route         []Coords `json:"route"`
}

type DragoPokemonMode struct {
	Workers     int      `json:"workers"`
	Route       []Coords `json:"route"`
	EnableScout bool     `json:"enable_scout"`
	Invasion    bool     `json:"invasion"`
}

type DragoArea struct {
	Name         string           `json:"name"`
	Enabled      bool             `json:"enabled"`
	PokemonMode  DragoPokemonMode `json:"pokemon_mode"`
	QuestMode    DragoQuestMode   `json:"quest_mode"`
	FortMode     DragoFortMode    `json:"fort_mode"`
	Geofence     []Coords         `json:"geofence"`
	EnableQuests bool             `json:"enable_quests"`
	ID           int              `json:"id"`
}

type DragoPagination struct {
	HasNext     bool `json:"hasNext"`
	HasPrevious bool `json:"hasPrevious"`
	Total       int  `json:"total"`
}

type DragoAreasResponse struct {
	Data       []DragoArea     `json:"data"`
	Pagination DragoPagination `json:"pagination"`
}

type Area struct {
	Name     string
	Location Coords
	Timezone *time.Location
	ID       int
	Enabled  bool
	Workers  int
}

var httpClient http.Client

func getDragoAreas() ([]Area, error) {
	areas := []Area{}
	done := false
	page := 0

	for !done {
		req, _ := http.NewRequest("GET", config.Config.Settings.DragoURL+"/areas/?order=ASC&page="+fmt.Sprint(page)+"&perPage=100&sortBy=name", nil)
		req.Header.Set("Cookie", "authorized="+config.Config.Settings.DragoAuth)

		resp, err := httpClient.Do(req)
		if err != nil {
			log.Debug("Error sending create request:", err)
			return nil, err
		}
		defer resp.Body.Close()

		decodedAreas := new(DragoAreasResponse)
		json.NewDecoder(resp.Body).Decode(&decodedAreas)

		for _, area := range decodedAreas.Data {
			loc := area.Geofence[0]
			tz := timezonemapper.LatLngToTimezoneString(loc.Lat, loc.Lon)
			location, _ := time.LoadLocation(tz)
			areas = append(areas, Area{
				Location: area.Geofence[0],
				ID:       area.ID,
				Timezone: location,
				Name:     area.Name,
				Enabled:  area.Enabled,
				Workers:  area.PokemonMode.Workers,
			})
		}
		page++
		if !decodedAreas.Pagination.HasNext {
			done = true
		}

	}
	return areas, nil
}

func enableArea(area Area) {
	req, _ := http.NewRequest("GET", config.Config.Settings.DragoURL+"/areas/"+fmt.Sprint(area.ID)+"/enable", nil)
	req.Header.Set("Cookie", "authorized="+config.Config.Settings.DragoAuth)

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Debug("Error sending enable request:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Errorf("Error enabling %s: %d", area.Name, resp.StatusCode)
		return
	}

	log.Infof("Enabled %s", area.Name)
}

func disableArea(area Area) {
	req, _ := http.NewRequest("GET", config.Config.Settings.DragoURL+"/areas/"+fmt.Sprint(area.ID)+"/disable", nil)
	req.Header.Set("Cookie", "authorized="+config.Config.Settings.DragoAuth)

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Debug("Error sending disable request:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Errorf("Error disabling %s: %d", area.Name, resp.StatusCode)
		return
	}

	log.Infof("Disabled %s", area.Name)
}

func analyzeAreas(areas []Area) {
	timeNow := time.Now()
	for _, area := range areas {
		localTime := timeNow.In(area.Timezone)
		if slices.Contains(config.Config.Settings.ExcludeAreas, area.ID) {
			continue
		}
		if localTime.Hour() >= config.Config.Settings.DisableHour || localTime.Hour() < config.Config.Settings.EnableHour || area.Workers == 0 {
			if area.Enabled {
				disableArea(area)
			} else {
				log.Infof("%s already disabled (%d)", area.Name, localTime.Hour())
			}
		} else {
			if !area.Enabled {
				enableArea(area)
			} else {
				log.Infof("%s already enabled (%d)", area.Name, localTime.Hour())
			}
		}
	}
}

func main() {
	config.ReadConfig()
	httpClient = http.Client{
		Timeout: time.Second * 20,
	}

	areas, err := getDragoAreas()
	if err != nil {
		log.Error(err)
	}
	analyzeAreas(areas)
}
