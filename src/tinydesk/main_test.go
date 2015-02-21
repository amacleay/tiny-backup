package main

import (
	"fmt"
	"testing"
)

func TestConcertUrls(t *testing.T) {
	no_concert_input := "<a href='http://www.npr.org/series/tiny-desk-concerts'"
	if x := ConcertUrls(no_concert_input); !x.IsEmpty() {
		t.Error("Expected empty list")
	}

	concert1 := "http://www.npr.org/event/music/374643866/trey-anastasio-tiny-desk-concert"
	concert2 := "http://www.npr.org/2010/12/15/131963244/the-heligoats-tiny-desk-concert"

	single_concert_input := fmt.Sprintf("<html>><a href='%s?autoplay=true'></html>", concert1)
	single_concert_out := ConcertUrls(single_concert_input)
	if single_concert_out.IsEmpty() || single_concert_out[0] != concert1 {
		t.Errorf("Expected '%s', got '%s'", concert1, single_concert_out)
	}

	double_concert_input := fmt.Sprintf(
		"<html><p><a href='%s?autoplay=true'></p><a href='%s'></html>",
		concert1,
		concert2)
	double_concert_out := ConcertUrls(double_concert_input)
	if double_concert_out.IsEmpty() {
		t.Error("Double concert output should have two URLs, have none")
	}
	if (double_concert_out[0] != concert1 || double_concert_out[1] != concert2) &&
		(double_concert_out[1] != concert1 || double_concert_out[0] != concert2) {
		t.Errorf("Double concert output should have included '%s' and '%s', received '%v'",
			concert1,
			concert2,
			double_concert_out)
	}

}

func TestIsEmpty(t *testing.T) {
	var emptygroup ConcertUrlGroup
	if !emptygroup.IsEmpty() {
		t.Errorf("Initialized group %v should be empty", emptygroup)
	}

	nonemptygroup := ConcertUrlGroup([]string{"http://www.example.com"})
	if nonemptygroup.IsEmpty() {
		t.Errorf("Group with something in it %v should not be empty", nonemptygroup)
	}

	groupwithoutstuff := ConcertUrlGroup(make([]string, 10))
	if !groupwithoutstuff.IsEmpty() {
		t.Errorf("Group with capacity but without items %v should be empty", groupwithoutstuff)
	}
}
