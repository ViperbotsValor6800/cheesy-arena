// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web routes for generating practice and qualification schedules.

package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"time"
)

// Global vars to hold schedules that are in the process of being generated.
var cachedMatchType string
var cachedScheduleBlocks []ScheduleBlock
var cachedMatches []Match
var cachedTeamFirstMatches map[int]string

// Shows the schedule editing page.
func ScheduleGetHandler(w http.ResponseWriter, r *http.Request) {
	if len(cachedScheduleBlocks) == 0 {
		tomorrow := time.Now().AddDate(0, 0, 1)
		location, _ := time.LoadLocation("Local")
		startTime := time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 9, 0, 0, 0, location)
		cachedScheduleBlocks = append(cachedScheduleBlocks, ScheduleBlock{startTime, 10, 360})
		cachedMatchType = "practice"
	}
	renderSchedule(w, r, cachedMatchType, cachedScheduleBlocks, cachedMatches, cachedTeamFirstMatches, "")
}

// Generates the schedule and presents it for review without saving it to the database.
func ScheduleGeneratePostHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	scheduleBlocks, err := getScheduleBlocks(r)
	if err != nil {
		renderSchedule(w, r, cachedMatchType, cachedScheduleBlocks, cachedMatches, cachedTeamFirstMatches,
			"Incomplete or invalid schedule block parameters specified.")
		return
	}

	// Build the schedule.
	teams, err := db.GetAllTeams()
	if err != nil {
		handleWebErr(w, err)
		return
	}
	if len(teams) == 0 {
		renderSchedule(w, r, cachedMatchType, cachedScheduleBlocks, cachedMatches, cachedTeamFirstMatches,
			"No team list is configured. Set up the list of teams at the event before generating the schedule.")
		return
	}
	if len(teams) < 18 {
		renderSchedule(w, r, cachedMatchType, cachedScheduleBlocks, cachedMatches, cachedTeamFirstMatches,
			fmt.Sprintf("There are only %d teams. There must be at least 18 teams to generate a schedule.", len(teams)))
		return
	}
	matches, err := BuildRandomSchedule(teams, scheduleBlocks, r.PostFormValue("matchType"))
	if err != nil {
		renderSchedule(w, r, cachedMatchType, cachedScheduleBlocks, cachedMatches, cachedTeamFirstMatches,
			fmt.Sprintf("Error generating schedule: %s.", err.Error()))
		return
	}

	// Determine each team's first match.
	teamFirstMatches := make(map[int]string)
	for _, match := range matches {
		checkTeam := func(team int) {
			_, ok := teamFirstMatches[team]
			if !ok {
				teamFirstMatches[team] = match.DisplayName
			}
		}
		checkTeam(match.Red1)
		checkTeam(match.Red2)
		checkTeam(match.Red3)
		checkTeam(match.Blue1)
		checkTeam(match.Blue2)
		checkTeam(match.Blue3)
	}

	cachedMatchType = r.PostFormValue("matchType")
	cachedScheduleBlocks = scheduleBlocks
	cachedMatches = matches
	cachedTeamFirstMatches = teamFirstMatches
	http.Redirect(w, r, "/setup/schedule", 302)
}

// Saves the generated schedule to the database.
func ScheduleSavePostHandler(w http.ResponseWriter, r *http.Request) {
	existingMatches, err := db.GetMatchesByType(cachedMatchType)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	if len(existingMatches) > 0 {
		renderSchedule(w, r, cachedMatchType, cachedScheduleBlocks, cachedMatches, cachedTeamFirstMatches,
			fmt.Sprintf("Can't save schedule because a schedule of %d %s matches already exists. Clear it first "+
				" on the Settings page.", len(existingMatches), cachedMatchType))
		return
	}

	for _, match := range cachedMatches {
		err = db.CreateMatch(&match)
		if err != nil {
			handleWebErr(w, err)
			return
		}
	}
	http.Redirect(w, r, "/setup/schedule", 302)
}

func renderSchedule(w http.ResponseWriter, r *http.Request, matchType string, scheduleBlocks []ScheduleBlock,
	matches []Match, teamFirstMatches map[int]string, errorMessage string) {
	teams, err := db.GetAllTeams()
	if err != nil {
		handleWebErr(w, err)
		return
	}
	template, err := template.ParseFiles("templates/schedule.html", "templates/base.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	data := struct {
		*EventSettings
		MatchType        string
		ScheduleBlocks   []ScheduleBlock
		NumTeams         int
		Matches          []Match
		TeamFirstMatches map[int]string
		ErrorMessage     string
	}{eventSettings, matchType, scheduleBlocks, len(teams), matches, teamFirstMatches, errorMessage}
	err = template.ExecuteTemplate(w, "base", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// Converts the post form variables into a slice of schedule blocks.
func getScheduleBlocks(r *http.Request) ([]ScheduleBlock, error) {
	numScheduleBlocks, err := strconv.Atoi(r.PostFormValue("numScheduleBlocks"))
	if err != nil {
		return []ScheduleBlock{}, err
	}
	scheduleBlocks := make([]ScheduleBlock, numScheduleBlocks)
	location, _ := time.LoadLocation("Local")
	for i := 0; i < numScheduleBlocks; i++ {
		scheduleBlocks[i].StartTime, err = time.ParseInLocation("2006-01-02 03:04:05 PM",
			r.PostFormValue(fmt.Sprintf("startTime%d", i)), location)
		if err != nil {
			return []ScheduleBlock{}, err
		}
		scheduleBlocks[i].NumMatches, err = strconv.Atoi(r.PostFormValue(fmt.Sprintf("numMatches%d", i)))
		if err != nil {
			return []ScheduleBlock{}, err
		}
		scheduleBlocks[i].MatchSpacingSec, err = strconv.Atoi(r.PostFormValue(fmt.Sprintf("matchSpacingSec%d", i)))
		if err != nil {
			return []ScheduleBlock{}, err
		}
	}
	return scheduleBlocks, nil
}