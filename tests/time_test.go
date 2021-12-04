package tests_test

import (
	"testing"
	"time"
)

func TestTimes(t *testing.T) {
	stringTime := "2017-08-30 16:40:41"
	loc, _ := time.LoadLocation("Local")
	the_time, err := time.ParseInLocation("2006-01-02 15:04:05", stringTime, loc)

	if err == nil {
		unix_time := the_time.Unix() //1504082441
		t.Log(unix_time)
	}

}

func TestMap(t *testing.T) {

}