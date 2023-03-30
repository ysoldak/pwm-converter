//go:build rp2040

//
// RC PWM (1000-2000us at 50Hz) to regular PWM duty cycle converter.
//
// Example usage:
// - control brightness of LEDs on RC model via transmitter/radio channel.
//
// Coded for and tested on Waveshare RP2040 Zero board, see https://www.waveshare.com/wiki/RP2040-Zero
//
// Usage:
// - 5v, GND and GP29 from receiver, GP29 being signal wire that brings RC PWM to the board;
// - GP28 as "+" and GP27 as "-" to LED.
//
// Author: Yurii Soldak <ysoldak@gmail.com>
// 2023

package main

import (
	"machine"
	"time"
)

const (
	input  = machine.GPIO29 // signal, from receiver
	output = machine.GPIO28 // positive, to LEDs
	auxgnd = machine.GPIO27 // negative, to LEDs
)

var (
	outpwm = machine.PWM6 // GPIO28 attached to this
	ts     = int64(0)
	ms     = int64(0)
)

func main() {

	// input (signal, from receiver)
	input.Configure(machine.PinConfig{Mode: machine.PinInputPulldown})

	err := input.SetInterrupt(machine.PinRising|machine.PinFalling, func(machine.Pin) {
		if input.Get() {
			ts = time.Now().UnixMicro()
		} else {
			ms = time.Now().UnixMicro() - ts
		}
	})
	if err != nil {
		println("could not configure pin interrupt:", err.Error())
	}

	// output (positive, to LEDs)
	err = outpwm.Configure(machine.PWMConfig{
		Period: 1000e3, // 1ms, do not make larger, will be visible on camera
	})
	if err != nil {
		println("failed to configure PWM")
		return
	}
	outch, err := outpwm.Channel(output)
	if err != nil {
		println("failed to configure channel")
		return
	}
	outpwm.Set(outch, 0)

	// aux ground (negative, to LEDs)
	auxgnd.Configure(machine.PinConfig{Mode: machine.PinOutput})
	auxgnd.Low()

	// main loop
	for {
		outpwm.Set(outch, uint32(float64(outpwm.Top())*percent(ms)))
		time.Sleep(20 * time.Millisecond)
	}

}

func percent(ms int64) float64 {
	if ms < 1000 {
		return 0
	}
	if ms > 2000 {
		return 1
	}
	return float64(ms-1000) / (2000 - 1000)
}
