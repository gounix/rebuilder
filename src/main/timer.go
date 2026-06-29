package main

import (
	"time"
	"rebuilder/environ"
	"rebuilder/logger"
)

var cycleStart time.Time

func startCycle() {
	cycleStart = time.Now()
}

func startOfCycle()  time.Time {
	now := time.Now()

	yy := now.Year()
	mm := now.Month()
	dd := now.Day()
	hh := environ.Env.BuildHourStart
	start := time.Date(yy, mm, dd, hh, 0, 0, 0, now.Location())

	logger.Info("main.startOfCycle", "now", now, "start", start)
	return start
}

func endOfCycle() time.Time {
	start := startOfCycle()
	end := start.Add(time.Duration(environ.Env.BuildHours) * time.Hour)

	logger.Info("main.endOfCycle", "end", end)
	return end
}

func waitForStart() {

	now := time.Now()
	start := startOfCycle()
	end := endOfCycle()
	if now.Before(start) {
		sleepTime := start.Sub(now)
		logger.Info("main.waitForStart", "sleeping", sleepTime)
		time.Sleep(sleepTime)
		return
	}

	if now.Before(end) {
		logger.Info("main.waitForStart in build window")
		return
	}

	// Wait for next window
	next := start.Add(time.Duration(24 * time.Hour))
	sleepTime := next.Sub(now)
	logger.Info("main.waitForStart wait for next", "sleeping", sleepTime, "next", next)
	time.Sleep(sleepTime)
	return
}


func waitForNext(seqNr int, numBuilds int) {
	// called at the start of the loop

	windowLength := endOfCycle().Sub(cycleStart)
	delta := int64(windowLength) / int64(numBuilds) // minutes per build
	offset := int64(seqNr) * delta // in minutes since start

	startTime := cycleStart.Add(time.Duration(offset)) // time at which the nth run should start
	sleepTime := startTime.Sub(time.Now())

	logger.Info("main.waitForNext", "seqNr", seqNr, "sleep", sleepTime)
	time.Sleep(sleepTime)
}

func waitForEnd() {
	// use the time left in the current window
	end := endOfCycle()
	now := time.Now()
	sleepTime := end.Sub(now)
	sleepTime = sleepTime + time.Minute // to make sure

	logger.Info("main.waitForEnd", "sleep", sleepTime)
	time.Sleep(sleepTime)
}

