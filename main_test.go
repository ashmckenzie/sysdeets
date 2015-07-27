package main

import "testing"

func TestFailure(t *testing.T) {
  if 1 != 1.5 {
    t.Error("Failure!")
  }
}

func TestSuccess(t *testing.T) {
  if 1 != 1 {
    t.Error("Failure!")
  }
}
