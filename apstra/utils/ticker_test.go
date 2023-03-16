package utils

import (
	"log"
	"testing"
	"time"
)

func TestImmediateTickerFirstTick(t *testing.T) {
	interval := time.Second
	threshold := interval / 2

	start := time.Now()
	ticker := ImmediateTicker(time.Second)
	defer ticker.Stop()
	firstTick := <-ticker.C

	elapsed := firstTick.Sub(start)
	if elapsed > threshold {
		t.Fatalf("first tick after %q exceeds threshold %q", elapsed, threshold)
	}
	log.Printf("first tick after %q within threshold %q", elapsed, threshold)
}

func TestImmediateTickerSecondTick(t *testing.T) {
	interval := time.Second
	threshold1 := interval / 2
	threshold2 := interval/2 + interval

	start := time.Now()
	ticker := ImmediateTicker(time.Second)
	defer ticker.Stop()
	_ = <-ticker.C
	secondTick := <-ticker.C

	elapsed := secondTick.Sub(start)
	if elapsed < threshold1 {
		t.Fatalf("second tick after only %q doesn't meet threshold %q", elapsed, threshold1)
	}
	if elapsed > threshold2 {
		t.Fatalf("second tick after %q exceeds threshold %q", elapsed, threshold2)
	}
	log.Printf("second tick after %q within expected zone %q - %q", elapsed, threshold1, threshold2)
}
