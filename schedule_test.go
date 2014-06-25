// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package main

import (
	"github.com/stretchr/testify/assert"
	"math/rand"
	"os"
	"testing"
	"time"
)

func TestNonExistentSchedule(t *testing.T) {
	teams := make([]Team, 6)
	scheduleBlocks := []ScheduleBlock{{time.Unix(0, 0).UTC(), 2, 60}}
	_, err := BuildRandomSchedule(teams, scheduleBlocks, "test")
	expectedErr := "No schedule template exists for 6 teams and 2 matches"
	if assert.NotNil(t, err) {
		assert.Equal(t, expectedErr, err.Error())
	}
}

func TestMalformedSchedule(t *testing.T) {
	scheduleFile, _ := os.Create("schedules/6_1.csv")
	defer os.Remove("schedules/6_1.csv")
	scheduleFile.WriteString("1,0,2,0,3,0,4,0,5,0,6,0\n6,0,5,0,4,0,3,0,2,0,1,0\n")
	scheduleFile.Close()
	teams := make([]Team, 6)
	scheduleBlocks := []ScheduleBlock{{time.Unix(0, 0).UTC(), 1, 60}}
	_, err := BuildRandomSchedule(teams, scheduleBlocks, "test")
	expectedErr := "Schedule file contains 2 matches, expected 1"
	if assert.NotNil(t, err) {
		assert.Equal(t, expectedErr, err.Error())
	}

	os.Remove("schedules/6_1.csv")
	scheduleFile, _ = os.Create("schedules/6_1.csv")
	scheduleFile.WriteString("1,0,asdf,0,3,0,4,0,5,0,6,0\n")
	scheduleFile.Close()
	_, err = BuildRandomSchedule(teams, scheduleBlocks, "test")
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "strconv.ParseInt")
	}
}

func TestScheduleTeams(t *testing.T) {
	rand.Seed(0)

	numTeams := 18
	teams := make([]Team, numTeams)
	for i := 0; i < numTeams; i++ {
		teams[i].Id = i + 101
	}
	scheduleBlocks := []ScheduleBlock{{time.Unix(0, 0).UTC(), 6, 60}}
	matches, err := BuildRandomSchedule(teams, scheduleBlocks, "test")
	assert.Nil(t, err)
	assert.Equal(t, Match{Type: "test", DisplayName: "1", Time: time.Unix(0, 0).UTC(), Red1: 115, Red2: 111,
		Red3: 108, Blue1: 109, Blue2: 116, Blue3: 117}, matches[0])
	assert.Equal(t, Match{Type: "test", DisplayName: "2", Time: time.Unix(60, 0).UTC(), Red1: 114, Red2: 112,
		Red3: 103, Blue1: 101, Blue2: 104, Blue3: 118}, matches[1])
	assert.Equal(t, Match{Type: "test", DisplayName: "3", Time: time.Unix(120, 0).UTC(), Red1: 110, Red2: 107,
		Red3: 105, Blue1: 106, Blue2: 113, Blue3: 102}, matches[2])
	assert.Equal(t, Match{Type: "test", DisplayName: "4", Time: time.Unix(180, 0).UTC(), Red1: 112, Red2: 108,
		Red3: 109, Blue1: 101, Blue2: 111, Blue3: 103}, matches[3])
	assert.Equal(t, Match{Type: "test", DisplayName: "5", Time: time.Unix(240, 0).UTC(), Red1: 113, Red2: 117,
		Red3: 115, Blue1: 110, Blue2: 114, Blue3: 102}, matches[4])
	assert.Equal(t, Match{Type: "test", DisplayName: "6", Time: time.Unix(300, 0).UTC(), Red1: 118, Red2: 105,
		Red3: 106, Blue1: 107, Blue2: 104, Blue3: 116}, matches[5])
}

func TestScheduleTiming(t *testing.T) {
	teams := make([]Team, 18)
	scheduleBlocks := []ScheduleBlock{{time.Unix(100, 0).UTC(), 10, 75},
		{time.Unix(20000, 0).UTC(), 5, 1000},
		{time.Unix(100000, 0).UTC(), 15, 29}}
	matches, err := BuildRandomSchedule(teams, scheduleBlocks, "test")
	assert.Nil(t, err)
	assert.Equal(t, time.Unix(100, 0).UTC(), matches[0].Time)
	assert.Equal(t, time.Unix(775, 0).UTC(), matches[9].Time)
	assert.Equal(t, time.Unix(20000, 0).UTC(), matches[10].Time)
	assert.Equal(t, time.Unix(24000, 0).UTC(), matches[14].Time)
	assert.Equal(t, time.Unix(100000, 0).UTC(), matches[15].Time)
	assert.Equal(t, time.Unix(100406, 0).UTC(), matches[29].Time)
}

func TestScheduleSurrogates(t *testing.T) {
	rand.Seed(0)

	numTeams := 38
	teams := make([]Team, numTeams)
	for i := 0; i < numTeams; i++ {
		teams[i].Id = i + 101
	}
	scheduleBlocks := []ScheduleBlock{{time.Unix(0, 0).UTC(), 64, 60}}
	matches, _ := BuildRandomSchedule(teams, scheduleBlocks, "test")
	for i, match := range matches {
		if i == 13 || i == 14 {
			if !match.Red1IsSurrogate || match.Red2IsSurrogate || match.Red3IsSurrogate ||
				!match.Blue1IsSurrogate || match.Blue2IsSurrogate || match.Blue3IsSurrogate {
				t.Errorf("Surrogates wrong for match %d", i+1)
			}
		} else {
			if match.Red1IsSurrogate || match.Red2IsSurrogate || match.Red3IsSurrogate ||
				match.Blue1IsSurrogate || match.Blue2IsSurrogate || match.Blue3IsSurrogate {
				t.Errorf("Expected match %d to be free of surrogates", i+1)
			}
		}
	}
}
