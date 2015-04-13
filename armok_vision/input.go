package main

import "time"

var LastInput time.Time

func Input() {
	now := time.Now()
	if now.Sub(LastInput) < 10*time.Millisecond {
		return
	}
	LastInput = now

	x, y, z := findCenter()
	moved := false

	if IsKeyPressed(KeyLeft, true) {
		x--
		moved = true
	}

	if IsKeyPressed(KeyRight, true) {
		x++
		moved = true
	}

	if IsKeyPressed(KeyUp, true) {
		y--
		moved = true
	}

	if IsKeyPressed(KeyDown, true) {
		y++
		moved = true
	}

	if IsKeyPressed(KeyPageDown, true) {
		z--
		moved = true
	}

	if IsKeyPressed(KeyPageUp, true) {
		z++
		moved = true
	}

	viewLock.Lock()
	defer viewLock.Unlock()

	if IsKeyPressed(KeyF, false) {
		viewOverride = nil
	} else if moved {
		pos := [3]int32{x, y, z}
		viewOverride = &pos
	}
}
