package main

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/spot-nyc/spot"
)

func TestShortID(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"", ""},
		{"abc", "abc"},
		{"abcdefgh", "abcdefgh"},
		{"abcdefghi", "abcdefgh…"},
		{"19ffde12-cb36-4db9-9252-66aade4dbb9a", "19ffde12…"},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			assert.Equal(t, tc.want, shortID(tc.in))
		})
	}
}

func TestFormatTime(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"00:00:00", "12:00 AM"},
		{"06:30:00", "6:30 AM"},
		{"12:00:00", "12:00 PM"},
		{"13:45:00", "1:45 PM"},
		{"18:00:00", "6:00 PM"},
		{"23:59:00", "11:59 PM"},
		{"18:00", "6:00 PM"},
		{"not-a-time", "not-a-time"},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			assert.Equal(t, tc.want, formatTime(tc.in))
		})
	}
}

func TestJoinRestaurantNames(t *testing.T) {
	cases := []struct {
		name    string
		targets []spot.SearchTarget
		want    string
	}{
		{"empty", nil, "—"},
		{
			"single",
			[]spot.SearchTarget{{Restaurant: &spot.Restaurant{Name: "Gramercy Tavern"}}},
			"Gramercy Tavern",
		},
		{
			"multiple",
			[]spot.SearchTarget{
				{Restaurant: &spot.Restaurant{Name: "Gramercy Tavern"}},
				{Restaurant: &spot.Restaurant{Name: "Shuko"}},
				{Restaurant: &spot.Restaurant{Name: "4 Charles"}},
			},
			"Gramercy Tavern, Shuko, 4 Charles",
		},
		{
			"skips targets without a restaurant",
			[]spot.SearchTarget{
				{Restaurant: &spot.Restaurant{Name: "Gramercy Tavern"}},
				{Restaurant: nil},
				{Restaurant: &spot.Restaurant{Name: ""}},
			},
			"Gramercy Tavern",
		},
		{
			"all targets missing restaurant -> em dash",
			[]spot.SearchTarget{{Restaurant: nil}},
			"—",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, joinRestaurantNames(tc.targets))
		})
	}
}
